package process

import (
	"encoding/hex"

	"github.com/hyperledger/fabric/orderer/consensus/pocr/message"
)

func (n *Node) updateleader() {
	n.viewLocker.Lock()
	defer n.viewLocker.Unlock()
	n.lastView = n.view
	var j int
	leaderID := n.id
	leaderJ := 0
	for k := 0; k < len(n.groupList); k++ {
		digestBytes, _ := hex.DecodeString(n.lastDigest)
		proofhash, vrfhash, _ := message.GenerateProofAndHash(n.pubKeyTable[n.groupList[k]], n.priKeyTable[n.groupList[k]], digestBytes)
		hashratio := message.CalculateRatio(vrfhash)
		j = message.CalculateJ(n.crTable[n.groupList[k]], hashratio)
		if n.groupList[k] == n.id {
			n.jValue = j
			n.proofHash = proofhash
			n.vrfHash = vrfhash
		}
		if j > leaderJ || (j == leaderJ && leaderID > n.groupList[k]) {
			leaderID = n.groupList[k]
			leaderJ = j
		}
	}
	n.appendToCommitteeList(int(n.electionTerm+1), leaderID)
	n.view = message.View(leaderID)
	return
}

func (n *Node) updateCRCredit() {
	if n.crTable[message.Identify(n.finalLeader)] >= float64(10) {
		n.crTable[message.Identify(n.finalLeader)] = n.crTable[message.Identify(n.finalLeader)] - float64(2)
	}
}
