package telehash

import (
	"crypto/rsa"
	"encoding/json"
)

type Switch struct {
	net      *net_controller
	channels *channel_controller
	peers    *peer_controller
	addr     string
	key      *rsa.PrivateKey
	mux      *SwitchMux
}

type Channel struct {
	c *channel_t
}

func NewSwitch(addr string, key *rsa.PrivateKey, handler Handler) (*Switch, error) {
	mux := NewSwitchMux()

	mux.HandleFallback(handler)

	s := &Switch{
		addr: addr,
		key:  key,
		mux:  mux,
	}

	return s, nil
}

func (s *Switch) Start() error {
	net, err := net_controller_open(s.addr)
	if err != nil {
		return err
	}
	s.net = net

	peers, err := peer_controller_open(s.key, s.mux)
	if err != nil {
		return err
	}
	s.peers = peers

	lines, err := line_controller_open(s)
	if err != nil {
		return nil, err
	}
	s.lines = lines

	conn, err := channel_controller_open(s.addr, s.key, s.mux, s.peers)
	if err != nil {
		return err
	}

	peers.conn = conn
	s.conn = conn
	return nil
}

func (s *Switch) Stop() error {
	s.conn.close()
	return nil
}

func (s *Switch) LocalHashname() Hashname {
	return s.peers.get_local_hashname()
}

func (s *Switch) Seed(addr string, key *rsa.PublicKey) (Hashname, error) {
	hashname, err := s.peers.add_peer(ZeroHashname, addr, key, ZeroHashname)
	if err != nil {
		return ZeroHashname, err
	}

	s.Seek(hashname, 15)
	return hashname, nil
}

func (s *Switch) Seek(hashname Hashname, n int) []Hashname {
	peers := s.peers.seek(hashname, n)
	hashnames := make([]Hashname, len(peers))

	for i, peer := range peers {
		hashnames[i] = peer.hashname
	}

	return hashnames
}

func (s *Switch) Open(hashname Hashname, typ string) (*Channel, error) {
	channel, err := s.conn.open_channel(hashname, &pkt_t{
		hdr: pkt_hdr_t{Type: typ},
	})

	if err != nil {
		return nil, err
	}

	return &Channel{channel}, nil
}

func (c *Channel) Close() error {
	return c.c.close()
}

func (c *Channel) Send(hdr interface{}, body []byte) error {
	pkt := &pkt_t{}

	if hdr != nil {
		custom, err := json.Marshal(hdr)
		if err != nil {
			return err
		}
		pkt.hdr.Custom = json.RawMessage(custom)
	}

	pkt.body = body

	return c.c.send(pkt)
}

func (c *Channel) Receive(hdr interface{}) (body []byte, err error) {
	pkt, err := c.c.receive()
	if err != nil {
		return nil, err
	}

	if len(pkt.hdr.Custom) > 0 {
		err = json.Unmarshal([]byte(pkt.hdr.Custom), hdr)
		if err != nil {
			return nil, err
		}
	}

	return pkt.body, nil
}
