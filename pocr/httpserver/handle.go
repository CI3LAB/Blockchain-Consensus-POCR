package httpserver

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/hyperledger/fabric/orderer/consensus/pocr/message"
)

func (s *HttpServer) HttpRequest(w http.ResponseWriter, r *http.Request) {
	var msg message.Request
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error] %s", err)
		return
	}
	s.requestRecv <- &msg
}

func (s *HttpServer) HttpP1leaders2nodes(w http.ResponseWriter, r *http.Request) {
	var msg message.P1leaders2nodes
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error] %s", err)
		return
	}
	s.p1leaders2nodesRecv <- &msg
}

func (s *HttpServer) HttpPrepare(w http.ResponseWriter, r *http.Request) {
	var msg message.Prepare
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error] %s", err)
		return
	}
	s.prepareRecv <- &msg
}

func (s *HttpServer) HttpP2verifiers2nodes(w http.ResponseWriter, r *http.Request) {
	var msg message.P2verifiers2nodes
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error] %s", err)
		return
	}
	s.p2verifiers2nodesRecv <- &msg
}

func (s *HttpServer) HttpCheckPoint(w http.ResponseWriter, r *http.Request) {
	var msg message.CheckPoint
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error] %s", err)
		return
	}
	s.checkPointRecv <- &msg
}
