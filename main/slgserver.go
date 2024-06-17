package main

import (
	"fmt"
	"os"

	"github.com/llr104/slgserver/config"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/run"
)

const SLGServer string = "slgserver"

func getServerAddr() string {
	host := config.File.MustValue(SLGServer, "host", "")
	port := config.File.MustValue(SLGServer, "port", "8001")
	return host + ":" + port
}

func main() {
	fmt.Println(os.Getwd())
	run.Init()
	needSecret := config.File.MustBool(SLGServer, "need_secret", false)
	s := net.NewServer(getServerAddr(), needSecret)
	s.Router(run.MyRouter)
	s.Start()
}
