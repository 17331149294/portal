package radius

import (
	"log"
	"net"
)

const AUTH_PORT = 1812
const ACCOUNTING_PORT = 1813

type Server struct {
	addr    string
	secret  string
	service Service
}

type Service interface {
	Authenticate(request *Packet) (*Packet, error)
}

type PasswordService struct{}

func (p *PasswordService) Authenticate(request *Packet) (*Packet, error) {
	npac := request.Reply()
	npac.Code = AccessReject
	npac.AVPs = append(npac.AVPs, AVP{Type: ReplyMessage, Value: []byte("you dick!")})
	return npac, nil
}
func NewServer(addr string, secret string) *Server {
	return &Server{addr, secret, nil}
}

func (s *Server) RegisterService(handler Service) {
	s.service = handler
}

func (s *Server) ListenAndServe() error {
	addr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	var b [512]byte
	for {
		n, addr, err := conn.ReadFrom(b[:])
		if err != nil {
			log.Printf("Radius err %s\n", err)
		}

		p := b[:n]
		pac := &Packet{server: s, nas: addr}
		go func(pac *Packet) {
			err = pac.Decode(p)
			if err != nil {
				log.Printf("Radius err %s\n", err)
			}

			npac, err := s.service.Authenticate(pac)
			if err != nil {
				log.Printf("Radius err %s\n", err)
			}
			err = npac.Send(conn, addr)
			if err != nil {
				log.Printf("Radius err %s\n", err)
			}
		}(pac)
	}
	return nil
}
