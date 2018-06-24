package simple_miner

import (
	"fmt"
)

type MsgType int

const (
	Join MsgType = iota // miners join
	Request  // clients request servers, servers request miners
	Result  // miners send results to servers, servers send results to client
)

type Message struct {
	Type MsgType
	Data string
	Lower, Upper uint64
	Hash, Nonce uint64
}

func NewRequest(data string, lower uint64, upper uint64) *Message {
	return &Message{
		Type: Request,
		Data: data,
		Lower: lower,
		Upper: upper,
	}
}

func NewJoin() *Message {
	return &Message{
		Type: Join,
	}
}

func NewResult(hash, nonce uint64) *Message {
	return &Message{
		Type: Result,
		Hash: hash,
		Nonce: nonce,
	}
}

func (m *Message) String() string {
	var result string
	switch m.Type {
	case Request:
		result = fmt.Sprintf("[%s %s %d %d]", "Request", m.Data, m.Lower, m.Upper)
	case Result:
		result = fmt.Sprintf("[%s %d %d]", "Result", m.Hash, m.Nonce)
	case Join:
		result = fmt.Sprintf("[%s]", "Join")
	}
	return result
}
