package rtsp_server

import (
	"fmt"
	"github.com/yangxianzhi/my-streaming-server/rtsp"
	"github.com/yangxianzhi/my-streaming-server/sdp"
	"net"
	"strings"
)

type RTSPClientConnection struct {
	socket         net.Conn
	localPort      string
	remotePort     string
	localAddr      string
	remoteAddr     string
	currentCSeq    string
	sessionIDStr   string
	responseBuffer string
	//clientSession  *RTSPClientSession
	server         *RTSPServer
	//digest         *auth.Digest
	sdpInfo			sdp.Info
	reqInfo			rtsp.Request
}

func newRTSPClientConnection(server *RTSPServer, socket net.Conn) *RTSPClientConnection {
	localAddr := strings.Split(socket.LocalAddr().String(), ":")
	remoteAddr := strings.Split(socket.RemoteAddr().String(), ":")
	return &RTSPClientConnection{
		server:     server,
		socket:     socket,
		localAddr:  localAddr[0],
		localPort:  localAddr[1],
		remoteAddr: remoteAddr[0],
		remotePort: remoteAddr[1],
		//digest:     auth.NewDigest(),
	}
}

func (c *RTSPClientConnection) handleMethodOptions() {
	c.responseBuffer = fmt.Sprintf("RTSP/1.0 200 OK\r\n"+
		"CSeq: %s\r\n"+
		"Server:%s %s\r\n"+
		"%s%s\r\n",
		c.currentCSeq, SERVER, VERSION, rtsp.DateHeader(), rtsp.PublicHeader())
}

func (c *RTSPClientConnection) handleMethodAnnounce() {
	c.responseBuffer = fmt.Sprintf("RTSP/1.0 200 OK\r\n"+
		"CSeq: %s\r\n"+
		"Server:%s %s\r\n"+
		"%s\r\n",
		c.currentCSeq, SERVER, VERSION, rtsp.DateHeader())
}

func (c *RTSPClientConnection) handleCommandGetParameter() {
	c.setRTSPResponse("200 OK")
}

func (c *RTSPClientConnection) handleCommandSetParameter() {
	c.setRTSPResponse("200 OK")
}

func (c *RTSPClientConnection) handleCommandNotFound() {
	c.setRTSPResponse("404 Stream Not Found")
}

func (c *RTSPClientConnection) handleCommandSessionNotFound() {
	c.setRTSPResponse("454 Session Not Found")
}

func (c *RTSPClientConnection) handleCommandUnsupportedTransport() {
	c.setRTSPResponse("461 Unsupported Transport")
}
func (c *RTSPClientConnection) setRTSPResponse(responseStr string) {
	c.responseBuffer = fmt.Sprintf("RTSP/1.0 %s\r\n"+
		"CSeq: %s\r\n"+
		"%s\r\n",
		responseStr, c.currentCSeq, rtsp.DateHeader())
}

func (c *RTSPClientConnection) setRTSPResponseWithSessionID(responseStr string, sessionID string) {
	c.responseBuffer = fmt.Sprintf("RTSP/1.0 %s\r\n"+
		"CSeq: %s\r\n"+
		"%sSession: %s\r\n\r\n",
		responseStr, c.currentCSeq, rtsp.DateHeader(), sessionID)
}

func (c *RTSPClientConnection) incomingRequestHandler() {
	defer c.socket.Close()

	for {
		if req,err := rtsp.ReadRequest(c.socket); err != nil {
			break
		} else {
			c.currentCSeq = req.Header.Get(rtsp.Headers[rtsp.MySSCSeqHeader])
			c.sessionIDStr = req.Header.Get(rtsp.Headers[rtsp.MySSCSeqHeader])
			switch req.Method {
			case rtsp.OPTIONS:
				c.handleMethodOptions()
			case rtsp.ANNOUNCE:
				if _,err := sdp.ParseSdp(req.Body); err!=nil{
					break
				}
				c.handleMethodAnnounce()
			case rtsp.SETUP:
			}
			sendBytes, err := c.socket.Write([]byte(c.responseBuffer))
			if err != nil {
				fmt.Printf("failed to send response buffer.%d", sendBytes)
				break
			}
			fmt.Printf("send response:\n%s", c.responseBuffer)
		}
	}
}