package main

import (
	"log"
	"simple_miner"
	"github.com/gorilla/websocket"
)

func min(x, y uint64) uint64 {
	if x <= y {
		return x
	}
	return y
}

func(s *Server) handleDisconnect(conn *websocket.Conn) {
	if ch, ok := s.clients[conn]; ok {
		delete(s.clients, conn)
		close(ch)
	} else if sm, ok := s.miners[conn]; ok {
		delete(s.miners, conn) 
		if sm != nil { // 有任务在身的话又得重新分配任务了
			go s.requestMiner(sm)
		}
	}
}

func (s *Server) handleRequests(sm *ServerMessage) {
	for s.availableMiners.Len() > 0 {
		conn := s.availableMiners.Remove(s.availableMiners.Front()).(*websocket.Conn)
		if _, ok := s.miners[conn]; ok {
			s.miners[conn] = sm
			s.WriteMsg(conn, sm.message)
			return
		}
	}
	s.pendingRequests.PushBack(sm) // 现在不能立即处理
}

func (s *Server) handleResults(sm *ServerMessage) {
	log.Println("sending the result!")
	s.WriteMsg(sm.conn, sm.message) 
	delete(s.clients, sm.conn)
	sm.conn.Close()
}

func (s *Server) handleIncoming(sm *ServerMessage) {
	switch sm.message.Type {
	case simple_miner.Request:
		log.Println("Get client request")
		s.handleClientRequests(sm)  // client发起挖矿请求
	case simple_miner.Result:
		s.handleMinerResult(sm)  // miner发送挖矿结果
	case simple_miner.Join:
		log.Println("New Miner is joining")
		s.handleJoin(sm)
	}
}

func (s *Server) handleClientRequests(message *ServerMessage) {
	res := make(chan *simple_miner.Message)
	s.clients[message.conn] = res
	go s.handleClient(message.conn, message.message.Data, message.message.Lower, message.message.Upper, res)
}

func (s *Server) handleClient(conn *websocket.Conn, data string, lower uint64, upper uint64, res chan *simple_miner.Message) {
	for nonce := lower; nonce <= upper; nonce += s.maxJobSize {
		go s.requestMiner(NewServerMessage(conn, simple_miner.NewRequest(data, nonce, min(nonce+s.maxJobSize, upper))))
	}
	var minHash, minNonce uint64
	for nonce := lower; nonce <= upper; nonce += s.maxJobSize {
		if m, ok := <-res; !ok {  // 会等待所有miner带回来的结果
			return 
		} else if nonce == lower || m.Hash < minHash {
			minHash = m.Hash
			minNonce = m.Nonce
		}
	}
	s.results <- NewServerMessage(conn, simple_miner.NewResult(minHash, minNonce))
}

func (s *Server) requestMiner(msg *ServerMessage) {
	s.requests <- msg
}


func (s *Server) handleMinerResult(sm *ServerMessage) {
	if result, ok := s.clients[s.miners[sm.conn].conn]; ok {
		result <- sm.message  // 得到结果
	}
	s.handleMinerAvailable(sm.conn) // 该miner又解放了
}

func (s *Server) handleJoin(sm *ServerMessage) { // 多了新的miner劳动力
	s.handleMinerAvailable(sm.conn)
}

func (s *Server) handleMinerAvailable(conn *websocket.Conn) {
	for s.pendingRequests.Len() > 0 {
		sm := s.pendingRequests.Remove(s.pendingRequests.Front()).(*ServerMessage)
		if _, ok := s.clients[sm.conn]; ok {
			s.miners[conn] = sm  // miner会把消息记录下来，到时候可以知道是哪个client发出的请求
			s.WriteMsg(conn, sm.message)
			return 
		}
	}
	s.miners[conn] = nil
	s.availableMiners.PushBack(conn)
}