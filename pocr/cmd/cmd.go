package cmd

import (
	"encoding/base64"
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/orderer/consensus/pocr/message"
)

type SharedConfig struct {
	ClientServer  bool
	Port          int
	Id            message.Identify
	View          message.View
	Table         map[message.Identify]string
	FaultNum      uint
	ExecuteMaxNum int
	CheckPointNum message.Sequence
	WaterL        message.Sequence
	WaterH        message.Sequence
	// keys and initial CR credits
	PriKeyTable map[message.Identify][]byte
	PubKeyTable map[message.Identify][]byte
	CRTable     map[message.Identify]float64
	// group
	GroupList         []message.Identify
	GroupNum          int
	CommitteeFaultNum uint
}

func ReadConfig() *SharedConfig {
	port, _ := GetConfigurePort()
	id, _ := GetConfigureID()
	table, _ := GetConfigureTable()
	priKeyTable, _ := GetConfigurePriKeyTable()
	pubKeyTable, _ := GetConfigurePubKeyTable()
	crtable, _ := GetConfigureCRTable()
	groupList, _ := GetConfigureGroupList()

	t := make(map[message.Identify]string)
	for k, v := range table {
		t[message.Identify(k)] = v
	}

	c := make(map[message.Identify]float64)
	for k, v := range crtable {
		c[message.Identify(k)] = v
	}

	d := make(map[message.Identify][]byte)
	for k, v := range priKeyTable {
		d[message.Identify(k)] = v
	}

	p := make(map[message.Identify][]byte)
	for k, v := range pubKeyTable {
		p[message.Identify(k)] = v
	}

	// calc the fault num
	if len(t)%3 != 1 {
		log.Fatalf("[Config Error] the incorrent node num : %d, need 3f + 1", len(t))
		return nil
	}

	flag.Parse()
	return &SharedConfig{
		Port:          port,
		Id:            message.Identify(id),
		View:          message.View(0),
		Table:         t,
		FaultNum:      uint(len(t) / 3),
		ExecuteMaxNum: 1,
		CheckPointNum: 100,
		WaterL:        0,
		WaterH:        200,
		// PoCr
		PriKeyTable: d,
		PubKeyTable: p,
		CRTable:     c,
		// group(work packages)
		GroupList:         groupList,
		GroupNum:          len(groupList),
		CommitteeFaultNum: uint(len(groupList) / 3),
	}
}

// obtain information from configuration files
func GetConfigureGroupList() ([]message.Identify, error) {
	rawTable := os.Getenv("POCR_GROUP_TABLE")
	tables := strings.Split(rawTable, ";")
	grouptable := make([]message.Identify, 0, len(tables))
	for _, t := range tables {
		idInt, _ := strconv.Atoi(t)
		grouptable = append(grouptable, message.Identify(idInt))
	}
	log.Printf("groupList: %v", grouptable)
	return grouptable, nil
}

func GetConfigureCRTable() (map[int]float64, error) {
	rawTable := os.Getenv("POCR_NODE_CRTABLE")
	nodeTable := make(map[int]float64, 0)
	tables := strings.Split(rawTable, ";")
	for index, t := range tables {
		nodeTable[index], _ = strconv.ParseFloat(t, 64)
	}
	return nodeTable, nil
}

func GetConfigurePubKeyTable() (map[int][]byte, error) {
	rawTable := os.Getenv("POCR_PUBKEY_TABLE")
	pubKeyTable := make(map[int][]byte)
	tables := strings.Split(rawTable, ";")
	for index, t := range tables {
		bytevalue, err := base64.StdEncoding.DecodeString(t)
		if err != nil {
			return nil, err
		}
		pubKeyTable[index] = bytevalue
	}
	return pubKeyTable, nil
}

func GetConfigurePriKeyTable() (map[int][]byte, error) {
	rawTable := os.Getenv("POCR_PRIBKEY_TABLE")
	priKeyTable := make(map[int][]byte)
	tables := strings.Split(rawTable, ";")
	for index, t := range tables {
		bytevalue, err := base64.StdEncoding.DecodeString(t)
		if err != nil {
			return nil, err
		}
		priKeyTable[index] = bytevalue
	}
	return priKeyTable, nil
}

func GetConfigureID() (id int, err error) {
	rawID := os.Getenv("POCR_NODE_ID")
	if id, err = strconv.Atoi(rawID); err != nil {
		return
	}
	return
}

func GetConfigureTable() (map[int]string, error) {
	rawTable := os.Getenv("POCR_NODE_TABLE") // websites of nodes' servers
	nodeTable := make(map[int]string, 0)

	tables := strings.Split(rawTable, ";")
	for index, t := range tables {
		nodeTable[index] = t
	}
	// satisfy fault tolenrace 3f + 1
	if len(tables) < 3 || len(tables)%3 != 1 {
		return nil, errors.New("")
	}
	return nodeTable, nil
}

func GetConfigurePort() (port int, err error) {
	rawPort := os.Getenv("POCR_LISTEN_PORT")
	if port, err = strconv.Atoi(rawPort); err != nil {
		return
	}
	return
}
