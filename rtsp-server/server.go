package rtsp_server

import (
	"fmt"
	"net"
	"runtime"
	"sync"
)

const (
	SERVER = "my-streaming-server"
	VERSION = "1.0"
)

type RTSPServer struct {
	rtspPort 				int
	rtspListen				*net.TCPListener
	sessionMutex           sync.Mutex
	clientSessions         map[string]*RTSPClientSession
}

func New() *RTSPServer {
	runtime.GOMAXPROCS(runtime.NumCPU())

	return &RTSPServer{
		clientSessions: make(map[string]*RTSPClientSession),
	}
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

func (server *RTSPServer) newClientConnection(conn net.Conn) {
	c := newRTSPClientConnection(server, conn)
	if c != nil {
		c.incomingRequestHandler()
	}
}

func (s *RTSPServer) getClientSession(sessionID string) (clientSession *RTSPClientSession, existed bool) {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()
	clientSession, existed = s.clientSessions[sessionID]
	return
}

func (s *RTSPServer) addClientSession(sessionID string, clientSession *RTSPClientSession) {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()
	s.clientSessions[sessionID] = clientSession
}

func (s *RTSPServer) removeClientSession(sessionID string) {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()
	delete(s.clientSessions, sessionID)
}
