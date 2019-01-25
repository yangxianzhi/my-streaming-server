package sdp

import (
	"bufio"
	"errors"
	"github.com/yangxianzhi/CommonUtilities"
	"io"
	"net"
	"strconv"
	"strings"
)

type RTPPayloadType uint32

const (
	UnknownPayloadType RTPPayloadType = iota
	VideoPayloadType
	AudioPayloadType
)

const (
	SDPTimeControl = iota
	RTSPSessionControl
)

type StreamInfo struct {
	fSrcIPAddr      string         // Src IP address of content (this may be 0 if not known for sure)
	fDestIPAddr     string         // Dest IP address of content (destination IP addr for source broadcast!)
	fPort           uint16         // Dest (RTP) port of source content
	fTimeToLive     uint16         // Ttl for this stream
	fPayloadType    RTPPayloadType // Payload type of this stream
	fPayloadName    string         // Payload name of this stream
	fTrackID        uint32         // ID of this stream
	fTrackName      string         //Track Name of this stream
	fBufferDelay    float32        // buffer delay (default is 3 seconds)
	fIsTCP          bool           // Is this a TCP broadcast? If this is the case, the port and ttl are not valid
	fSetupToReceive bool           // If true then a push to the server is setup on this stream.
	fTimeScale      uint32
}
type OutputInfo struct {
	fDestAddr     string   // Destination address to forward the input onto
	fLocalAddr    string   // Address of local interface to send out on (may be 0)
	fTimeToLive   uint16   // Time to live for resulting output (if multicast)
	fPortArray    []uint16 // 1 destination RTP port for each Stream.
	fNumPorts     uint32   // Size of the fPortArray (usually equal to fNumStreams)
	fBasePort     uint16   // The base destination RTP port - for i=1 to fNumStreams fPortArray[i] = fPortArray[i-1] + 2
	fAlreadySetup bool     // A flag used in QTSSReflectorModule.cpp
}

func (o *OutputInfo) Equal(info OutputInfo) bool {
	if o.fDestAddr == info.fDestAddr && o.fLocalAddr == info.fLocalAddr && o.fTimeToLive == info.fTimeToLive {
		if o.fBasePort != 0 && o.fBasePort == info.fBasePort {
			return true
		} else if o.fNumPorts == 0 || (o.fNumPorts == info.fNumPorts && o.fPortArray[0] == info.fPortArray[0]) {
			return true
		}
	}
	return false
}

type Info struct {
	Version               int
	Originator            string
	SessionName           string
	SessionInformation    string
	URI                   string
	Email                 string
	Phone                 string
	ConnectionInformation string
	BandwidthInformation  string
	StreamInfoArray       []*StreamInfo
	OutputInfoArray       []*OutputInfo
	SessionControlType    uint32
	StartTimeUnixSecs	  uint32
	EndTimeUnixSecs       uint32
	HasValidTime          bool
}

func ParseSdp(r io.Reader) (Info, error) {
	var packet Info
	hasGlobalStreamInfo := false
	var globalStreamInfo StreamInfo
	var theStreamIndex, currentTrack uint32 = 0, 1
	for {
		if b, ok := r.(*bufio.Reader); ok {
			// TODO: allow CR, LF, or CRLF
			if s, err := b.ReadString('\n'); err != nil {
				return packet, err
			} else {
				parts := strings.SplitN(s, "=", 2)
				if len(parts) == 2 {
					if len(parts[0]) != 1 {
						return packet, errors.New("SDP only allows 1-character variables")
					}
					key := parts[0]
					value := parts[1]
					switch key {
					// version
					case "v":
						ver, err := strconv.Atoi(strings.TrimRight(value, "\r\n"))
						if err != nil {
							return packet, err
						}
						packet.Version = ver
					// owner/creator and session identifier
					case "o":
						// o=<username> <session id> <version> <network type> <address type> <address>
						// TODO: parse this
						packet.Originator = strings.TrimRight(value, "\r\n")
					// session name
					case "s":
						packet.SessionName = strings.TrimRight(value, "\r\n")
					// session information
					case "i":
						packet.SessionInformation = strings.TrimRight(value, "\r\n")
					// URI of description
					case "u":
						packet.URI = strings.TrimRight(value, "\r\n")
					// email address
					case "e":
						packet.Email = strings.TrimRight(value, "\r\n")
					// phone number
					case "p":
						packet.Phone = strings.TrimRight(value, "\r\n")
					case "t":
						tParser := commonutilities.New(value)
						tParser.ConsumeUntil(commonutilities.DigitMask)
						_,ntpStart := tParser.ConsumeInteger()
						tParser.ConsumeUntil(commonutilities.DigitMask)
						_,ntpEnd := tParser.ConsumeInteger()
						if ntpStart > 0 && ntpEnd > 0 && ntpStart > ntpEnd {
							continue
						}
						var startTimeUnixSecs, endTimeUnixSecs uint32
						if ntpStart != 0 && isValidNTPSecs(ntpStart) { // allow anything less than 1970
							startTimeUnixSecs = ntpSecs_to_UnixSecs(ntpStart) // convert to 1970 time
						}
						if ntpEnd !=0 && !isValidNTPSecs(ntpEnd) {
							continue
						}
						if ntpEnd != 0 { // convert to 1970 time
							endTimeUnixSecs = ntpSecs_to_UnixSecs(ntpEnd)
						}
						packet.StartTimeUnixSecs = startTimeUnixSecs
						packet.EndTimeUnixSecs = endTimeUnixSecs
						packet.HasValidTime = true
					// connection information - not required if included in all media
					case "c":
						// TODO: parse this
						packet.ConnectionInformation = value
						values := strings.SplitN(value[7:], "/", 3)
						var tempIPAddr string
						if len(values) > 0 {
							tempIPAddr = values[0]
						}
						tempTtl := 15
						if len(values) > 1 {
							tempTtl, err = strconv.Atoi(values[0])
							if tempTtl >= 65535 || tempTtl < 0 {
								panic("Ttl invalid")
							}
						}
						if theStreamIndex > 0 {
							if len(packet.StreamInfoArray) < int(theStreamIndex) {
								packet.StreamInfoArray = append(packet.StreamInfoArray, new(StreamInfo))
							}
							packet.StreamInfoArray[theStreamIndex-1].fDestIPAddr = tempIPAddr
							packet.StreamInfoArray[theStreamIndex-1].fTimeToLive = uint16(tempTtl)
						} else {
							globalStreamInfo.fDestIPAddr = tempIPAddr
							globalStreamInfo.fTimeToLive = uint16(tempTtl)
							hasGlobalStreamInfo = true
						}
					// bandwidth information
					case "b":
						// TODO: parse this
						packet.BandwidthInformation = strings.TrimRight(value, "\r\n")
					// media info
					case "m":
						if len(packet.StreamInfoArray) < int(theStreamIndex)+1 {
							packet.StreamInfoArray = append(packet.StreamInfoArray, new(StreamInfo))
						}
						if hasGlobalStreamInfo {
							packet.StreamInfoArray[theStreamIndex].fDestIPAddr = globalStreamInfo.fDestIPAddr
							packet.StreamInfoArray[theStreamIndex].fTimeToLive = globalStreamInfo.fTimeToLive
						}
						packet.StreamInfoArray[theStreamIndex].fTrackID = currentTrack
						currentTrack++

						mParser := commonutilities.New(value)
						theStreamType := mParser.ConsumeWord()
						if theStreamType == "audio" {
							packet.StreamInfoArray[theStreamIndex].fPayloadType = AudioPayloadType
						} else if theStreamType == "video" {
							packet.StreamInfoArray[theStreamIndex].fPayloadType = VideoPayloadType
						}
						mParser.ConsumeUntil(commonutilities.DigitMask)
						_, tempPort := mParser.ConsumeInteger()
						if tempPort > 0 && tempPort < 65536 {
							packet.StreamInfoArray[theStreamIndex].fPort = uint16(tempPort)
						}
						// find out whether this is TCP or UDP
						mParser.ConsumeWhitespace()
						transportID := mParser.ConsumeUntilStop(' ')
						if transportID == "RTP/AVP/TCP" {
							packet.StreamInfoArray[theStreamIndex].fIsTCP = true
						}
						theStreamIndex++
					case "a":
						aParser := commonutilities.New(value)
						aLineType := aParser.ConsumeWord()
						if aLineType == "x-broadcastcontrol" {
							// found a control line for the broadcast (delete at time or delete at end of broadcast/server startup)
							// qtss_printf("found =%s\n",sBroadcastControlStr);
							aParser.ConsumeUntil(commonutilities.WordMask)
							sessionControlType := aParser.ConsumeWord()
							if sessionControlType == "RTSP" {
								packet.SessionControlType = RTSPSessionControl
							} else if sessionControlType == "TIME" {
								packet.SessionControlType = SDPTimeControl
							}
						}

						//if we haven't even hit an 'm' line yet, just ignore all 'a' lines
						if theStreamIndex == 0 {
							continue
						}

						if aLineType == "rtpmap" {
							//mark the codec type if this line has a codec name on it. If we already
							//have a codec type for this track, just ignore this line
							if _, ok := aParser.GetThru(' '); len(packet.StreamInfoArray[theStreamIndex-1].fPayloadName) == 0 && ok {
								if payloadName, isOk := aParser.GetThruEOL(); isOk {
									packet.StreamInfoArray[theStreamIndex-1].fPayloadName = payloadName
								}
							}
						} else if aLineType == "control" {
							if _, ok := aParser.GetThru(':'); ok {
								if trackName, isOk := aParser.GetThruEOL(); isOk {
									packet.StreamInfoArray[theStreamIndex-1].fTrackName = trackName
									trackParser := commonutilities.New(trackName)
									trackParser.ConsumeUntilStop('=')
									trackParser.ConsumeUntil(commonutilities.DigitMask)
									_, packet.StreamInfoArray[theStreamIndex-1].fTrackID = trackParser.ConsumeInteger()
								}
							}
						} else if aLineType == "x-bufferdelay" {
							aParser.ConsumeUntil(commonutilities.DigitMask)
							globalStreamInfo.fBufferDelay = aParser.ConsumeFloat()
						}
					}
				}
			}
		}
	}
	return packet, nil
}

// Convert uint to net.IP http://www.sharejs.com
func inet_ntoa(ipnr int64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

// Convert net.IP to int64 ,  http://www.sharejs.com
func inet_aton(ipnr net.IP) int64 {
	bits := strings.Split(ipnr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

const kNTP_Offset_From_1970 = 2208988800
func isValidNTPSecs(time uint32) bool {
	if  time >= kNTP_Offset_From_1970 {
		return true
	} else {
		return false
	}
}

func ntpSecs_to_UnixSecs(time uint32) uint32 {
	return (uint32) (time - kNTP_Offset_From_1970)
}
func unixSecs_to_NTPSecs(time uint32) uint32 {
	return (uint32) (time + kNTP_Offset_From_1970)
}
