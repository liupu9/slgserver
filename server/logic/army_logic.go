package logic

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/server"
	"slgserver/server/model_to_proto"
	"slgserver/server/proto"
	"slgserver/util"
	"time"
)


var ArmyLogic *armyLogic

func init() {
	ArmyLogic = &armyLogic{arriveArmys: make(chan *model.Army, 100),
		giveUpId: make(chan int, 100),
		updateArmys:  make(chan *model.Army, 100),
		posArmys: make(map[int]map[int]*model.Army),
		sys: NewSysArmy()}
	go ArmyLogic.running()
}

type armyLogic struct {
	sys         *sysArmyLogic
	giveUpId    chan int
	arriveArmys chan *model.Army
	updateArmys chan *model.Army
	posArmys    map[int]map[int]*model.Army	//玩家驻守的军队 key:posId,armyId
}

func (this *armyLogic) running(){
	for {
		select {
			case army := <-this.updateArmys:{
				army.DB.Sync()
				ap := &proto.ArmyStatePush{}
				ap.CityId = army.CityId
				model_to_proto.Army(army, &ap.Army)
				//通知部队变化了
				server.DefaultConnMgr.PushByRoleId(army.RId, "general.armyState", ap)

				if army.Cmd == model.ArmyCmdBack {
					posId := ToPosition(army.ToX, army.ToY)
					if _, ok := this.posArmys[posId]; ok {
						delete(this.posArmys[posId], army.Id)
					}
				}
			}
			case army := <-this.arriveArmys:{
				if army.Cmd == model.ArmyCmdAttack {
					if IsCanArrive(army.ToX, army.ToY, army.RId){
						this.battle(army)
					}
					AMgr.ArmyBack(army)
				}else if army.Cmd == model.ArmyCmdDefend {
					//呆在哪里不动
					posId := ToPosition(army.ToX, army.ToY)
					roleBuild, ok := RBMgr.PositionBuild(army.ToX, army.ToY)
					if ok && roleBuild.RId == army.RId {
						//目前是自己的领地才能驻守
						if _, ok := this.posArmys[posId]; ok == false {
							this.posArmys[posId] = make(map[int]*model.Army)
						}
						this.posArmys[posId][army.Id] = army
						army.State = model.ArmyStop
						this.Update(army)
					}else{
						AMgr.ArmyBack(army)
					}

				} else if army.Cmd == model.ArmyCmdBack {
					//如果该队伍在驻守，需要移除
					army.State = model.ArmyStop
					army.Cmd = model.ArmyCmdIdle
					this.Update(army)
				}
				army.DB.Sync()
			}
			case giveUpId := <- this.giveUpId:{
				armys, ok := this.posArmys[giveUpId]
				if ok {
					for _, army := range armys {
						AMgr.ArmyBack(army)
					}
					delete(this.posArmys, giveUpId)
				}
			}
		}
	}
}

func (this *armyLogic) Arrive(army* model.Army) {
	this.arriveArmys <- army
}

func (this *armyLogic) Update(army* model.Army) {
	this.updateArmys <- army
}

func (this *armyLogic) GiveUp(posId int) {
	this.giveUpId <- posId
}
//简单战斗
func (this *armyLogic) battle(army* model.Army) {
	_, ok := RCMgr.PositionCity(army.ToX, army.ToY)
	if ok {
		//打城池
		AMgr.ArmyBack(army)
		return
	}

	this.executeBuild(army)
}

func (this* armyLogic) executeBuild(army* model.Army)  {
	roleBuid, isRoleBuild := RBMgr.PositionBuild(army.ToX, army.ToY)

	posId := ToPosition(army.ToX, army.ToY)
	posArmys, isRoleEnemy := this.posArmys[posId]
	var enemys []*model.Army
	if isRoleEnemy == false {
		enemys = this.sys.GetArmy(army.ToX, army.ToY)
	}else{
		for _, v := range posArmys {
			enemys = append(enemys, v)
		}
	}

	warReports := make([]*model.WarReport, 0)
	general1 := make([]proto.General, 0)
	for _, id := range army.GeneralArray {
		g, ok := GMgr.FindGeneral(id)
		if ok {
			pg := proto.General{}
			model_to_proto.General(g, &pg)
			general1 = append(general1, pg)
		}
	}

	gdata1, _ := json.Marshal(general1)

	for _, enemy := range enemys {
		//战报处理
		general2 := make([]proto.General, 0)
		for _, id := range enemy.GeneralArray {
			g, ok := GMgr.FindGeneral(id)
			if ok {
				pg := proto.General{}
				model_to_proto.General(g, &pg)
				general2 = append(general2, pg)
			}
		}

		gdata2, _ := json.Marshal(general2)
		begArmy1, _ := json.Marshal(army)
		begArmy2, _ := json.Marshal(enemy)

		winCnt := 0
		for i, s1 := range army.SoldierArray {
			s2 := enemy.SoldierArray[i]
			enemy.SoldierArray[i] = util.MaxInt(0, enemy.SoldierArray[i]-s1)
			army.SoldierArray[i] = util.MaxInt(0, army.SoldierArray[i]-s2)
			if army.SoldierArray[i] > 0{
				winCnt+=1
			}
		}
		endArmy1, _ := json.Marshal(army)
		endArmy2, _ := json.Marshal(enemy)

		wr := &model.WarReport{X: army.ToX, Y: army.ToY, AttackRid: army.RId,
			AttackIsRead: false, DefenseIsRead: false, DefenseRid: enemy.RId,
			BegAttackArmy: string(begArmy1), BegDefenseArmy: string(begArmy2),
			EndAttackArmy: string(endArmy1), EndDefenseArmy: string(endArmy2),
			AttackIsWin: winCnt>=2, CTime: time.Now(), AttackGeneral: string(gdata1),
			DefenseGeneral: string(gdata2),
		}

		warReports = append(warReports, wr)
		enemy.ToSoldier()
		enemy.ToGeneral()

		if isRoleEnemy {
			if winCnt >= 2 {
				if isRoleEnemy {
					delete(this.posArmys, posId)
				}
				AMgr.ArmyBack(enemy)
			}
			enemy.DB.Sync()
		}

	}
	army.DB.Sync()

	//三盘两胜
	isWinCnt := 0
	for _, s := range army.SoldierArray {
		if s>0{
			isWinCnt+=1
		}
	}
	if isWinCnt>=2 {
		if isRoleBuild {
			destory := GMgr.GetDestroy(army)
			wr := warReports[len(warReports)-1]
			wr.DestroyDurable = util.MinInt(destory, roleBuid.CurDurable)
			roleBuid.CurDurable = util.MaxInt(0, roleBuid.CurDurable - destory)
			if roleBuid.CurDurable == 0{
				//攻占了玩家的领地
				wr.Occupy = 1
				RBMgr.RemoveFromRole(roleBuid)
				RBMgr.AddBuild(army.RId, army.ToX, army.ToY)
				roleBuid.CurDurable = roleBuid.MaxDurable
				this.OccupyRoleBuild(army.RId, army.ToX, army.ToY)
			}else{
				wr.Occupy = 0
			}
			roleBuid.DB.Sync()
		}else{
			//占领系统领地
			this.OccupySystemBuild(army.RId, army.ToX, army.ToY)
			wr := warReports[len(warReports)-1]
			wr.DestroyDurable = 100
			wr.Occupy = 1
			this.sys.DelArmy(army.ToX, army.ToY)
		}
	}

	//领地发生变化
	if newRoleBuild, ok := RBMgr.PositionBuild(army.ToX, army.ToY); ok {
		statePush := &proto.RoleBuildStatePush{}
		name := ""
		vRole, ok := RMgr.Get(newRoleBuild.RId)
		if ok {
			name = vRole.NickName
		}

		model_to_proto.MRBuild(newRoleBuild, &statePush.MRBuild, name)
		server.DefaultConnMgr.PushByRoleId(army.RId, "role.roleBuildState", statePush)
		if isRoleBuild {
			server.DefaultConnMgr.PushByRoleId(roleBuid.RId, "role.roleBuildState", statePush)
		}
	}


	push := &proto.WarReportPush{}
	push.List = make([]proto.WarReport, len(warReports))
	for i, wr := range warReports {
		fmt.Println(wr.CTime)
		_, err := db.MasterDB.InsertOne(wr)
		fmt.Println(wr.CTime)

		if err != nil{
			log.DefaultLog.Warn("db error", zap.Error(err))
		}else{
			model_to_proto.WarReport(wr, &push.List[i])
		}

		server.DefaultConnMgr.PushByRoleId(army.RId, "war.reportPush", push)
		if isRoleBuild {
			server.DefaultConnMgr.PushByRoleId(roleBuid.RId, "war.reportPush", push)
		}
	}

}

func (this* armyLogic) OccupyRoleBuild(rid, x, y int)  {
	newId := rid

	if b, ok := RBMgr.PositionBuild(x, y); ok {

		oldId := b.RId
		log.DefaultLog.Info("battle in role build",
			zap.Int("oldRId", oldId),
			zap.Int("newRId", newId))

		//被占领的减产
		if oldRole, ok := RResMgr.Get(oldId); ok{
			oldRole.WoodYield -= b.Wood
			oldRole.GrainYield -= b.Grain
			oldRole.StoneYield -= b.Stone
			oldRole.IronYield -= b.Iron
			oldRole.DB.Sync()
		}
		//占领的增加产量
		if newRole, ok := RResMgr.Get(newId); ok{
			newRole.WoodYield += b.Wood
			newRole.GrainYield += b.Grain
			newRole.StoneYield += b.Stone
			newRole.IronYield += b.Iron
			newRole.DB.Sync()
		}
		b.DB.Sync()
		b.RId = rid

	}
}

func (this* armyLogic) OccupySystemBuild(rid, x, y int)  {
	newId := rid

	if _, ok := RBMgr.PositionBuild(x, y); ok {
		return
	}

	if NMMgr.IsCanBuild(x, y){
		rb, ok := RBMgr.AddBuild(rid, x, y)
		if ok {
			//占领的增加产量
			if newRole, ok := RResMgr.Get(newId); ok{
				newRole.WoodYield += rb.Wood
				newRole.GrainYield += rb.Grain
				newRole.StoneYield += rb.Stone
				newRole.IronYield += rb.Iron
				newRole.DB.Sync()
			}
		}
	}
}