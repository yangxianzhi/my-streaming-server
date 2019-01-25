package rtsp_server

import (
	"fmt"
	"net"
	"runtime"
)

const (
	SERVER = "my-streaming-server"
	VERSION = "1.0"
)

type RTSPServer struct {
	rtspPort 				int
	rtspListen				*net.TCPListener

}

func New() *RTSPServer {
	runtime.GOMAXPROCS(runtime.NumCPU())

	return &RTSPServer{}
}

func (s *RTSPServer) Destroy() {
	s.rtspListen.Close()
}

func (server *RTSPServer) Listen(port int) (err error) {
	server.rtspPort = port

	server.rtspListen, err = server.setupOurSocket(port)

	return err
}

func (server *RTSPServer) Start() {
	go server.incomingConnectionHandler(server.rtspListen)
}

func (server *RTSPServer) setupOurSocket(port int) (*net.TCPListener, error) {
	tcpAddr := fmt.Sprintf("0.0.0.0:%d", port)
	addr, _ := net.ResolveTCPAddr("tcp", tcpAddr)

	return net.ListenTCP("tcp", addr)
}

func (server *RTSPServer) incomingConnectionHandler(l *net.TCPListener) {
	for {
		tcpConn, err := l.AcceptTCP()
		if err != nil {
			fmt.Printf("failed to accept client.%s", err.Error())
			continue
		}

		tcpConn.SetReadBuffer(50 * 1024)

		// Create a new object for handling server RTSP connection:
		go server.newClientConnection(tcpConn)
	}
}

func (s *RTSPServer) newClientConnection(conn net.Conn) {
	c := newRTSPClientConnection(s, conn)
	if c != nil {
		c.incomingRequestHandler()
	}
}
