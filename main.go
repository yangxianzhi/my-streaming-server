package main

import "github.com/yangxianzhi/my-streaming-server/rtsp-server"

func main(){
	server := rtsp_server.New()
	server.Listen(8554)
	server.Start()
	select{}
}
