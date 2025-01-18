package message

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"
)

type Buffer struct {
	requestQueue  []*Request
	requestLocker *sync.RWMutex

	tempLeader       map[int]*P1leaders2nodes
	tempLeaderLocker *sync.RWMutex

	p1leaders2nodesBuffer      map[string]*P1leaders2nodes
	p1leaders2nodesTempBuffer  map[int]map[Identify]*P1leaders2nodes
	p1leaders2nodesLeaderSet   map[int]map[Identify]bool
	p1leaders2nodesLeaderState map[int]bool
	p1leaders2nodesLocker      *sync.RWMutex

	prepareSet        map[string]map[Identify]bool
	prepareState      map[string]bool
	prepareTempBuffer map[int]map[Identify]*Prepare
	prepareLocker     *sync.RWMutex

	p2verifiers2nodesSet        map[string]map[Identify]bool
	p2verifiers2nodesState      map[string]bool
	p2verifiers2nodesTempBuffer map[int]map[Identify]*P2verifiers2nodes
	p2verifiers2nodesLocker     *sync.RWMutex

	executeQueue         []*P1leaders2nodes
	executeLocker        *sync.RWMutex
	readyToexecuteLocker *sync.RWMutex

	checkPointBuffer map[string]map[Identify]bool
	checkPointState  map[string]bool
	checkPointLocker *sync.RWMutex
}

func NewBuffer() *Buffer {
	return &Buffer{
		requestQueue:  make([]*Request, 0),
		requestLocker: new(sync.RWMutex),

		tempLeader:       make(map[int]*P1leaders2nodes),
		tempLeaderLocker: new(sync.RWMutex),

		p1leaders2nodesBuffer:      make(map[string]*P1leaders2nodes),
		p1leaders2nodesTempBuffer:  make(map[int]map[Identify]*P1leaders2nodes),
		p1leaders2nodesLeaderSet:   make(map[int]map[Identify]bool),
		p1leaders2nodesLeaderState: make(map[int]bool),
		p1leaders2nodesLocker:      new(sync.RWMutex),

		prepareSet:        make(map[string]map[Identify]bool),
		prepareState:      make(map[string]bool),
		prepareTempBuffer: make(map[int]map[Identify]*Prepare),
		prepareLocker:     new(sync.RWMutex),

		p2verifiers2nodesSet:        make(map[string]map[Identify]bool),
		p2verifiers2nodesState:      make(map[string]bool),
		p2verifiers2nodesTempBuffer: make(map[int]map[Identify]*P2verifiers2nodes),
		p2verifiers2nodesLocker:     new(sync.RWMutex),

		executeQueue:         make([]*P1leaders2nodes, 0),
		executeLocker:        new(sync.RWMutex),
		readyToexecuteLocker: new(sync.RWMutex),

		checkPointBuffer: make(map[string]map[Identify]bool),
		checkPointState:  make(map[string]bool),
		checkPointLocker: new(sync.RWMutex),
	}
}

func (b *Buffer) UpdateTempLeader(msg *P1leaders2nodes, pk []byte, lastDigest string) {
	b.tempLeaderLocker.Lock()
	if int(msg.Sequence) == 1 {
		return
	}
	lastDigestBytes, _ := hex.DecodeString(lastDigest)
	if _, ok := b.tempLeader[int(msg.Sequence)]; !ok {
		boolValue, _ := VerifyVRF(pk, msg.Proof, lastDigestBytes)
		if boolValue {
			b.tempLeader[int(msg.Sequence)] = msg
		}
	} else {
		temp := b.tempLeader[int(msg.Sequence)]
		if (msg.JValue > temp.JValue) || (msg.JValue == temp.JValue && temp.Identify > msg.Identify) {
			boolValue, _ := VerifyVRF(pk, msg.Proof, lastDigestBytes)
			if boolValue {
				b.tempLeader[int(msg.Sequence)] = msg
			}
		}
	}
	b.tempLeaderLocker.Unlock()
}

func (b *Buffer) FetchTempLeader(electionTerm int) *P1leaders2nodes {
	b.tempLeaderLocker.RLock()
	if _, ok := b.tempLeader[electionTerm]; !ok {
		return nil
	}
	ret := b.tempLeader[electionTerm]
	b.tempLeaderLocker.RUnlock()
	return ret
}

func (b *Buffer) UpdateTimeStampList(request []*Request, timeStampList []TimeStamp) []TimeStamp {
	for _, r := range request {
		timeStampList = append(timeStampList, r.TimeStamp)
	}
	return timeStampList
}

func (b *Buffer) ClearRequestQueue(timeStampList []TimeStamp) ([]TimeStamp, int) {
	var deletelength int
	deletelength = 0
	timeStampListTemp := make([]TimeStamp, 0)
	b.requestLocker.Lock()
	for i := 0; i < len(timeStampList); i++ {
		found := false
		for j := 0; j < len(b.requestQueue); j++ {
			if b.requestQueue[j].TimeStamp == timeStampList[i] {
				b.requestQueue = append(b.requestQueue[:j], b.requestQueue[j+1:]...)
				found = true
				deletelength++
				break
			}
		}
		if !found {
			timeStampListTemp = append(timeStampListTemp, timeStampList[i])
		}
	}
	b.requestLocker.Unlock()
	return timeStampListTemp, deletelength
}

func (b *Buffer) ClearRequestQueue_Original(Cleanlen int) {
	b.requestLocker.Lock()
	if len(b.requestQueue) <= Cleanlen {
		b.requestQueue = make([]*Request, 0)
	} else {
		b.requestQueue = b.requestQueue[Cleanlen:]
	}
	b.requestLocker.Unlock()
}

func (b *Buffer) BufferPreprepareTempMsg(msg *P1leaders2nodes) {
	b.p1leaders2nodesLocker.Lock()
	if b.p1leaders2nodesTempBuffer[int(msg.Sequence)] == nil {
		b.p1leaders2nodesTempBuffer[int(msg.Sequence)] = make(map[Identify]*P1leaders2nodes)
	}
	b.p1leaders2nodesTempBuffer[int(msg.Sequence)][msg.Identify] = msg
	b.p1leaders2nodesLocker.Unlock()
}

func (b *Buffer) ClearPreprepareTempMsg(electionTerm int, identify Identify) {
	b.p1leaders2nodesLocker.Lock()
	delete(b.p1leaders2nodesTempBuffer[electionTerm], identify)
	b.p1leaders2nodesLocker.Unlock()
}

func (b *Buffer) IsExistPreprepareTempMsg(electionTerm int) bool {
	b.p1leaders2nodesLocker.RLock()
	if _, ok := b.p1leaders2nodesTempBuffer[electionTerm]; ok {
		b.p1leaders2nodesLocker.RUnlock()
		return true
	}
	b.p1leaders2nodesLocker.RUnlock()
	return false
}

func (b *Buffer) FetchP1leaders2nodesTempMsg(electionTerm int) (ret map[Identify]*P1leaders2nodes) {
	ret = nil
	b.p1leaders2nodesLocker.RLock()
	if _, ok := b.p1leaders2nodesTempBuffer[electionTerm]; !ok {
		return
	}
	ret = b.p1leaders2nodesTempBuffer[electionTerm]
	b.p1leaders2nodesLocker.RUnlock()
	return
}

func (b *Buffer) BufferPrepareTempMsg(msg *Prepare) {
	b.prepareLocker.Lock()
	if b.prepareTempBuffer[int(msg.Sequence)] == nil {
		b.prepareTempBuffer[int(msg.Sequence)] = make(map[Identify]*Prepare)
	}
	b.prepareTempBuffer[int(msg.Sequence)][msg.Identify] = msg
	b.prepareLocker.Unlock()
}

func (b *Buffer) ClearPrepareTempMsg(electionTerm int, identify Identify) {
	b.prepareLocker.Lock()
	delete(b.prepareTempBuffer[electionTerm], identify)
	b.prepareLocker.Unlock()
}

func (b *Buffer) IsExistPrepareTempMsg(electionTerm int) bool {
	b.prepareLocker.RLock()
	if _, ok := b.prepareTempBuffer[electionTerm]; ok {
		b.prepareLocker.RUnlock()
		return true
	}
	b.prepareLocker.RUnlock()
	return false
}

func (b *Buffer) FetchPrepareTempMsg(electionTerm int) (ret map[Identify]*Prepare) {
	ret = nil
	b.prepareLocker.RLock()
	if _, ok := b.prepareTempBuffer[electionTerm]; !ok {
		return
	}
	ret = b.prepareTempBuffer[electionTerm]
	b.prepareLocker.RUnlock()
	return
}

func (b *Buffer) BufferP2verifiers2nodesTempMsg(msg *P2verifiers2nodes) {
	b.p2verifiers2nodesLocker.Lock()
	if b.p2verifiers2nodesTempBuffer[int(msg.Sequence)] == nil {
		b.p2verifiers2nodesTempBuffer[int(msg.Sequence)] = make(map[Identify]*P2verifiers2nodes)
	}
	b.p2verifiers2nodesTempBuffer[int(msg.Sequence)][msg.Identify] = msg
	b.p2verifiers2nodesLocker.Unlock()
}

func (b *Buffer) ClearP2verifiers2nodesTempMsg(electionTerm int, identify Identify) {
	b.p2verifiers2nodesLocker.Lock()
	delete(b.p2verifiers2nodesTempBuffer[electionTerm], identify)
	b.p2verifiers2nodesLocker.Unlock()
}

func (b *Buffer) IsExistP2verifiers2nodesTempMsg(electionTerm int) bool {
	b.p2verifiers2nodesLocker.RLock()
	if _, ok := b.p2verifiers2nodesTempBuffer[electionTerm]; ok {
		b.p2verifiers2nodesLocker.RUnlock()
		return true
	}
	b.p2verifiers2nodesLocker.RUnlock()
	return false
}

func (b *Buffer) FetchP2verifiers2nodesTempMsg(electionTerm int) (ret map[Identify]*P2verifiers2nodes) {
	ret = nil
	b.p2verifiers2nodesLocker.RLock()
	if _, ok := b.p2verifiers2nodesTempBuffer[electionTerm]; !ok {
		return
	}
	ret = b.p2verifiers2nodesTempBuffer[electionTerm]
	b.p2verifiers2nodesLocker.RUnlock()
	return
}

func (b *Buffer) AppendToRequestQueue(req *Request) {
	b.requestLocker.Lock()
	b.requestQueue = append(b.requestQueue, req)
	b.requestLocker.Unlock()
}

func (b *Buffer) BatchRequest(timeStampList []TimeStamp) ([]*Request, []TimeStamp, int) {
	var deletelength int
	deletelength = 0
	timeStampListTemp := make([]TimeStamp, 0)
	batch := make([]*Request, 0)
	b.requestLocker.Lock()
	for i := 0; i < len(timeStampList); i++ {
		found := false
		for j := 0; j < len(b.requestQueue); j++ {
			if b.requestQueue[j].TimeStamp == timeStampList[i] {
				b.requestQueue = append(b.requestQueue[:j], b.requestQueue[j+1:]...)
				found = true
				deletelength++
				break
			}
		}
		if !found {
			timeStampListTemp = append(timeStampListTemp, timeStampList[i])
		}
	}
	batch = make([]*Request, len(b.requestQueue))
	copy(batch, b.requestQueue)
	b.requestLocker.Unlock()
	return batch, timeStampListTemp, deletelength
}

func (b *Buffer) SizeofRequestQueue() (l int) {
	b.requestLocker.RLock()
	l = len(b.requestQueue)
	b.requestLocker.RUnlock()
	return
}

func (b *Buffer) BufferPreprepareMsg(msg *P1leaders2nodes) {
	b.p1leaders2nodesLocker.Lock()
	b.p1leaders2nodesBuffer[msg.Digest] = msg
	if _, ok := b.p1leaders2nodesLeaderSet[int(msg.Sequence)]; !ok {
		b.p1leaders2nodesLeaderSet[int(msg.Sequence)] = make(map[Identify]bool)
	}
	b.p1leaders2nodesLeaderSet[int(msg.Sequence)][msg.Identify] = true
	b.p1leaders2nodesLocker.Unlock()
}

func (b *Buffer) ClearPreprepareMsg(digest string) {
	b.p1leaders2nodesLocker.Lock()
	msg := b.p1leaders2nodesBuffer[digest]
	if msg == nil {
	} else {
	}
	delete(b.p1leaders2nodesBuffer, digest)
	delete(b.p1leaders2nodesLeaderSet, int(msg.Sequence))
	delete(b.p1leaders2nodesLeaderState, int(msg.Sequence))
	delete(b.tempLeader, int(msg.Sequence))
	b.p1leaders2nodesLocker.Unlock()
}

func (b *Buffer) IsTrueOfPreprepareMsg(election int, leaderNum int) bool {
	b.p1leaders2nodesLocker.Lock()
	num := len(b.p1leaders2nodesLeaderSet[election])
	_, ok := b.p1leaders2nodesLeaderState[election]
	if num < leaderNum || ok {
		b.p1leaders2nodesLocker.Unlock()
		return false
	}
	b.p1leaders2nodesLeaderState[election] = true
	b.p1leaders2nodesLocker.Unlock()
	return true
}

func (b *Buffer) IsExistPreprepareMsg(view View, seq Sequence) bool {
	b.p1leaders2nodesLocker.RLock()
	if _, ok := b.p1leaders2nodesLeaderSet[int(seq)][Identify(view)]; ok {
		b.p1leaders2nodesLocker.RUnlock()
		return true
	}
	b.p1leaders2nodesLocker.RUnlock()
	return false
}

func (b *Buffer) FetchPreprepareMsg(digest string) (ret *P1leaders2nodes) {
	ret = nil
	b.p1leaders2nodesLocker.RLock()
	if _, ok := b.p1leaders2nodesBuffer[digest]; !ok {
		return
	}
	ret = b.p1leaders2nodesBuffer[digest]
	b.p1leaders2nodesLocker.RUnlock()
	return
}

func (b *Buffer) BufferPrepareMsg(msg *Prepare) {
	b.prepareLocker.Lock()
	if _, ok := b.prepareSet[msg.Digest]; !ok {
		b.prepareSet[msg.Digest] = make(map[Identify]bool)
	}
	b.prepareSet[msg.Digest][msg.Identify] = true
	b.prepareLocker.Unlock()
}

func (b *Buffer) ClearPrepareMsg(digest string) {
	b.prepareLocker.Lock()
	delete(b.prepareSet, digest)
	delete(b.prepareState, digest)
	b.prepareLocker.Unlock()
}

func (b *Buffer) IsTrueOfPrepareMsg(digest string, falut uint) bool {
	b.prepareLocker.Lock()
	num := uint(len(b.prepareSet[digest]))
	_, ok := b.prepareState[digest]
	if num < 2*falut || ok {
		b.prepareLocker.Unlock()
		return false
	}
	b.prepareState[digest] = true
	b.prepareLocker.Unlock()
	return true
}

func (b *Buffer) IsExistOfDigest(digest string) bool {
	b.prepareLocker.Lock()
	_, ok := b.prepareState[digest]
	if !ok {
		return false
	}
	b.prepareLocker.Unlock()
	return true
}

func (b *Buffer) BufferP2verifiers2nodesMsg(msg *P2verifiers2nodes) {
	b.p2verifiers2nodesLocker.Lock()
	if _, ok := b.p2verifiers2nodesSet[msg.Digest]; !ok {
		b.p2verifiers2nodesSet[msg.Digest] = make(map[Identify]bool)
	}
	b.p2verifiers2nodesSet[msg.Digest][msg.Identify] = true
	b.p2verifiers2nodesLocker.Unlock()
}

func (b *Buffer) ClearP2verifiers2nodesMsg(digest string) {
	b.p2verifiers2nodesLocker.Lock()
	delete(b.p2verifiers2nodesSet, digest)
	delete(b.p2verifiers2nodesState, digest)
	b.p2verifiers2nodesLocker.Unlock()
}

func (b *Buffer) IsTrueOfP2verifiers2nodesMsg(digest string, falut uint) bool {
	b.p2verifiers2nodesLocker.Lock()
	num := uint(len(b.p2verifiers2nodesSet[digest]))
	_, ok := b.p2verifiers2nodesState[digest]
	if num < 2*falut+1 || ok {
		b.p2verifiers2nodesLocker.Unlock()
		return false
	}
	b.p2verifiers2nodesState[digest] = true
	b.p2verifiers2nodesLocker.Unlock()
	return true
}

func (b *Buffer) IsReadyToExecute_original(digest string, fault uint, view View, sequence Sequence) bool {
	b.p1leaders2nodesLocker.RLock()
	defer b.p1leaders2nodesLocker.RUnlock()
	if b.IsExistPreprepareMsg(view, sequence) && b.IsTrueOfP2verifiers2nodesMsg(digest, fault) {
		return true
	}
	return false
}

func (b *Buffer) IsReadyToExecute(digest string, fault uint, view View, sequence Sequence) bool {
	b.readyToexecuteLocker.RLock()
	defer b.readyToexecuteLocker.RUnlock()
	var isExitPrepreparemsg bool
	if _, ok := b.p1leaders2nodesLeaderSet[int(sequence)][Identify(view)]; ok {
		isExitPrepreparemsg = true
	} else {
		isExitPrepreparemsg = false
	}
	if isExitPrepreparemsg && b.IsTrueOfP2verifiers2nodesMsg(digest, fault) {
		return true
	}
	return false
}

func (b *Buffer) AppendToExecuteQueue(msg *P1leaders2nodes) {
	b.executeLocker.Lock()
	// upper bound first index greater than value
	count := len(b.executeQueue)
	first := 0
	for count > 0 {
		step := count / 2
		index := step + first
		if !(msg.Sequence < b.executeQueue[index].Sequence) {
			first = index + 1
			count = count - step - 1
		} else {
			count = step
		}
	}
	b.executeQueue = append(b.executeQueue, msg)
	copy(b.executeQueue[first+1:], b.executeQueue[first:])
	b.executeQueue[first] = msg
	b.executeLocker.Unlock()
}

func (b *Buffer) BatchExecute(lastSequence Sequence) ([]*P1leaders2nodes, Sequence) {
	b.executeLocker.Lock()
	batchs := make([]*P1leaders2nodes, 0)
	index := lastSequence
	loop := 0
	for {
		if loop == len(b.executeQueue) {
			b.executeQueue = make([]*P1leaders2nodes, 0)
			b.executeLocker.Unlock()
			return batchs, index
		}
		if b.executeQueue[loop].Sequence != index+1 {
			b.executeQueue = b.executeQueue[loop:]
			b.executeLocker.Unlock()
			return batchs, index
		}
		batchs = append(batchs, b.executeQueue[loop])
		loop = loop + 1
		index = index + 1
	}
}

func (b *Buffer) CheckPoint(sequence Sequence, id Identify) ([]byte, *CheckPoint) {
	clearSet := make(map[Sequence]string, 0)
	minSequence := sequence
	content := ""

	b.p1leaders2nodesLocker.RLock()
	for k, v := range b.p1leaders2nodesBuffer {
		if v.Sequence <= sequence {
			clearSet[v.Sequence] = k
			if v.Sequence < minSequence {
				minSequence = v.Sequence
			}
		}
	}
	b.p1leaders2nodesLocker.RUnlock()

	for minSequence <= sequence {
		d := clearSet[minSequence]
		content = content + d
		minSequence = minSequence + 1
	}

	msg := &CheckPoint{
		Sequence: sequence,
		Digest:   Hash([]byte(content)),
		Id:       id,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, nil
	}
	return data, msg
}

func (b *Buffer) ClearBuffer(msg *CheckPoint) {
	clearSet := make(map[Sequence]string, 0)
	minSequence := msg.Sequence

	b.p1leaders2nodesLocker.RLock()
	for k, v := range b.p1leaders2nodesBuffer {
		if v.Sequence <= msg.Sequence {
			clearSet[v.Sequence] = k
			if v.Sequence < minSequence {
				minSequence = v.Sequence
			}
		}
	}
	b.p1leaders2nodesLocker.RUnlock()

	for minSequence <= msg.Sequence {
		b.ClearPreprepareMsg(clearSet[minSequence])
		b.ClearPrepareMsg(clearSet[minSequence])
		b.ClearP2verifiers2nodesMsg(clearSet[minSequence])
		minSequence = minSequence + 1
	}
}

func (b *Buffer) BufferCheckPointMsg(msg *CheckPoint, id Identify) {
	b.checkPointLocker.Lock()
	if _, ok := b.checkPointBuffer[msg.Digest]; !ok {
		b.checkPointBuffer[msg.Digest] = make(map[Identify]bool)
	}
	b.checkPointBuffer[msg.Digest][id] = true
	b.checkPointLocker.Unlock()
}

func (b *Buffer) IsTrueOfCheckPointMsg(digest string, f uint) (ret bool) {
	ret = false
	b.checkPointLocker.RLock()
	num := uint(len(b.checkPointBuffer[digest]))
	_, ok := b.checkPointState[digest]
	if num < 2*f || ok {
		b.checkPointLocker.RUnlock()
		return
	}
	b.checkPointState[digest] = true
	ret = true
	b.checkPointLocker.RUnlock()
	return
}

func (b *Buffer) Show() {
	log.Printf("[Buffer] node buffer size: p1leaders2nodesBuffer(%d),p2verifiers2nodes(%d)",
		len(b.p1leaders2nodesBuffer), len(b.p2verifiers2nodesSet))
}
