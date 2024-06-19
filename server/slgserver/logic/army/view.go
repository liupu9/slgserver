package army

import (
	"github.com/llr104/slgserver/server/slgserver/global"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/logic/union"
	"github.com/llr104/slgserver/util"
)

var ViewWidth = 5
var ViewHeight = 5

// ArmyIsInView 判断坐标（x,y）是否在视野范围内
func ArmyIsInView(rid, x, y int) bool {
	unionId := union.GetUnionId(rid)
	xMin := util.MaxInt(x-ViewWidth, 0)
	xMax := util.MinInt(x+ViewWidth, global.MapWith)
	yMin := util.MaxInt(y-ViewHeight, 0)
	yMax := util.MinInt(y+ViewHeight, global.MapHeight)

	for i := xMin; i < xMax; i++ {
		for j := yMin; j < yMax; j++ {
			build, ok := mgr.RBMgr.PositionBuild(i, j)
			if ok {
				tUnionId := union.GetUnionId(build.RId)
				if (tUnionId != 0 && unionId == tUnionId) || build.RId == rid {
					return true
				}
			}

			city, ok := mgr.RCMgr.PositionCity(i, j)
			if ok {
				tUnionId := union.GetUnionId(city.RId)
				if (tUnionId != 0 && unionId == tUnionId) || city.RId == rid {
					return true
				}
			}
		}
	}

	return false
}
