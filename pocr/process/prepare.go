package process

import (
	"log"

	"github.com/hyperledger/fabric/orderer/consensus/pocr/httpserver"
	"github.com/hyperledger/fabric/orderer/consensus/pocr/message"
)

func (n *Node) prepareRecvAndP2verifiers2nodesSendThread() {
	for {
		select {
		case msg := <-n.prepareRecv:
			if n.electionTerm > int(msg.Sequence) {
				continue
			}
			if n.electionTerm < int(msg.Sequence) {
				n.buffer.BufferPrepareTempMsg(msg)
				continue
			}
			if !n.checkPrepareMsg(msg) {
				continue
			}
			n.buffer.BufferPrepareMsg(msg)
			if int(msg.Sequence) == 1 {
				if n.buffer.IsTrueOfPrepareMsg(msg.Digest, n.cfg.FaultNum) {
					content, msg, err := message.NewP2verifiers2nodesMsg(n.id, msg)
					if err != nil {
						continue
					}
					n.buffer.BufferP2verifiers2nodesMsg(msg)
					n.BroadCast(content, httpserver.P2verifiers2nodesEntry)
					if n.buffer.IsReadyToExecute(msg.Digest, n.cfg.FaultNum, msg.View, msg.Sequence) {
						n.readytoExecute(msg.Digest)
					}
				}
			}
			if int(msg.Sequence) > 1 {
				if n.buffer.IsTrueOfPrepareMsg(msg.Digest, n.cfg.CommitteeFaultNum) {
					content, msg, err := message.NewP2verifiers2nodesMsg(n.id, msg)
					if err != nil {
						continue
					}
					n.buffer.BufferP2verifiers2nodesMsg(msg)
					n.BroadCast(content, httpserver.P2verifiers2nodesEntry)
				} else {
					log.Printf("[Prepare] not satisfy the IsTrueOfPrepareMsg, sequence is (%d)", int(msg.Sequence))
				}
				if n.buffer.IsReadyToExecute(msg.Digest, n.cfg.CommitteeFaultNum, msg.View, msg.Sequence) {
					n.readytoExecute(msg.Digest)
				}
			}

		}
	}
}

func (n *Node) processBufferPrepareMsgTemp() {
	var msg *message.Prepare
	if !n.buffer.IsExistPrepareTempMsg(n.electionTerm) {
		return
	}
	P2List := n.buffer.FetchPrepareTempMsg(n.electionTerm)
	for _, v := range P2List {
		msg = v
		n.buffer.BufferPrepareMsg(msg)
		n.buffer.ClearPrepareTempMsg(n.electionTerm, msg.Identify)
		if n.buffer.IsTrueOfPrepareMsg(msg.Digest, n.cfg.CommitteeFaultNum) {
			content, msg, err := message.NewP2verifiers2nodesMsg(n.id, msg)
			if err != nil {
				continue
			}
			n.buffer.BufferP2verifiers2nodesMsg(msg)
			n.BroadCast(content, httpserver.P2verifiers2nodesEntry)
		}
		if n.buffer.IsReadyToExecute(msg.Digest, n.cfg.CommitteeFaultNum, msg.View, msg.Sequence) {
			n.readytoExecute(msg.Digest)
		}
	}
}

func (n *Node) checkPrepareMsg(msg *message.Prepare) bool {
	if !n.sequence.CheckBound(msg.Sequence) {
		return false
	}
	return true
}
