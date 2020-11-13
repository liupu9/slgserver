package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"sync"
)

var nameArr = []string{	"校场", "募兵所", "疾风营", "铁壁营", "军机营", "尚武营",
						"统帅厅", "汉点将台", "魏点将台", "蜀点将台", "吴点将台", "群点将台",
						"兵营", "封禅台", "民居", "仓库", "伐木场", "炼铁场",
						"磨坊", "采石场", "城墙", "集市", "警戒所", "女墙",
						"烽火台", "守将府", "武神巨像", "沙盘阵图", "社稷坛",}

const (
	FacilityBEG		= iota
	FacilityJC      //校场
	FacilityMBS     //募兵所
	FacilityJFY     //疾风营
	FacilityTBY     //铁壁营
	FacilityJJY     //军机营
	FacilitySWY     //尚武营
	FacilityTST     //统帅厅
	FacilityHDJT    //汉点将台
	FacilityWEIDJT  //魏点将台
	FacilitySUDJT   //蜀点将台
	FacilityWUDJT   //吴点将台
	FacilityQUDJT   //群点将台
	FacilityBY		//兵营
	FacilityFCT		//封禅台
	FacilityMJ		//民居
	FacilityCK		//仓库
	FacilityFMC		//伐木场
	FacilityLTC		//炼铁场
	FacilityMF		//磨坊
	FacilityCSC		//采石场
	FacilityCQ		//城墙
	FacilityJS		//集市
	FacilityJJS		//警戒所
	FacilityNQ		//女墙
	FacilityFHT		//烽火台
	FacilitySJH		//守将府
	FacilityWSJX	//武神巨像
	FacilitySPZT	//沙盘阵图
	FacilitySJT		//社稷坛
	FacilityEND
)

type Facility struct {
	Name	string		`json:"name"`
	MLevel	int8		`json:"mLevel"`
	CLevel	int8		`json:"cLevel"`
	Type	int8		`json:"type"`
}

var RFMgr = FacilityMgr{
	facilities: make(map[int]*model.CityFacility),
}

type FacilityMgr struct {
	mutex sync.RWMutex
	facilities map[int]*model.CityFacility
}

func (this* FacilityMgr) Load() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	err := db.MasterDB.Find(this.facilities)
	if err != nil {
		log.DefaultLog.Error("FacilityMgr load city_facility table error")
	}
}

func (this* FacilityMgr) Get(cid int) (*model.CityFacility, error){
	this.mutex.RLock()
	r, ok := this.facilities[cid]
	if ok {
		return r, nil
	}
	this.mutex.RUnlock()

	r = &model.CityFacility{}
	ok, err := db.MasterDB.Table(r).Where("cityId=?", cid).Get(r)
	if ok {
		this.mutex.Lock()
		this.facilities[cid] = r
		this.mutex.Unlock()
		return r, nil
	}else{
		if err != nil{
			return nil, err
		}else{
			str := fmt.Sprintf("cid:%d CityFacility not found", cid)
			return nil, errors.New(str)
		}
	}
}

/*
如果不存在尝试去创建
*/
func (this* FacilityMgr) GetAndTryCreate(cid int) (*model.CityFacility, error){
	r, err := this.Get(cid)
	if err == nil {
		return r, nil
	}else{
		if _, err:= RCMgr.Get(cid); err == nil {
			//创建
			fs := make([]Facility, 30)
			for i := FacilityBEG; i < FacilityEND ; i++ {
				f := Facility{Type: int8(i+1), CLevel: int8(1), MLevel: int8(10), Name: nameArr[i]}
				fs[i] = f
			}

			sdata, _ := json.Marshal(fs)
			cf := &model.CityFacility{CityId: cid, Facilities: string(sdata)}
			db.MasterDB.Table(cf).Insert(cf)

			this.mutex.Lock()
			this.facilities[cid] = cf
			this.mutex.Unlock()

			return cf, nil
		}else{
			str := fmt.Sprintf("cid:%d not found", cid)
			return nil, errors.New(str)
		}
	}
}