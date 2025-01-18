package process

import (
	"bytes"
	"encoding/json"

	"log"
	"net/http"

	"github.com/hyperledger/fabric/orderer/consensus/pocr/httpserver"
	"github.com/hyperledger/fabric/orderer/consensus/pocr/message"
)

func (n *Node) SendAllNodes(msg *message.Request) {
	content, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error to marshal json")
		return
	}
	for k, v := range n.table {
		if k == n.id {
			continue
		}
		go SendPost(content, v+httpserver.RequestEntry)
	}
	n.buffer.AppendToRequestQueue(msg)
	n.totalReceive = n.totalReceive + 1
	n.p1leaders2nodesSendNotify <- true
}

func (n *Node) BroadCast(content []byte, handle string) {
	for k, v := range n.table {
		// do not send self
		if k == n.id {
			continue
		}
		go SendPost(content, v+handle)
	}
}

func (n *Node) BroadCastToCommittee(content []byte, handle string, msgelectionterm int) {
	var v string
	for i := 0; i < len(n.committeeList[msgelectionterm]); i++ {
		if n.committeeList[msgelectionterm][i] == n.id {
			continue
		}
		v = n.table[n.committeeList[msgelectionterm][i]]
		go SendPost(content, v+handle)
	}
}

func SendPost(content []byte, url string) {
	buff := bytes.NewBuffer(content)
	if _, err := http.Post(url, "application/json", buff); err != nil {
		log.Printf("[Send] send to %s error: %s", url, err)
	}
}
