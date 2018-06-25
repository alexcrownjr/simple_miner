package main 

import (
	"github.com/gorilla/websocket"
	"strconv"
	"os"
	"fmt"
	"log"
	"simple_miner"
)

// Client should know conn only
type Client struct {
	conn *websocket.Conn
}

// ClientJoin 发起连接 
func ClientJoin(hostport string) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(hostport, nil)
	if err != nil {
		return nil, err
	}
	return &Client{conn}, nil
}

func main() {
	const numArgs = 4
	if len(os.Args) != numArgs {
		fmt.Printf("Usage: ./%s <hostport>, <message>, <maxNonce>", os.Args[0])
		return
	}
	hostport := "ws://localhost:" + os.Args[1]
	message := os.Args[2]  
	maxNonce, err := strconv.ParseUint(os.Args[3], 10, 64)
	if err != nil {
		fmt.Printf("%s is not a number.\n", os.Args[3])
		return
	}
	client, err := ClientJoin(hostport)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("client start, connect to server!")

	err = client.conn.WriteJSON(simple_miner.NewRequest(message, 0, maxNonce)) //连上了以后发送request
	if err != nil {
		log.Println("we lost server")
		return
	}
	fmt.Println("send the request success!")
	
	var msg simple_miner.Message
	err = client.conn.ReadJSON(&msg)
	if err != nil {
		log.Println("we lost server")
		return
	}
	fmt.Println("Result ", strconv.FormatUint(msg.Hash, 10), strconv.FormatUint(msg.Nonce, 10))
}