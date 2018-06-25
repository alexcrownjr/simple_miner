package main 

import (
	"log"
	"simple_miner"
	"os"
	"fmt"
	"github.com/gorilla/websocket"
)

type Miner struct {
	conn *websocket.Conn
	jobs chan *simple_miner.Message
	quit chan interface{}
}

func JoinWithServer(hostport string) (*Miner, error) {
	conn, _, err := websocket.DefaultDialer.Dial(hostport, nil)
	if err != nil {
		return nil, err
	}
	jobs := make(chan *simple_miner.Message)
	return &Miner{conn, jobs, make(chan interface{})}, nil
}

func (m *Miner) ReadMsg() {
	defer m.conn.Close()
	for {
		var message simple_miner.Message
		err := m.conn.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			close(m.quit)
			return 
		}
		if message.Type == simple_miner.Request {
			m.jobs <- &message
		}
	}
}

func (m *Miner) WriteMsg(msg *simple_miner.Message) error {
	err := m.conn.WriteJSON(msg)
	if err != nil {
		log.Println(err)
		m.conn.Close()
	}
	return err
}

func (m * Miner) Mining(data string, lower, upper uint64) (hash, nonce uint64) {
	var minHash, minNonce uint64 
	for nonce := lower; nonce <= upper; nonce++ {
		hash := simple_miner.Hash(data, nonce)
		if nonce == lower || hash < minHash {
			minHash, minNonce = hash, nonce
		}
	}
	return minHash, minNonce
}

func main() {
	const numArgs = 2
	if len(os.Args) != numArgs && len(os.Args) > 0 {
		fmt.Printf("Usage: ./%s <hostport>", os.Args[0])
		return
	}
	hostport := "ws://localhost:" + os.Args[1]
	miner, err := JoinWithServer(hostport) // 发起连接
	if err != nil {
		fmt.Println("Failed to join with server: ", err)
	}
	fmt.Println("miner start success!")
	
	err = miner.WriteMsg(simple_miner.NewJoin())
	if err != nil {
		log.Println("fail to join!")
		return
	}
	go miner.ReadMsg()  // get mining jobs

	for {
		select {
		case msg := <- miner.jobs: // waiting for things in channel 
			log.Printf("get job, Data: %s, Lower: %d, Upper: %d\n", msg.Data, msg.Lower, msg.Upper)
			hash, nonce := miner.Mining(msg.Data, msg.Lower, msg.Upper)
			miner.WriteMsg(simple_miner.NewResult(hash, nonce))
		case <- miner.quit:
			return 
		}
	}
	
}
