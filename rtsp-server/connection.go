package rtsp_server

import (
	"fmt"

	"github.com/golang/go/src/pkg/strconv"
	"github.com/yangxianzhi/CommonUtilities"
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
	clientSession  *RTSPClientSession
	server         *RTSPServer
	//digest         *auth.Digest
	sdpInfo sdp.Info
	reqInfo rtsp.Request
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

const rtspBufferSize = 10000

func (c *RTSPClientConnection) incomingRequestHandler() {
	defer c.socket.Close()

	var isClose bool
	buffer := make([]byte, rtspBufferSize)
	for {
		//timeoutTCPConn := &RichConn{c.socket, (time.Duration(1) * time.Millisecond)}
		//b := bufio.NewReadWriter(bufio.NewReaderSize(timeoutTCPConn,rtspBufferSize),bufio.NewWriterSize(timeoutTCPConn,rtspBufferSize))
		//length, err := io.ReadFull(b, buffer)
		length, err := c.socket.Read(buffer)

		switch err {
		case nil:
			err = c.handleRequestBytes(buffer, length)
			if err != nil {
				fmt.Printf("Failed to handle Request Bytes: %v", err)
				isClose = true
			}
		default:
			fmt.Printf("default: %v", err)
			if err.Error() == "EOF" {
				isClose = true
			}
		}

		if isClose {
			break
		}
	}

	fmt.Printf("disconnected the connection[%s:%s].", c.remoteAddr, c.remotePort)
	if c.clientSession != nil {
		c.clientSession.destroy()
	}
}

func (c *RTSPClientConnection) handleRequestBytes(buffer []byte, length int) error {
	if length > 0 && buffer[0] == '$' {
		return nil
	}

	if req, err := rtsp.ReadRequest(buffer, length); err != nil {
		return err
	} else {
		c.currentCSeq = req.Header.Get(rtsp.Headers[rtsp.MySSCSeqHeader])
		c.sessionIDStr = req.Header.Get(rtsp.Headers[rtsp.MySSSessionHeader])
		switch req.Method {
		case rtsp.OPTIONS:
			c.handleMethodOptions()
		case rtsp.ANNOUNCE:
			contentLength, _ := strconv.Atoi(req.Header.Get(rtsp.Headers[rtsp.MySSContentLengthHeader]))
			if contentLength > 0 {
				if len(req.Body) < contentLength {
					buffer1 := make([]byte, rtspBufferSize)
					c.socket.Read(buffer1)
					req.Body = req.Body + string(buffer1[:])
				}
				if _, err := sdp.ParseSdp(req.Body); err != nil {
					break
				}
			}
			c.handleMethodAnnounce()
		case rtsp.SETUP:
			if c.sessionIDStr == "" {
				for {
					c.sessionIDStr = fmt.Sprintf("%08X", commonutilities.OurRandom32())
					if _, existed := c.server.getClientSession(c.sessionIDStr); !existed {
						break
					}
				}
				c.clientSession = c.newClientSession(c.sessionIDStr)
				c.server.addClientSession(c.sessionIDStr, c.clientSession)
			} else {
				var existed bool
				if c.clientSession, existed = c.server.getClientSession(c.sessionIDStr); !existed {
					c.handleCommandSessionNotFound()
				}
			}

			if c.clientSession != nil {
				c.clientSession.handleCommandSetup(req.UrlPreSuffix, req.UrlSuffix, reqStr)
			}
		}
		sendBytes, err := c.socket.Write([]byte(c.responseBuffer))
		if err != nil {
			fmt.Printf("failed to send response buffer.%d", sendBytes)
			return err
		}
		fmt.Printf("send response:\n%s", c.responseBuffer)
	}
	return nil
}

func (c *RTSPClientConnection) newClientSession(sessionID string) *RTSPClientSession {
	return newRTSPClientSession(c, sessionID)
}
