package process

import (
	"log"
	"time"

	"github.com/hyperledger/fabric/orderer/consensus/pocr/httpserver"
	"github.com/hyperledger/fabric/orderer/consensus/pocr/message"
)

func (n *Node) p1leaders2nodesSendThread() {
	duration := time.Second
	timer := time.After(duration)
	for {
		select {
		// recv request or time out
		case <-n.p1leaders2nodesSendNotify:
			n.p1leaders2nodesSendHandleFunc()
		case <-timer:
			timer = nil
			n.p1leaders2nodesSendHandleFunc()
			timer = time.After(duration)
		}
	}
}

func (n *Node) p1leaders2nodesSendHandleFunc() {
	// buffer is empty or execute op num max
	n.executeNum.Lock()
	defer n.executeNum.UnLock()
	if n.executeNum.Get() >= n.cfg.ExecuteMaxNum {
		return
	}
	if n.id != message.Identify(n.view) {
		return
	}
	if n.buffer.SizeofRequestQueue() < 1 {
		return
	}
	batch, newTimeStampList, deletelength := n.buffer.BatchRequest(n.timeStampList)

	if len(batch) < 1 {
		return
	}
	n.executeNum.Inc()
	n.timeStampList = newTimeStampList
	n.totalDelete = n.totalDelete + deletelength
	seq := n.sequence.Get()
	content, msg, _, err := message.NewP1leaders2nodesMsg(n.view, n.id, seq, batch, n.jValue, n.proofHash, n.vrfHash)
	if err != nil {
		return
	}
	if n.electionTerm > 1 {
		n.buffer.UpdateTempLeader(msg, n.pubKeyTable[n.id], n.lastDigest)
	}
	n.buffer.BufferPreprepareMsg(msg)
	n.BroadCast(content, httpserver.P1leaders2nodesEntry)
	if n.electionTerm > 1 {
		if n.buffer.IsTrueOfPreprepareMsg(int(msg.Sequence), n.groupNum) {
			leaderMsg := n.buffer.FetchTempLeader(int(msg.Sequence))
			_, prepare, _ := message.NewPrepareMsg(n.id, leaderMsg, n.jValue, n.proofHash, n.vrfHash)
			content, p2verifiers2nodesmsg, _ := message.NewP2verifiers2nodesMsg(n.id, prepare)
			n.buffer.BufferP2verifiers2nodesMsg(p2verifiers2nodesmsg)
			n.BroadCast(content, httpserver.P2verifiers2nodesEntry)
			if n.buffer.IsReadyToExecute(leaderMsg.Digest, n.cfg.CommitteeFaultNum, leaderMsg.View, leaderMsg.Sequence) {
				n.readytoExecute(leaderMsg.Digest)
			}
		}
	}
}

func (n *Node) p1leaders2nodesRecvAndPrepareSendThread() {
	for {
		select {
		case msg := <-n.p1leaders2nodesRecv:
			if n.electionTerm > int(msg.Sequence) {
				continue
			}
			if n.electionTerm < int(msg.Sequence) {
				n.buffer.BufferPreprepareTempMsg(msg)
				continue
			}
			if !n.checkP1leaders2nodesMsg(msg) {
				continue
			}
			if int(msg.Sequence) > 1 {
				n.buffer.UpdateTempLeader(msg, n.pubKeyTable[msg.Identify], n.lastDigest)
			}
			n.buffer.BufferPreprepareMsg(msg)
			n.appendToCommitteeList(int(msg.Sequence), message.Identify(msg.View))
			if int(msg.Sequence) == 1 { // first round, tx conduct PBFT (leader election begins after frist round)
				content, prepare, _ := message.NewPrepareMsg(n.id, msg, n.jValue, n.proofHash, n.vrfHash)
				n.buffer.BufferPrepareMsg(prepare)
				n.BroadCast(content, httpserver.PrepareEntry)
				if n.buffer.IsReadyToExecute(msg.Digest, n.cfg.FaultNum, msg.View, msg.Sequence) {
					n.readytoExecute(msg.Digest)
				}
			}
			if int(msg.Sequence) > 1 {
				if n.buffer.IsTrueOfPreprepareMsg(int(msg.Sequence), n.groupNum) {
					leaderMsg := n.buffer.FetchTempLeader(int(msg.Sequence))
					if !n.isPotentialLeader(int(msg.Sequence)) {
						if n.buffer.IsReadyToExecute(leaderMsg.Digest, n.cfg.CommitteeFaultNum, leaderMsg.View, leaderMsg.Sequence) {
							n.readytoExecute(leaderMsg.Digest)
						}
						continue
					}
					_, prepare, _ := message.NewPrepareMsg(n.id, leaderMsg, n.jValue, n.proofHash, n.vrfHash)
					content, p2verifiers2nodesmsg, _ := message.NewP2verifiers2nodesMsg(n.id, prepare)
					n.buffer.BufferP2verifiers2nodesMsg(p2verifiers2nodesmsg)
					n.BroadCast(content, httpserver.P2verifiers2nodesEntry)
					if n.buffer.IsReadyToExecute(leaderMsg.Digest, n.cfg.CommitteeFaultNum, leaderMsg.View, leaderMsg.Sequence) {
						n.readytoExecute(leaderMsg.Digest)
					}
				}
			}
		}
	}
}

func (n *Node) processBufferP1leaders2nodesMsgTemp() {
	if !n.buffer.IsExistPreprepareTempMsg(n.electionTerm) {
		return
	}
	P1List := n.buffer.FetchP1leaders2nodesTempMsg(n.electionTerm)
	for _, v := range P1List {
		msg := v
		if !n.checkP1leaders2nodesMsg(msg) {
			break
		}
		n.buffer.UpdateTempLeader(msg, n.pubKeyTable[msg.Identify], n.lastDigest)
		n.buffer.BufferPreprepareMsg(msg)
		n.appendToCommitteeList(int(msg.Sequence), message.Identify(msg.View))
		n.buffer.ClearPreprepareTempMsg(int(msg.Sequence), msg.Identify)
		if n.buffer.IsTrueOfPreprepareMsg(int(msg.Sequence), n.groupNum) {
			leaderMsg := n.buffer.FetchTempLeader(int(msg.Sequence))
			if !n.isPotentialLeader(int(msg.Sequence)) {
				if n.buffer.IsReadyToExecute(leaderMsg.Digest, n.cfg.CommitteeFaultNum, leaderMsg.View, leaderMsg.Sequence) {
					n.readytoExecute(leaderMsg.Digest)
				}
				continue
			}
			_, prepare, _ := message.NewPrepareMsg(n.id, leaderMsg, n.jValue, n.proofHash, n.vrfHash)
			content, p2verifiers2nodesmsg, _ := message.NewP2verifiers2nodesMsg(n.id, prepare)
			n.buffer.BufferP2verifiers2nodesMsg(p2verifiers2nodesmsg)
			n.BroadCast(content, httpserver.P2verifiers2nodesEntry)
			if n.buffer.IsReadyToExecute(leaderMsg.Digest, n.cfg.CommitteeFaultNum, leaderMsg.View, leaderMsg.Sequence) {
				n.readytoExecute(leaderMsg.Digest)
			}
		}
	}
}

func (n *Node) checkP1leaders2nodesMsg(msg *message.P1leaders2nodes) bool {
	if n.buffer.IsExistPreprepareMsg(msg.View, msg.Sequence) {
		return false
	}
	messageAndView := message.MessageAndView{Requests: msg.Message.Requests, View: msg.View}
	d, err := message.Digest(messageAndView)
	if err != nil {
		log.Printf("[Pre-Prepare] False: digest is false")
		return false
	}
	if d != msg.Digest {
		return false
	}
	if !n.sequence.CheckBound(msg.Sequence) {
		return false
	}
	return true
}

func (n *Node) appendToCommitteeList_orginal(msg *message.P1leaders2nodes) {
	n.committeeListLocker.Lock()
	defer n.committeeListLocker.Unlock()
	if _, ok := n.committeeList[int(msg.Sequence)]; !ok {
		n.committeeList[int(msg.Sequence)] = []message.Identify{message.Identify(msg.View)}
	} else {
		n.committeeList[int(msg.Sequence)] = append(n.committeeList[int(msg.Sequence)], message.Identify(msg.View))
	}
}

func (n *Node) appendToCommitteeList(seqInt int, viewID message.Identify) {
	n.committeeListLocker.Lock()
	defer n.committeeListLocker.Unlock()
	if _, ok := n.committeeList[seqInt]; !ok {
		n.committeeList[seqInt] = []message.Identify{viewID}
	} else {
		n.committeeList[seqInt] = append(n.committeeList[seqInt], viewID)
	}
}

func (n *Node) isPotentialLeader(msgelectionterm int) bool {
	n.committeeListLocker.Lock()
	defer n.committeeListLocker.Unlock()
	var isPotentialLeader bool
	isPotentialLeader = false
	for i := 0; i < len(n.committeeList[msgelectionterm]); i++ {
		if n.committeeList[msgelectionterm][i] == n.id {
			isPotentialLeader = true
			break
		}
	}
	return isPotentialLeader
}
