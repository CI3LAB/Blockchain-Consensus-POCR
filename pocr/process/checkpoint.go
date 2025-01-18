package process

func (n *Node) checkPointRecvThread() {
	for {
		select {
		case msg := <-n.checkPointRecv:
			n.buffer.BufferCheckPointMsg(msg, msg.Id)
			if n.buffer.IsTrueOfCheckPointMsg(msg.Digest, n.cfg.FaultNum) {
				n.buffer.Show()
				n.buffer.ClearBuffer(msg)
				n.sequence.CheckPoint()
				n.buffer.Show()
			}
		}
	}
}
