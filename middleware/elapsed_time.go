package middleware

import (
	"fmt"
	"time"

	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/net"
	"go.uber.org/zap"
)

// ElapsedTime 计算处理逻辑花费的时间（毫秒）
func ElapsedTime() net.MiddlewareFunc {
	return func(next net.HandlerFunc) net.HandlerFunc {
		return func(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
			bt := time.Now().UnixNano()
			next(req, rsp)
			et := time.Now().UnixNano()
			diff := (et - bt) / int64(time.Millisecond)

			log.DefaultLog.Info(">> ElapsedTime:",
				zap.String("msgName", req.Body.Name),
				zap.String("cost", fmt.Sprintf("%dms", diff)))
		}
	}
}
