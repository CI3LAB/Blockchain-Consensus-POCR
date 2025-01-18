package message

import (
	"encoding/json"

	cb "github.com/hyperledger/fabric/protos/common"
)

type TimeStamp uint64
type Identify uint64
type View Identify
type Sequence int64

const TYPENORMAL = "normal"
const TYPECONFIG = "config"

type Operation struct {
	Envelope  *cb.Envelope
	ChannelID string
	ConfigSeq uint64
	Type      string
}

type Result struct {
}

type Request struct {
	Op        Operation `json:"operation"`
	TimeStamp TimeStamp `json:"timestamp"`
	ID        Identify  `json:"clientID"`
}

type Message struct {
	Requests []*Request `json:"requests"`
}

type MessageAndView struct {
	Requests []*Request `json:"requests"`
	View     View       `json:"view"`
}

type P1leaders2nodes struct {
	View     View     `json:"view"`
	Sequence Sequence `json:"sequence"`
	Digest   string   `json:"digest"`
	Message  Message  `json:"message"`
	Identify Identify `json:"id"`
	// VRF
	VRFHash []byte `json:"vrfhash"`
	Proof   []byte `json:"proofhash"`
	JValue  int    `json:"jvalue"`
}

type Prepare struct {
	View     View     `json:"view"`
	Sequence Sequence `json:"sequence"`
	Digest   string   `json:"digest"`
	Identify Identify `json:"id"`
	// VRF
	VRFHash []byte `json:"vrfhash"`
	Proof   []byte `json:"proofhash"`
	JValue  int    `json:"jvalue"`
}

type P2verifiers2nodes struct {
	View     View     `json:"view"`
	Sequence Sequence `json:"sequence"`
	Digest   string   `json:"digest"`
	Identify Identify `json:"id"`
}

type Reply struct {
	View      View      `json:"view"`
	TimeStamp TimeStamp `json:"timestamp"`
	Id        Identify  `json:"nodeID"`
	Result    Result    `json:"result"`
}

type CheckPoint struct {
	Sequence Sequence `json:"sequence"`
	Digest   string   `json:"digest"`
	Id       Identify `json:"nodeID"`
}

func NewP1leaders2nodesMsg(view View, id Identify, seq Sequence, batch []*Request, jValue int, proofHash []byte, vrfHash []byte) ([]byte, *P1leaders2nodes, string, error) {
	message := Message{Requests: batch}
	messageAndView := MessageAndView{Requests: batch, View: view}
	d, err := Digest(messageAndView)
	if err != nil {
		return []byte{}, nil, "", nil
	}
	//
	p1leaders2nodes := &P1leaders2nodes{
		View:     view,
		Sequence: seq,
		Digest:   d,
		Message:  message,
		Identify: id,
		// vrf
		VRFHash: vrfHash,
		Proof:   proofHash,
		JValue:  jValue,
	}
	ret, err := json.Marshal(p1leaders2nodes)
	if err != nil {
		return []byte{}, nil, "", nil
	}
	return ret, p1leaders2nodes, d, nil
}

func NewPrepareMsg(id Identify, msg *P1leaders2nodes, jValue int, proofHash []byte, vrfHash []byte) ([]byte, *Prepare, error) {
	prepare := &Prepare{
		View:     msg.View,
		Sequence: msg.Sequence,
		Digest:   msg.Digest,
		Identify: id,
		VRFHash:  vrfHash,
		Proof:    proofHash,
		JValue:   jValue,
	}
	content, err := json.Marshal(prepare)
	if err != nil {
		return []byte{}, nil, err
	}
	return content, prepare, nil
}

func NewP2verifiers2nodesMsg(id Identify, msg *Prepare) ([]byte, *P2verifiers2nodes, error) {
	p2verifiers2nodes := &P2verifiers2nodes{
		View:     msg.View,
		Sequence: msg.Sequence,
		Digest:   msg.Digest,
		Identify: id,
	}
	content, err := json.Marshal(p2verifiers2nodes)
	if err != nil {
		return []byte{}, nil, err
	}
	return content, p2verifiers2nodes, nil
}
