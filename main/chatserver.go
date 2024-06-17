package main

import (
	"fmt"
	"github.com/llr104/slgserver/server/chatserver"
	"os"

	"github.com/llr104/slgserver/config"
	"github.com/llr104/slgserver/net"
)

const chatServer string = "chatserver"

func getChatServerAddr() string {
	host := config.File.MustValue(chatServer, "host", "")
	port := config.File.MustValue(chatServer, "port", "8002")
	return host + ":" + port
}

func main() {
	fmt.Println(os.Getwd())
	chatserver.Init()
	needSecret := config.File.MustBool(chatServer, "need_secret", false)
	s := net.NewServer(getChatServerAddr(), needSecret)
	s.Router(chatserver.MyRouter)
	s.Start()
}
