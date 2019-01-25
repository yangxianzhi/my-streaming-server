package sdp

import (
	"testing"
)

const (
	sdp1 = "v=0\r\n" +
		"o=QTSS_Play_List 3668593260 3668611686 IN IP4 59.64.180.135\r\n" +
		"s=C:\\Program Files\\Darwin Streaming Server\\Playlists\\untitled\\unti@\r\n" +
		"c=IN IP4 127.0.0.1\r\n" +
		"b=AS:2097279\r\n" +
		"t=3757287422 3757373822\r\n" +
		"a=x-broadcastcontrol:RTSP\r\n" +
		"a=isma-compliance:2,2.0,2\r\n" +
		"m=video 0 RTP/AVP 96\r\n" +
		"b=AS:2097151\r\n" +
		"a=rtpmap:96 H264/90000\r\n" +
		"a=control:trackID=1\r\n" +
		"a=cliprect:0,0,480,380\r\n" +
		"a=framesize:96 380-480\r\n" +
		"a=fmtp:96 packetization-mode=1;profile-level-id=4D401E;sprop-parameter-sets=J01AHqkYMB73oA==,KM4C+IA=\r\n" +
		"a=mpeg4-esid:201\r\n" +
		"m=audio 0 RTP/AVP 97\r\n" +
		"b=AS:127\r\n" +
		"a=rtpmap:97 mpeg4-generic/48000/2\r\n" +
		"a=control:trackID=2\r\n" +
		"a=fmtp:97 profile-level-id=15;mode=AAC-hbr;sizelength=13;indexlength=3;indexdeltalength=3;config=1190\r\n" +
		"a=mpeg4-esid:101\r\n"

	sdp2 = "v=0\r\n" +
		"o=- 0 0 IN IP4 127.0.0.1\r\n" +
		"s=RTSP Session\r\n" +
		"c=IN IP4 192.168.199.137\r\n" +
		"t=0 0\r\n" +
		"a=tool:libavformat 57.83.100\r\n" +
		"m=video 0 RTP/AVP 96\r\n" +
		"a=rtpmap:96 H264/90000\r\n" +
		"a=fmtp:96 packetization-mode=1; sprop-parameter-sets=Z0IAH52oFAFum4CAgIE=,aM48gA==; profile-level-id=42001F\r\n" +
		"a=control:streamid=0\r\n" +
		"m=audio 0 RTP/AVP 8\r\n" +
		"b=AS:64\r\n" +
		"a=control:streamid=1\r\n"
)

func TestParseSdp(t *testing.T) {
	var tests = []struct {
		input string
	}{
		{sdp1},
		{sdp2},
	}
	for _, test := range tests {
		if _,err:=ParseSdp(test.input); err !=nil{
			t.Errorf("ParseSdp(%v)",err)
		}
	}
}
