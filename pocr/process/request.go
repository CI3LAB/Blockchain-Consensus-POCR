package process

func (n *Node) requestRecvThread() {
	for {
		msg := <-n.requestRecv
		n.buffer.AppendToRequestQueue(msg)
		n.totalReceive = n.totalReceive + 1
		n.p1leaders2nodesSendNotify <- true
	}
}
