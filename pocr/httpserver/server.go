package httpserver

import (
	"log"
	"net/http"
	"strconv"

	"github.com/hyperledger/fabric/orderer/consensus/pocr/cmd"
	"github.com/hyperledger/fabric/orderer/consensus/pocr/message"
)

const (
	RequestEntry           = "/request"
	P1leaders2nodesEntry   = "/p1leaders2nodes"
	PrepareEntry           = "/prepare"
	P2verifiers2nodesEntry = "/p2verifiers2nodes"
	CheckPointEntry        = "/checkpoint"
)

type HttpServer struct {
	port   int
	server *http.Server

	requestRecv           chan *message.Request
	p1leaders2nodesRecv   chan *message.P1leaders2nodes
	prepareRecv           chan *message.Prepare
	p2verifiers2nodesRecv chan *message.P2verifiers2nodes
	checkPointRecv        chan *message.CheckPoint
}

func NewServer(cfg *cmd.SharedConfig) *HttpServer {
	httpServer := &HttpServer{
		port:   cfg.Port,
		server: nil,
	}
	return httpServer
}

func (s *HttpServer) RegisterChan(r chan *message.Request, pre chan *message.P1leaders2nodes,
	p chan *message.Prepare, c chan *message.P2verifiers2nodes, cp chan *message.CheckPoint) {
	log.Printf("[Server] register the chan for listen func")
	s.requestRecv = r
	s.p1leaders2nodesRecv = pre
	s.prepareRecv = p
	s.p2verifiers2nodesRecv = c
	s.checkPointRecv = cp
}

func (s *HttpServer) Run() {
	log.Printf("[Node] start the listen server")
	s.registerServer()
}

func (s *HttpServer) registerServer() {
	log.Printf("[Server] set listen port:%d\n", s.port)
	httpRegister := map[string]func(http.ResponseWriter, *http.Request){
		RequestEntry:           s.HttpRequest,
		P1leaders2nodesEntry:   s.HttpP1leaders2nodes,
		PrepareEntry:           s.HttpPrepare,
		P2verifiers2nodesEntry: s.HttpP2verifiers2nodes,
		CheckPointEntry:        s.HttpCheckPoint,
	}

	mux := http.NewServeMux()
	for k, v := range httpRegister {
		mux.HandleFunc(k, v)
	}

	s.server = &http.Server{
		Addr:    ":" + strconv.Itoa(s.port),
		Handler: mux,
	}

	if err := s.server.ListenAndServe(); err != nil {
		log.Printf("[Server Error] %s", err)
		return
	}
}
