package main

import (
	"fmt"
	"os"

	"github.com/llr104/slgserver/config"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/loginserver"
)

const loginServer string = "loginserver"

func getLoginServerAddr() string {
	host := config.File.MustValue(loginServer, "host", "")
	port := config.File.MustValue(loginServer, "port", "8003")
	return host + ":" + port
}

func main() {
	fmt.Println(os.Getwd())
	loginserver.Init()
	needSecret := config.File.MustBool(loginServer, "need_secret", false)
	s := net.NewServer(getLoginServerAddr(), needSecret)
	s.Router(loginserver.MyRouter)
	s.Start()
}
