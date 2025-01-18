package process

import (
	"github.com/hyperledger/fabric/orderer/consensus/pocr/message"
)

func (n *Node) p2verifiers2nodesRecvThread() {
	for {
		select {
		case msg := <-n.p2verifiers2nodesRecv:
			if n.electionTerm > int(msg.Sequence) {
				continue
			}
			if n.electionTerm < int(msg.Sequence) {
				n.buffer.BufferP2verifiers2nodesTempMsg(msg)
				continue
			}
			if !n.checkP2verifiers2nodesMsg(msg) {
				continue
			}
			n.buffer.BufferP2verifiers2nodesMsg(msg)
			if int(msg.Sequence) == 1 {
				if n.buffer.IsReadyToExecute(msg.Digest, n.cfg.FaultNum, msg.View, msg.Sequence) {
					n.readytoExecute(msg.Digest)
				}
			}
			if int(msg.Sequence) > 1 {
				if n.buffer.IsReadyToExecute(msg.Digest, n.cfg.CommitteeFaultNum, msg.View, msg.Sequence) {
					n.readytoExecute(msg.Digest)
				}
			}
		}
	}
}

func (n *Node) processBufferP2verifiers2nodesMsgTemp() {
	var msg *message.P2verifiers2nodes
	if !n.buffer.IsExistP2verifiers2nodesTempMsg(n.electionTerm) {
		return
	}
	P2List := n.buffer.FetchP2verifiers2nodesTempMsg(n.electionTerm)
	for _, v := range P2List {
		msg = v
		n.buffer.BufferP2verifiers2nodesMsg(msg)
		n.buffer.ClearP2verifiers2nodesTempMsg(int(msg.Sequence), msg.Identify)
		if n.buffer.IsReadyToExecute(msg.Digest, n.cfg.CommitteeFaultNum, msg.View, msg.Sequence) {
			n.readytoExecute(msg.Digest)
		}
	}
}

func (n *Node) checkP2verifiers2nodesMsg(msg *message.P2verifiers2nodes) bool {
	if !n.sequence.CheckBound(msg.Sequence) {
		return false
	}
	return true
}
