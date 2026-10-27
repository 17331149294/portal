package v2

import (
	"net"
	"testing"

	"github.com/17331149294/portal/huawei/portal"
)

// Regression tests: malformed UDP packets must never panic the process
// (previously AttrLen<2 would make a negative slice and panic).
func TestUnmarshallNoPanicOnMalformed(t *testing.T) {
	v := new(Version)
	cases := [][]byte{
		{0x02}, // truncated header
		makeHeaderWithAttrs(0x03, 0x01),  // AttrLen=1 (<2)
		makeHeaderWithAttrs(0x03, 0xff),  // AttrLen=255, no data
	}
	for _, c := range cases {
		v.Unmarshall(c) // must not panic
	}
}

func makeHeaderWithAttrs(attrType, attrLen byte) []byte {
	b := make([]byte, 0, 40)
	b = append(b, 0x02, 0x02, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00)
	b = append(b, 0xc0, 0xa8, 0x01, 0x01, 0x00, 0x00, 0x00, 0x01) // 16-byte header
	b = append(b, make([]byte, 16)...)                             // 16-byte authenticator
	b = append(b, attrType, attrLen)
	return b
}

// Regression test: CheckFor must verify the authenticator on the success
// path too (previously ErrCode==0 returned nil without any verification, so
// a forged success response would be trusted).
func TestCheckForRejectsForgedSuccess(t *testing.T) {
	secret := "test-secret"
	req := newMessage(portal.REQ_CHALLENGE, net.IPv4(10, 0, 0, 1), 1234, 0)
	req.AuthBy(secret)

	// forged success response: wrong authenticator, ErrCode=0
	forged := newMessage(portal.ACK_CHALLENGE, net.IPv4(10, 0, 0, 1), 1234, 0)
	forged.Header.ErrCode = 0
	forged.Header.Authenticator = make([]byte, 16)
	if err := forged.CheckFor(req, secret); err == nil {
		t.Error("forged success response should be rejected")
	}

	// valid success response must pass
	valid := newMessage(portal.ACK_CHALLENGE, net.IPv4(10, 0, 0, 1), 1234, 0)
	valid.Header.ErrCode = 0
	valid.Header.Authenticator = req.Header.Authenticator
	valid.AuthBy(secret)
	if err := valid.CheckFor(req, secret); err != nil {
		t.Errorf("valid success response should pass, got %v", err)
	}
}
