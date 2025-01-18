package process

import (
	"log"

	"github.com/hyperledger/fabric/orderer/consensus/pocr/message"
)

func (n *Node) readytoExecute(digest string) {
	msg := n.buffer.FetchPreprepareMsg(digest)
	n.buffer.AppendToExecuteQueue(n.buffer.FetchPreprepareMsg(digest))
	n.executeNotify <- true
	n.finalLeader = msg.View
	log.Printf("[Execute] the final leader is %d and the election is %d", int(n.finalLeader), n.electionTerm)
	n.timeStampList = n.buffer.UpdateTimeStampList(msg.Message.Requests, n.timeStampList)
	newtimeStampList, deletelength := n.buffer.ClearRequestQueue(n.timeStampList)
	n.timeStampList = newtimeStampList
	n.totalDelete = n.totalDelete + deletelength
	n.sequence.sequence = msg.Sequence
	n.lastDigest = msg.Digest
	// updateleader
	n.updateleader()
	n.electionTerm = n.electionTerm + 1

	if n.id == message.Identify(n.lastView) {
		n.executeNum.num = 0
	}
	// check bufferTemp
	n.processBufferP1leaders2nodesMsgTemp()
	n.processBufferP2verifiers2nodesMsgTemp()
}
