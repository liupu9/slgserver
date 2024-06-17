package main

import (
	"fmt"
	"os"

	"github.com/llr104/slgserver/config"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/gateserver"
	"github.com/llr104/slgserver/server/gateserver/controller"
)

const gateServer string = "gateserver"

func getGateServerAddr() string {
	host := config.File.MustValue(gateServer, "host", "")
	port := config.File.MustValue(gateServer, "port", "8004")
	return host + ":" + port
}

func main() {
	fmt.Println(os.Getwd())
	gateserver.Init()
	needSecret := config.File.MustBool(gateServer, "need_secret", false)
	s := net.NewServer(getGateServerAddr(), needSecret)
	s.Router(gateserver.MyRouter)
	s.SetOnBeforeClose(controller.GHandle.OnServerConnClose)
	s.Start()
}
