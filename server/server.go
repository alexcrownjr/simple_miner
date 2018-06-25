package main 

import (
	"container/list"
	"os"
	"fmt"
	"simple_miner"
	// "strings"
	// "strconv"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)
var upgrader = websocket.Upgrader{}

type Server struct {
	clients map[*websocket.Conn]chan *simple_miner.Message
	miners map[*websocket.Conn]*ServerMessage

	incoming chan *ServerMessage
	requests chan *ServerMessage
	results chan *ServerMessage
	disconnect chan *websocket.Conn	

	maxJobSize uint64
	pendingRequests *list.List
	availableMiners *list.List
}	

type ServerMessage struct {
	conn *websocket.Conn  // we should now where it is from
	message *simple_miner.Message
}

func NewServerMessage(conn *websocket.Conn, message *simple_miner.Message) *ServerMessage{
	return &ServerMessage{conn, message}
}

func NewServer(port string, maxJobSize uint64) (*Server) {
	return &Server{
		pendingRequests: list.New(),
		availableMiners: list.New(),
		clients: make(map[*websocket.Conn]chan *simple_miner.Message),
		miners: make(map[*websocket.Conn]*ServerMessage),

		incoming: make(chan *ServerMessage),
		requests: make(chan *ServerMessage),
		results: make(chan *ServerMessage),
		disconnect: make(chan *websocket.Conn),
		maxJobSize: maxJobSize,
	}
}

// receive 持续监听,将message放入channel里,异步处理
func receive(w http.ResponseWriter, r *http.Request, s *Server) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return 
	}
	defer c.Close()
	for {
		var message simple_miner.Message
		err := c.ReadJSON(&message)
		if err != nil {
			s.disconnect <- c
			break
		}
		s.incoming <- NewServerMessage(c, &message)
	}
} 

// WriteMsg 在对应连接上发送消息
func (s *Server) WriteMsg(conn *websocket.Conn, m *simple_miner.Message) error {
	err := conn.WriteJSON(m)
	if err != nil {
		log.Println(err)
		conn.Close()
		return err
	}
	return nil
}

func main() {
	const numArgs = 2
	if len(os.Args) != numArgs {
		fmt.Println("Usage: ./server <port>")
		return
	}
	port := os.Args[1]
	
	s := NewServer(port, 10000)
	fmt.Printf("server start, listen on %s...\n", port);

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			receive(w, r, s)
		})
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}()

	for {
		select { // 多路复用,select只接受IO操作 
		case sm := <- s.incoming:
			s.handleIncoming(sm)  // handle client request, miner result
		case sm := <- s.requests:
			s.handleRequests(sm)  // request也是异步来做的,先发到channel里然后再处理
		case sm := <- s.results:
			s.handleResults(sm)  // 回发结果给client的时候也是异步来做的,先发到channel里
		case conn := <-s.disconnect:
			s.handleDisconnect(conn)
		}
	}
}



