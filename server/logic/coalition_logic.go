package logic

var Union *coalitionLogic

func init() {
	Union = &coalitionLogic{}
}

type coalitionLogic struct {

}

func (this* coalitionLogic) MemberEnter(rid, unionId int)  {
	RAttributeMgr.EnterUnion(rid, unionId)

	if ra, ok := RAttributeMgr.Get(rid); ok {
		ra.UnionId = unionId
	}

	if rbs, ok := RBMgr.GetRoleBuild(rid); ok {
		for _, rb := range rbs {
			rb.UnionId = unionId
		}
	}

	if rcs, ok := RCMgr.GetByRId(rid); ok {
		for _, rc := range rcs {
			rc.UnionId = unionId
			rc.SyncExecute()
		}
	}

	if armys, ok := AMgr.GetByRId(rid); ok {
		for _, army := range armys {
			army.UnionId = unionId
		}
	}
}

func (this* coalitionLogic) MemberExit(rid int) {

	if ra, ok := RAttributeMgr.Get(rid); ok {
		ra.UnionId = 0
	}

	if rbs, ok := RBMgr.GetRoleBuild(rid); ok {
		for _, rb := range rbs {
			rb.UnionId = 0
		}
	}

	if rcs, ok := RCMgr.GetByRId(rid); ok {
		for _, rc := range rcs {
			rc.UnionId = 0
			rc.SyncExecute()
		}
	}

	if armys, ok := AMgr.GetByRId(rid); ok {
		for _, army := range armys {
			army.UnionId = 0
		}
	}

}