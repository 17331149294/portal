package v2

import (
	"crypto/md5"
	"fmt"
	"testing"
)

func TestRawAuth(t *testing.T) {
	hash := md5.New()
	hash.Write([]byte{2, 1, 0, 0, 0x14, 0xe, 0, 0, 0x6a, 0x78, 0x2b, 0xab, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	hash.Write([]byte("ctbri4008118114"))
	bts := hash.Sum(nil)
	for _, v := range bts {
		fmt.Printf(" %02x", v)
	}
}

// TestUnmarshal checks that a captured v2 packet is parsed with the correct
// header, 16-byte authenticator and attribute boundaries. (The original
// version tried to reproduce the wire authenticator via AuthBy with a
// fabricated placeholder {0xde...}; that comparison can never pass because the
// request authenticator is unknown here. Real authenticator verification is
// covered by security_test.go with a self-consistent request/response pair.)
func TestUnmarshal(t *testing.T) {
	bts := []byte{0x02, 0x03, 0x00, 0x00, 0x9e, 0xd8, 0x0a, 0xb7, 0x7c, 0x7f, 0x4e, 0xd8, 0x00, 0x00, 0x00, 0x02, 0x23, 0x5e, 0xed, 0x66, 0x3f, 0x9b, 0x18, 0x6e, 0xbc, 0x64, 0xde, 0x52, 0xfd, 0x54, 0x20, 0x81, 0x01, 0x15, 0x40, 0x77, 0x6c, 0x61, 0x6e, 0x2d, 0x78, 0x69, 0x6e, 0x6a, 0x69, 0x65, 0x6b, 0x6f, 0x75, 0x2d, 0x6e, 0x65, 0x77, 0x04, 0x12, 0x95, 0xac, 0xed, 0xcb, 0xb9, 0xa1, 0x25, 0x51, 0xc9, 0xed, 0x8d, 0x4c, 0xbc, 0x45, 0x3a, 0xf4}
	v := new(Version)
	m := v.Unmarshall(bts)
	mq := m.(*T_Message)
	if mq.Header.Version != 0x02 {
		t.Errorf("version = %d, want 2", mq.Header.Version)
	}
	if mq.Header.Type != 0x03 {
		t.Errorf("type = %d, want 3", mq.Header.Type)
	}
	if len(mq.Header.Authenticator) != 16 {
		t.Errorf("authenticator len = %d, want 16", len(mq.Header.Authenticator))
	}
	if len(mq.Attrs) != 2 {
		t.Fatalf("attrs = %d, want 2", len(mq.Attrs))
	}
	if mq.Attrs[0].AttrType != 0x01 || mq.Attrs[0].AttrLen != 19 {
		t.Errorf("attr[0] = type %d len %d, want type 1 len 19", mq.Attrs[0].AttrType, mq.Attrs[0].AttrLen)
	}
	if string(mq.Attrs[0].AttrStr) != "@wlan-xinjiekou-new" {
		t.Errorf("attr[0] data = %q, want %q", string(mq.Attrs[0].AttrStr), "@wlan-xinjiekou-new")
	}
	if mq.Attrs[1].AttrType != 0x04 || mq.Attrs[1].AttrLen != 16 {
		t.Errorf("attr[1] = type %d len %d, want type 4 len 16", mq.Attrs[1].AttrType, mq.Attrs[1].AttrLen)
	}
}

func TestAuth(t *testing.T) {
	bts := []byte{0x02, 0x02, 0x00, 0x00, 0x53, 0x6c, 0x00, 0x31, 0xc0, 0xa8, 0x0a, 0x03, 0x00, 0x00, 0x00, 0x01, 0x03, 0x12, 0xc7, 0x1a, 0x83, 0x1b, 0xec, 0x63, 0x34, 0xf9, 0x55, 0xd4, 0x84, 0x23, 0xb6, 0x2a, 0x5e, 0x39}
	v := new(Version)
	m := v.Unmarshall(bts)
	fmt.Println(m)
	mq := m.(*T_Message)
	fmt.Println(mq.Attrs)
}
