package ILogic

import "github.com/llr104/slgserver/server/slgserver/model"

// IArmyLogic 军队功能接口
type IArmyLogic interface {
	ArmyBack(army *model.Army)
	GetStopArmys(posId int) []*model.Army
	DeleteStopArmy(posId int)
	GetSysArmy(x, y int) []*model.Army
	DelSysArmy(x, y int)
}
