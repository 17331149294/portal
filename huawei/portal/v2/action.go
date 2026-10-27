package v2

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"github.com/17331149294/portal/huawei/portal"
	"net"
)

type Version struct{}

func (v *Version) NewChallenge(userip net.IP, secret string) portal.Message {
	msg := newMessage(portal.REQ_CHALLENGE, userip, portal.NewSerialNo(), 0)
	msg.AuthBy(secret)
	return msg
}

func (v *Version) NewLogout(userip net.IP, secret string) portal.Message {
	msg := newMessage(portal.REQ_LOGOUT, userip, portal.NewSerialNo(), 0)
	msg.AuthBy(secret)
	return msg
}

func (v *Version) NewAffAckAuth(userip net.IP, secret string, serial uint16, reqid uint16) portal.Message {
	msg := newMessage(portal.AFF_ACK_AUTH, userip, serial, reqid)
	msg.AuthBy(secret)
	return msg
}

func (v *Version) NewAuth(userip net.IP, secret string, username []byte, userpwd []byte, req uint16, cha []byte) portal.Message {
	msg := newMessage(3, userip, portal.NewSerialNo(), req)
	msg.Header.AttrNum = 3
	hash := md5.New()
	hash.Write([]byte{byte(req)})
	hash.Write(userpwd)
	hash.Write(cha)
	cpwd := hash.Sum(nil)
	msg.Attrs = []T_Attr{
		{AttrType: byte(1), AttrLen: byte(len(username)), AttrStr: username},
		{AttrType: byte(3), AttrLen: byte(len(cha)), AttrStr: cha},
		{AttrType: byte(4), AttrLen: byte(len(cpwd)), AttrStr: cpwd},
	}
	msg.AuthBy(secret)
	return msg
}

func (v *Version) NewReqInfo(userip net.IP, secret string) portal.Message {
	msg := newMessage(portal.REQ_INFO, userip, portal.NewSerialNo(), 0)
	msg.Header.AttrNum = 2
	msg.Attrs = []T_Attr{{AttrType: byte(6), AttrLen: 0}, {AttrType: byte(7), AttrLen: 0}}
	msg.AuthBy(secret)
	return msg
}

func (v *Version) IsResponse(mesg portal.Message) bool {
	switch mesg.Type() {
	case 2, 4, 6, 10:
		return true
	}
	return false
}

func (v *Version) Unmarshall(bts []byte) portal.Message {
	msg := new(T_Message)
	buf := bytes.NewBuffer(bts)
	// v2 fixed header is 16 bytes + 16-byte authenticator = 32 bytes.
	// Reject truncated packets instead of panicking on malformed lengths.
	if buf.Len() < 32 {
		return msg
	}
	var ipbts [4]byte
	binary.Read(buf, binary.BigEndian, &msg.Header.Version)
	binary.Read(buf, binary.BigEndian, &msg.Header.Type)
	binary.Read(buf, binary.BigEndian, &msg.Header.Pap)
	binary.Read(buf, binary.BigEndian, &msg.Header.Rsv)
	binary.Read(buf, binary.BigEndian, &msg.Header.SerialNo)
	binary.Read(buf, binary.BigEndian, &msg.Header.ReqIdentifier)
	for i := 0; i < 4; i++ {
		binary.Read(buf, binary.BigEndian, &ipbts[i])
	}
	msg.Header.UserIp = net.IPv4(ipbts[0], ipbts[1], ipbts[2], ipbts[3]).To4()
	binary.Read(buf, binary.BigEndian, &msg.Header.UserPort)
	binary.Read(buf, binary.BigEndian, &msg.Header.ErrCode)
	binary.Read(buf, binary.BigEndian, &msg.Header.AttrNum)
	var auth [16]byte
	binary.Read(buf, binary.BigEndian, &auth)
	msg.Header.Authenticator = auth[:]
	attrNum := int(msg.Header.AttrNum)
	if attrNum > 0 {
		msg.Attrs = make([]T_Attr, attrNum)
		for i := 0; i < attrNum; i++ {
			// Each attribute on the wire is: type(1) + length(1) + data(len-2).
			// Guard against AttrLen < 2 (would make a negative slice) and
			// against lengths exceeding the remaining buffer.
			if buf.Len() < 2 {
				break
			}
			attrType, _ := buf.ReadByte()
			attrLen, _ := buf.ReadByte()
			if attrLen < 2 {
				break
			}
			dataLen := int(attrLen) - 2
			if dataLen > buf.Len() {
				break
			}
			attr := &msg.Attrs[i]
			attr.AttrType = attrType
			// v2 keeps AttrLen as the data length (header excluded).
			attr.AttrLen = byte(dataLen)
			attr.AttrStr = make([]byte, dataLen)
			binary.Read(buf, binary.BigEndian, &attr.AttrStr)
		}
	}
	return msg
}

var expect chan *T_Message
var timeout = fmt.Errorf("请求超时")

func newMessage(typ byte, userip net.IP, serialNo uint16, reqId uint16) *T_Message {
	msg := new(T_Message)
	msg.Header.Version = 0x02
	msg.Header.Type = typ
	msg.Header.SerialNo = serialNo
	msg.Header.ReqIdentifier = reqId
	msg.Header.UserIp = userip.To4()
	msg.Header.Authenticator = make([]byte, 16)
	return msg
}

// func Challenge(userip net.IP, secret string, basip net.IP) (response *T_Message, err error) {
// 	expect = make(chan *T_Message)
// 	cha := NewChallenge(userip, secret)
// 	go func() {
// 		err := cha.SendTo(basip, 2000)
// 		fmt.Println(err)
// 	}()
// 	//send challenge
// 	//expect res
// 	select {
// 	case res := <-expect:
// 		return res, res.CheckFor(ACK_CHALLENGE, cha.Header.Authenticator, secret)
// 	case <-time.After(15 * time.Second):
// 		return nil, timeout
// 	}
// }

// func Auth(userip net.IP, secret string, basip net.IP, username string, userpwd string, req uint16) (err error) {
// 	var res *T_Message
// 	res, err = Challenge(userip, secret, basip)
// 	if err == nil {
// 		cha := NewAuth(userip, secret, []byte(username), []byte(userpwd), res.Header.ReqId)
// 		go cha.SendTo(basip, 2000)
// 		//send challenge
// 		//expect res
// 		select {
// 		case res := <-expect:
// 			return res.CheckFor(ACK_AUTH, cha.Header.Authenticator, secret)
// 		case <-time.After(15 * time.Second):
// 			return timeout
// 		}
// 	} else {
// 		return
// 	}
// }
