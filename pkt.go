package telehash

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
)

type pkt_t struct {
	addr *net.UDPAddr
	hdr  pkt_hdr_t
	body []byte
	peer Hashname
}

type pkt_hdr_t struct {
	Type   string          `json:"type,omitempty"`
	Line   string          `json:"line,omitempty"`
	Iv     string          `json:"iv,omitempty"`
	Open   string          `json:"open,omitempty"`
	Sig    string          `json:"sig,omitempty"`
	C      string          `json:"c,omitempty"`
	To     string          `json:"to,omitempty"`
	At     int64           `json:"at,omitempty"`
	Family string          `json:"family,omitempty"`
	Seq    int             `json:"seq,omitempty"`
	Ack    *int            `json:"ack,omitempty"`
	Miss   []int           `json:"miss,omitempty"`
	End    bool            `json:"end,omitempty"`
	Err    string          `json:"err,omitempty"`
	Seek   string          `json:"seek,omitempty"`
	See    []string        `json:"see,omitempty"`
	Peer   string          `json:"peer,omitempty"`
	IP     string          `json:"ip,omitempty"`
	Port   int             `json:"port,omitempty"`
	Custom json.RawMessage `json:"_,omitempty"`
}

func (p *pkt_t) JustAck() bool {
	return len(p.body) == 0 && p.hdr.JustAck()
}

func (p *pkt_hdr_t) JustAck() bool {
	if p.Type == "" &&
		p.Line == "" &&
		p.Iv == "" &&
		p.Open == "" &&
		p.Sig == "" &&
		p.To == "" &&
		p.At == 0 &&
		p.Family == "" &&
		p.Seq == 0 &&
		p.Ack != nil &&
		p.End == false &&
		p.Err == "" &&
		p.Seek == "" &&
		p.See == nil &&
		p.Peer == "" &&
		p.IP == "" &&
		p.Port == 0 &&
		len(p.Custom) == 0 {
		return true
	}
	return false
}

func (p *pkt_t) format_pkt() ([]byte, error) {
	if p == nil {
		panic("p cannot be nil")
	}

	var (
		data = make([]byte, 0, 1500)
		buf  = bytes.NewBuffer(data)
		err  error
		l    int
	)

	// make room for the length
	buf.WriteByte(0)
	buf.WriteByte(0)

	// write the header
	err = json.NewEncoder(buf).Encode(p.hdr)
	if err != nil {
		return nil, err
	}

	// get the header length
	l = buf.Len() - 2

	// write the body
	if len(p.body) > 0 {
		buf.Write(p.body)
	}

	// get the packet
	data = buf.Bytes()

	// put the header length
	binary.BigEndian.PutUint16(data[0:2], uint16(l))

	return data, nil
}

func parse_pkt(in []byte, addr *net.UDPAddr) (*pkt_t, error) {
	var (
		hdr_len  int
		body_len int
		err      error
		body     []byte
		pkt      = &pkt_t{addr: addr}
	)

	if len(in) < 4 {
		return nil, fmt.Errorf("pkt is too short")
	}

	// determin header length
	hdr_len = int(binary.BigEndian.Uint16(in[:2]))

	// determin body length
	body_len = len(in) - (hdr_len + 2)
	// Log.Debugf("pkt-len=%d hdr-len=%d body-len=%d", len(in), binary.BigEndian.Uint16(in[:2]), body_len)
	if body_len < 0 {
		return nil, fmt.Errorf("pkt is too short")
	}

	// decode the header
	err = json.NewDecoder(bytes.NewReader(in[2 : hdr_len+2])).Decode(&pkt.hdr)
	if err != nil {
		return nil, err
	}

	// no body
	if body_len == 0 {
		return pkt, nil
	}

	// copy the body
	body = make([]byte, body_len)
	copy(body, in[hdr_len+2:])
	pkt.body = body

	return pkt, nil
}
