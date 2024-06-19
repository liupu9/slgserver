package net

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/llr104/slgserver/log"
	"go.uber.org/zap"
)

// http升级websocket协议的配置
var wsUpgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// server 服务器
type server struct {
	addr        string       // 地址
	router      *Router      // 路由
	needSecret  bool         // 是否密钥
	beforeClose func(WSConn) //
}

func NewServer(addr string, needSecret bool) *server {
	s := server{
		addr:       addr,
		needSecret: needSecret,
	}
	return &s
}

func (this *server) Router(router *Router) {
	this.router = router
}

func (this *server) Start() {
	log.DefaultLog.Info("server starting")
	http.HandleFunc("/", this.wsHandler)
	http.ListenAndServe(this.addr, nil)
}

func (this *server) SetOnBeforeClose(hookFunc func(WSConn)) {
	this.beforeClose = hookFunc
}

func (this *server) wsHandler(resp http.ResponseWriter, req *http.Request) {

	wsSocket, err := wsUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		return
	}

	conn := ConnMgr.NewConn(wsSocket, this.needSecret)
	log.DefaultLog.Info("client connect", zap.String("addr", wsSocket.RemoteAddr().String()))

	conn.SetRouter(this.router)
	conn.SetOnClose(ConnMgr.RemoveConn)
	conn.SetOnBeforeClose(this.beforeClose)
	conn.Start()
	conn.Handshake()

}
