package main

import (
	"flag"
	"tcpproxy/internal/client"
	"tcpproxy/internal/server"
)

const ServerServer = "server"
const ServerClient = "client"

func main() {

	run := flag.String("run", "server", "需要运行的服务 server 或者 client")
	flag.Parse()
	switch *run {
	case ServerClient:
		client.RunClient()
		break
	case ServerServer:
		server.StartServer()
	}

}
