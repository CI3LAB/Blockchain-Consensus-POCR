package pocr

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric/orderer/consensus"
	"github.com/hyperledger/fabric/orderer/consensus/pocr/cmd"
	"github.com/hyperledger/fabric/orderer/consensus/pocr/message"
	"github.com/hyperledger/fabric/orderer/consensus/pocr/process"

	cb "github.com/hyperledger/fabric/protos/common"
)

type Chain struct {
	exitChan chan struct{}
	support  consensus.ConsenterSupport
	pocrNode *process.Node
}

func NewChain(support consensus.ConsenterSupport) *Chain {
	logger.Info("NewChain - ", support.ChainID())
	if process.GNode == nil {
		process.GNode = process.NewNode(cmd.ReadConfig(), support)
		process.GNode.Run()
	} else {
		process.GNode.RegisterChain(support)
	}

	c := &Chain{
		exitChan: make(chan struct{}),
		support:  support,
		pocrNode: process.GNode,
	}
	return c
}

func (ch *Chain) Start() {
	logger.Info("start")
}

func (ch *Chain) Errored() <-chan struct{} {
	return ch.exitChan
}

func (ch *Chain) Halt() {
	logger.Info("halt")
	select {
	case <-ch.exitChan:
	default:
		close(ch.exitChan)
	}
}

func (ch *Chain) WaitReady() error {
	logger.Info("wait ready")
	return nil
}

func (ch *Chain) Order(env *cb.Envelope, configSeq uint64) error {
	logger.Info("Normal")
	select {
	case <-ch.exitChan:
		logger.Info("[CHAIN error exit normal]")
		return fmt.Errorf("Exiting")
	default:

	}
	req := &message.Request{
		Op: message.Operation{
			Envelope:  env,
			ChannelID: ch.support.ChainID(),
			ConfigSeq: configSeq,
			Type:      message.TYPENORMAL,
		},
		TimeStamp: message.TimeStamp(time.Now().UnixNano()),
		ID:        0,
	}
	ch.pocrNode.SendAllNodes(req)
	return nil
}

func (ch *Chain) Configure(config *cb.Envelope, configSeq uint64) error {
	logger.Info("Config")
	select {
	case <-ch.exitChan:
		logger.Info("[CHAIN error exit config]")
		return fmt.Errorf("Exiting")
	default:
	}
	req := &message.Request{
		Op: message.Operation{
			Envelope:  config,
			ChannelID: ch.support.ChainID(),
			ConfigSeq: configSeq,
			Type:      message.TYPECONFIG,
		},
		TimeStamp: message.TimeStamp(time.Now().UnixNano()),
		ID:        0,
	}
	ch.pocrNode.SendAllNodes(req)
	return nil
}
