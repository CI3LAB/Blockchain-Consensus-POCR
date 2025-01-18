package process

import (
	"log"
	"sync"

	"github.com/hyperledger/fabric/orderer/consensus"
	"github.com/hyperledger/fabric/orderer/consensus/pocr/cmd"
	"github.com/hyperledger/fabric/orderer/consensus/pocr/httpserver"
	"github.com/hyperledger/fabric/orderer/consensus/pocr/message"
)

var GNode *Node = nil

type Node struct {
	cfg    *cmd.SharedConfig
	server *httpserver.HttpServer

	id            message.Identify
	view          message.View
	viewLocker    *sync.RWMutex
	lastView      message.View
	table         map[message.Identify]string
	faultNum      uint
	electionTerm  int
	totalReceive  int
	totalDelete   int
	timeStampList []message.TimeStamp
	// PoCr
	role        string
	priKeyTable map[message.Identify][]byte
	pubKeyTable map[message.Identify][]byte
	crTable     map[message.Identify]float64
	lastDigest  string
	// vrf (self)
	jValue      int
	proofHash   []byte
	vrfHash     []byte
	finalLeader message.View

	// group
	groupList           []message.Identify
	groupNum            int
	committeeList       map[int][]message.Identify
	committeeFaultNum   uint
	committeeListLocker *sync.RWMutex

	lastReply  *message.LastReply
	sequence   *Sequence
	executeNum *ExecuteOpNum

	buffer *message.Buffer

	requestRecv           chan *message.Request
	p1leaders2nodesRecv   chan *message.P1leaders2nodes
	prepareRecv           chan *message.Prepare
	p2verifiers2nodesRecv chan *message.P2verifiers2nodes
	checkPointRecv        chan *message.CheckPoint

	p1leaders2nodesSendNotify chan bool
	executeNotify             chan bool

	supports map[string]consensus.ConsenterSupport
}

func NewNode(cfg *cmd.SharedConfig, support consensus.ConsenterSupport) *Node {
	node := &Node{
		// config
		cfg: cfg,
		// http server
		server: httpserver.NewServer(cfg),
		// information about node
		id:            cfg.Id,
		view:          message.View(0), // initial setting
		viewLocker:    new(sync.RWMutex),
		lastView:      message.View(0),
		table:         cfg.Table,
		faultNum:      cfg.FaultNum,
		electionTerm:  1,
		totalReceive:  0,
		totalDelete:   0,
		timeStampList: make([]message.TimeStamp, 0),

		role:        "LeaderOrVerifier", // initial setting
		priKeyTable: cfg.PriKeyTable,
		pubKeyTable: cfg.PubKeyTable,
		crTable:     cfg.CRTable,
		lastDigest:  "hello", // initial setting
		// vrf
		jValue:      0,
		proofHash:   make([]byte, 64),
		vrfHash:     make([]byte, 0),
		finalLeader: message.View(0),
		// group
		groupList:           cfg.GroupList,
		groupNum:            cfg.GroupNum,
		committeeList:       make(map[int][]message.Identify),
		committeeFaultNum:   cfg.CommitteeFaultNum,
		committeeListLocker: new(sync.RWMutex),

		// lastReply state
		lastReply:  message.NewLastReply(),
		sequence:   NewSequence(cfg),
		executeNum: NewExecuteOpNum(),
		// the message buffer to store msg
		buffer: message.NewBuffer(),
		// chan for server and recv thread
		requestRecv:           make(chan *message.Request),
		p1leaders2nodesRecv:   make(chan *message.P1leaders2nodes),
		prepareRecv:           make(chan *message.Prepare),
		p2verifiers2nodesRecv: make(chan *message.P2verifiers2nodes),
		checkPointRecv:        make(chan *message.CheckPoint),

		p1leaders2nodesSendNotify: make(chan bool),
		executeNotify:             make(chan bool, 100),
		supports:                  make(map[string]consensus.ConsenterSupport),
	}
	node.RegisterChain(support)
	return node
}

func (n *Node) RegisterChain(support consensus.ConsenterSupport) {
	if _, ok := n.supports[support.ChainID()]; ok {
		return
	}
	log.Printf("[Node] Register the chain(%s)", support.ChainID())
	n.supports[support.ChainID()] = support
}

func (n *Node) Run() {
	// first register chan for server
	n.server.RegisterChan(n.requestRecv, n.p1leaders2nodesRecv, n.prepareRecv, n.p2verifiers2nodesRecv, n.checkPointRecv)
	go n.server.Run()
	go n.requestRecvThread()
	go n.p1leaders2nodesSendThread()
	go n.p1leaders2nodesRecvAndPrepareSendThread()
	go n.prepareRecvAndP2verifiers2nodesSendThread()
	go n.p2verifiers2nodesRecvThread()
	go n.executeAndReplyThread()
	go n.checkPointRecvThread()
}
