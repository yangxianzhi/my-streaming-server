package rtsp

import (
	"bufio"
	"fmt"
	"github.com/yangxianzhi/CommonUtilities"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// Client to server for presentation and stream objects; recommended
	DESCRIBE = "DESCRIBE"
	// Bidirectional for client and stream objects; optional
	ANNOUNCE = "ANNOUNCE"
	// Bidirectional for client and stream objects; optional
	GET_PARAMETER = "GET_PARAMETER"
	// Bidirectional for client and stream objects; required for Client to server, optional for server to client
	OPTIONS = "OPTIONS"
	// Client to server for presentation and stream objects; recommended
	PAUSE = "PAUSE"
	// Client to server for presentation and stream objects; required
	PLAY = "PLAY"
	// Client to server for presentation and stream objects; optional
	RECORD = "RECORD"
	// Server to client for presentation and stream objects; optional
	REDIRECT = "REDIRECT"
	// Client to server for stream objects; required
	SETUP = "SETUP"
	// Bidirectional for presentation and stream objects; optional
	SET_PARAMETER = "SET_PARAMETER"
	// Client to server for presentation and stream objects; required
	TEARDOWN = "TEARDOWN"
)

const (
	// all requests
	Continue = 100

	// all requests
	OK = 200
	// RECORD
	Created = 201
	// RECORD
	LowOnStorageSpace = 250

	// all requests
	MultipleChoices = 300
	// all requests
	MovedPermanently = 301
	// all requests
	MovedTemporarily = 302
	// all requests
	SeeOther = 303
	// all requests
	UseProxy = 305

	// all requests
	BadRequest = 400
	// all requests
	Unauthorized = 401
	// all requests
	PaymentRequired = 402
	// all requests
	Forbidden = 403
	// all requests
	NotFound = 404
	// all requests
	MethodNotAllowed = 405
	// all requests
	NotAcceptable = 406
	// all requests
	ProxyAuthenticationRequired = 407
	// all requests
	RequestTimeout = 408
	// all requests
	Gone = 410
	// all requests
	LengthRequired = 411
	// DESCRIBE, SETUP
	PreconditionFailed = 412
	// all requests
	RequestEntityTooLarge = 413
	// all requests
	RequestURITooLong = 414
	// all requests
	UnsupportedMediaType = 415
	// SETUP
	Invalidparameter = 451
	// SETUP
	IllegalConferenceIdentifier = 452
	// SETUP
	NotEnoughBandwidth = 453
	// all requests
	SessionNotFound = 454
	// all requests
	MethodNotValidInThisState = 455
	// all requests
	HeaderFieldNotValid = 456
	// PLAY
	InvalidRange = 457
	// SET_PARAMETER
	ParameterIsReadOnly = 458
	// all requests
	AggregateOperationNotAllowed = 459
	// all requests
	OnlyAggregateOperationAllowed = 460
	// all requests
	UnsupportedTransport = 461
	// all requests
	DestinationUnreachable = 462

	// all requests
	InternalServerError = 500
	// all requests
	NotImplemented = 501
	// all requests
	BadGateway = 502
	// all requests
	ServiceUnavailable = 503
	// all requests
	GatewayTimeout = 504
	// all requests
	RTSPVersionNotSupported = 505
	// all requests
	OptionNotsupport = 551
)

const maxCommandNum = 11

// Handler routines for specific RTSP commands:
var AllowedMethods = [maxCommandNum]string{
	OPTIONS,
	ANNOUNCE,
	DESCRIBE,
	SETUP,
	TEARDOWN,
	PLAY,
	PAUSE,
	RECORD,
	REDIRECT,
	GET_PARAMETER,
	SET_PARAMETER,
}
var Headers = []string{
	"Accept",
	"Cseq",
	"User-Agent",
	"Transport",
	"Session",
	"Range",

	"Accept-Encoding",
	"Accept-Language",
	"Authorization",
	"Bandwidth",
	"Blocksize",
	"Cache-Control",
	"Conference",
	"Connection",
	"Content-Base",
	"Content-Encoding",
	"Content-Language",
	"Content-length",
	"Content-Location",
	"Content-Type",
	"Date",
	"Expires",
	"From",
	"Host",
	"If-Match",
	"If-Modified-Since",
	"Last-Modified",
	"Location",
	"Proxy-Authenticate",
	"Proxy-Require",
	"Referer",
	"Retry-After",
	"Require",
	"RTP-Info",
	"Scale",
	"Speed",
	"Timestamp",
	"Vary",
	"Via",
	"Allow",
	"Public",
	"Server",
	"Unsupported",
	"WWW-Authenticate",
	",",
	"x-Retransmit",
	"x-Accept-Retransmit",
	"x-RTP-Meta-Info",
	"x-Transport-Options",
	"x-Packet-Range",
	"x-Prebuffer",
	"x-Dynamic-Rate",
	"x-Accept-Dynamic-Rate",
	// DJM PROTOTYPE
	"x-Random-Data-Size",
}

const (
	//These are the common request headers (optimized)
	MySSAcceptHeader    = 0
	MySSCSeqHeader      = 1
	MySSUserAgentHeader = 2
	MySSTransportHeader = 3
	MySSSessionHeader   = 4
	MySSRangeHeader     = 5
	MySSNumVIPHeaders   = 6

	//Other request headers
	MySSAcceptEncodingHeader    = 6
	MySSAcceptLanguageHeader    = 7
	MySSAuthorizationHeader     = 8
	MySSBandwidthHeader         = 9
	MySSBlockSizeHeader         = 10
	MySSCacheControlHeader      = 11
	MySSConferenceHeader        = 12
	MySSConnectionHeader        = 13
	MySSContentBaseHeader       = 14
	MySSContentEncodingHeader   = 15
	MySSContentLanguageHeader   = 16
	MySSContentLengthHeader     = 17
	MySSContentLocationHeader   = 18
	MySSContentTypeHeader       = 19
	MySSDateHeader              = 20
	MySSExpiresHeader           = 21
	MySSFromHeader              = 22
	MySSHostHeader              = 23
	MySSIfMatchHeader           = 24
	MySSIfModifiedSinceHeader   = 25
	MySSLastModifiedHeader      = 26
	MySSLocationHeader          = 27
	MySSProxyAuthenticateHeader = 28
	MySSProxyRequireHeader      = 29
	MySSRefererHeader           = 30
	MySSRetryAfterHeader        = 31
	MySSRequireHeader           = 32
	MySSRTPInfoHeader           = 33
	MySSScaleHeader             = 34
	MySSSpeedHeader             = 35
	MySSTimestampHeader         = 36
	MySSVaryHeader              = 37
	MySSViaHeader               = 38
	MySSNumRequestHeaders       = 39

	//Additional response headers
	MySSAllowHeader           = 39
	MySSPublicHeader          = 40
	MySSServerHeader          = 41
	MySSUnsupportedHeader     = 42
	MySSWWWAuthenticateHeader = 43
	MySSSameAsLastHeader      = 44

	//Newly added headers
	MySSExtensionHeaders = 45

	MySSXRetransmitHeader        = 45
	MySSXAcceptRetransmitHeader  = 46
	MySSXRTPMetaInfoHeader       = 47
	MySSXTransportOptionsHeader  = 48
	MySSXPacketRangeHeader       = 49
	MySSXPreBufferHeader         = 50
	MySSXDynamicRateHeader       = 51
	MySSXAcceptDynamicRateHeader = 52

	// QT Player random data request
	MySSXRandomDataSizeHeader = 53

	MySSNumHeaders    = 54
	MySSIllegalHeader = 54
)

// DateHeader A "Date:" header that can be used in a RTSP (or HTTP) response
func DateHeader() string {
	return fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123))
}

func PublicHeader() string {
	return fmt.Sprintf("Public: %s\r\n", strings.Join(AllowedMethods[0:], ","))
}

type ResponseWriter interface {
	http.ResponseWriter
}

type Request struct {
	Method        string
	URL           *url.URL
	Proto         string
	ProtoMajor    int
	ProtoMinor    int
	Header        http.Header
	ContentLength int
	Body          string
}

func (r Request) String() string {
	s := fmt.Sprintf("%s %s %s/%d.%d\r\n", r.Method, r.URL, r.Proto, r.ProtoMajor, r.ProtoMinor)
	for k, v := range r.Header {
		for _, v := range v {
			s += fmt.Sprintf("%s: %s\r\n", k, v)
		}
	}
	s += "\r\n" + r.Body
	return s
}

func NewRequest(method, urlStr, cSeq string, body string) (*Request, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req := &Request{
		Method:     method,
		URL:        u,
		Proto:      "RTSP",
		ProtoMajor: 1,
		ProtoMinor: 0,
		Header:     map[string][]string{"CSeq": []string{cSeq}},
		Body:       body,
	}
	return req, nil
}

type Session struct {
	cSeq    int
	conn    net.Conn
	session string
}

func NewSession() *Session {
	return &Session{}
}

func (s *Session) nextCSeq() string {
	s.cSeq++
	return strconv.Itoa(s.cSeq)
}

func (s *Session) Describe(urlStr string) (*Response, error) {
	req, err := NewRequest(DESCRIBE, urlStr, s.nextCSeq(), "")
	if err != nil {
		panic(err)
	}

	req.Header.Add(Headers[MySSAcceptHeader], "application/sdp")

	if s.conn == nil {
		s.conn, err = net.Dial("tcp", req.URL.Host)
		if err != nil {
			return nil, err
		}
	}

	_, err = io.WriteString(s.conn, req.String())
	if err != nil {
		return nil, err
	}
	return ReadResponse(s.conn)
}

func (s *Session) Options(urlStr string) (*Response, error) {
	req, err := NewRequest(OPTIONS, urlStr, s.nextCSeq(), "")
	if err != nil {
		panic(err)
	}

	if s.conn == nil {
		s.conn, err = net.Dial("tcp", req.URL.Host)
		if err != nil {
			return nil, err
		}
	}

	_, err = io.WriteString(s.conn, req.String())
	if err != nil {
		return nil, err
	}
	return ReadResponse(s.conn)
}

func (s *Session) Setup(urlStr, transport string) (*Response, error) {
	req, err := NewRequest(SETUP, urlStr, s.nextCSeq(), "")
	if err != nil {
		panic(err)
	}

	req.Header.Add(Headers[MySSTransportHeader], transport)

	if s.conn == nil {
		s.conn, err = net.Dial("tcp", req.URL.Host)
		if err != nil {
			return nil, err
		}
	}

	_, err = io.WriteString(s.conn, req.String())
	if err != nil {
		return nil, err
	}
	resp, err := ReadResponse(s.conn)
	s.session = resp.Header.Get(Headers[MySSSessionHeader])
	return resp, err
}

func (s *Session) Play(urlStr, sessionId string) (*Response, error) {
	req, err := NewRequest(PLAY, urlStr, s.nextCSeq(), "")
	if err != nil {
		panic(err)
	}

	req.Header.Add(Headers[MySSSessionHeader], sessionId)

	if s.conn == nil {
		s.conn, err = net.Dial("tcp", req.URL.Host)
		if err != nil {
			return nil, err
		}
	}

	_, err = io.WriteString(s.conn, req.String())
	if err != nil {
		return nil, err
	}
	return ReadResponse(s.conn)
}

type closer struct {
	*bufio.Reader
	r io.Reader
}

func (c closer) Close() error {
	if c.Reader == nil {
		return nil
	}
	defer func() {
		c.Reader = nil
		c.r = nil
	}()
	if r, ok := c.r.(io.ReadCloser); ok {
		return r.Close()
	}
	return nil
}

func ParseRTSPVersion(s string) (proto string, major int, minor int, err error) {
	parts := strings.SplitN(s, "/", 2)
	proto = parts[0]
	parts = strings.SplitN(parts[1], ".", 2)
	if major, err = strconv.Atoi(parts[0]); err != nil {
		return
	}
	if minor, err = strconv.Atoi(parts[0]); err != nil {
		return
	}
	return
}

// super simple RTSP parser; would be nice if net/http would allow more general parsing
func ReadRequest(buffer []byte, length int) (req *Request, err error) {
	req = new(Request)
	req.Header = make(map[string][]string)

	reqParser := commonutilities.New(string(buffer[:length]))
	var s string
	var ok bool
	if s, ok = reqParser.GetThruEOL(); !ok {
		return
	}
	parts := strings.SplitN(s, " ", 3)
	req.Method = parts[0]
	if req.URL, err = url.Parse(parts[1]); err != nil {
		return
	}

	req.Proto, req.ProtoMajor, req.ProtoMinor, err = ParseRTSPVersion(parts[2])
	if err != nil {
		return
	}

	// read headers
	for {
		if s, ok = reqParser.GetThruEOL(); !ok || s == "" {
			break
		}

		parts := strings.SplitN(s, ":", 2)
		req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	req.ContentLength, _ = strconv.Atoi(req.Header.Get(Headers[MySSContentLengthHeader]))
	if req.ContentLength > 0 {
		fmt.Println(Headers[MySSContentLengthHeader], req.ContentLength)
		req.Body = reqParser.ConsumeLength(req.ContentLength)
	}
	return
}

type Response struct {
	Proto      string
	ProtoMajor int
	ProtoMinor int

	StatusCode int
	Status     string

	ContentLength int64

	Header http.Header
	Body   *bufio.Reader
}

func (res Response) String() string {
	s := fmt.Sprintf("%s/%d.%d %d %s\n", res.Proto, res.ProtoMajor, res.ProtoMinor, res.StatusCode, res.Status)
	for k, v := range res.Header {
		for _, v := range v {
			s += fmt.Sprintf("%s: %s\n", k, v)
		}
	}
	return s
}

func ReadResponse(r io.Reader) (res *Response, err error) {
	res = new(Response)
	res.Header = make(map[string][]string)

	b := bufio.NewReader(r)
	var s string

	// TODO: allow CR, LF, or CRLF
	if s, err = b.ReadString('\n'); err != nil {
		return
	}

	parts := strings.SplitN(s, " ", 3)
	res.Proto, res.ProtoMajor, res.ProtoMinor, err = ParseRTSPVersion(parts[0])
	if err != nil {
		return
	}

	if res.StatusCode, err = strconv.Atoi(parts[1]); err != nil {
		return
	}

	res.Status = strings.TrimSpace(parts[2])

	// read headers
	for {
		if s, err = b.ReadString('\n'); err != nil {
			return
		} else if s = strings.TrimRight(s, "\r\n"); s == "" {
			break
		}

		parts := strings.SplitN(s, ":", 2)
		res.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	res.ContentLength, _ = strconv.ParseInt(res.Header.Get(Headers[MySSContentLengthHeader]), 10, 64)
	if res.ContentLength > 0 {
		res.Body = b
	}
	return
}
