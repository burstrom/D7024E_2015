package dht

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time" // added time method
)

type DHTMsg struct {
	//Timestamp int64  `json:"time"`
	Key    string `json:"key"`    // Key value
	Src    string `json:"src"`    // Source of message
	Req    string `json:"req"`    //destnation
	Origin string `json:"origin"` // Original sender of message.
	Data   string `json:"data"`   // Data?

	//Dst    string `json:"dst"`    // Destination of message
	// Opt string `json:"opt"` // Option?
	//Origin string `json:"origin"`
	//Bytes string `json:"bytes"`
}

// Create msg creates the DHTMsg with Timestamp, Key, Src, Req, Origin and Data.
func CreateMsg(req, src, origin, key, data string) *DHTMsg {
	dhtMsg := &DHTMsg{}
	//dhtMsg.Timestamp = Now() // Might cause problems?
	dhtMsg.Key = key
	dhtMsg.Src = src
	dhtMsg.Origin = origin
	dhtMsg.Data = data
	dhtMsg.Req = req
	/*Req := strings.Split(req, ",")
	for key := range Req {
		fmt.Println("Key:", key)
	}*
	dhtMsg.Req = ""*/

	return dhtMsg
}

func (node *DHTNode) handler() {
	for {
		select {
		case msg := <-node.queue:
			switch msg.Req {
			case "lookupResponse":
				// vNode := makeVNode(&msg.Key, msg.Data)
				Notice(node.nodeId + " Response: " + msg.Data + ", from " + msg.Key + "\n")
			case "joinResponse":
				node.successor = makeDHTNode(&msg.Key, msg.Src)
				if node.predecessor == nil {
					node.predecessor = node.successor
				}
				node.send("notify", msg.Src, "", "", "")
			case "join":
				// node.joinRing(msg)
				node.lookup("join", msg)
			case "update":
				node.updateNode(msg)
			case "lookup":
				node.lookup("key", msg)
			case "fingerQuery":
				node.fingerQuery(msg)
			case "fingerSetup":
				node.setupFingers()
			case "fingerResponse":
				node.fingerResponse(msg)
			case "printAll":
				// fmt.Println(node.predecessor.nodeId + "\t" + node.nodeId + "\t" + node.successor.nodeId)
				node.printQuery(msg)
			case "PredQuery":
				node.PredResponse(msg)
			case "PredQueryResponse":
				node.PredQueryResponse(msg)
			case "notify":
				node.notify(msg)
			case "kill":
				fmt.Println("Kill all connections/threads related to this node?")
			}

		}
	}
}
func (dhtNode *DHTNode) PredResponse(msg *DHTMsg) {
	// fmt.Println("src " + msg.Src + " dst : " + dhtNode.bindAddress)
	if dhtNode.predecessor == nil {
		dhtNode.send("PredQueryResponse", dhtNode.predecessor.bindAddress, "", "", "")
	} else {
		dhtNode.send("PredQueryResponse", dhtNode.predecessor.bindAddress, "", "", dhtNode.predecessor.nodeId+";"+dhtNode.predecessor.bindAddress)
	}

}

func (dhtNode *DHTNode) PredQueryResponse(msg *DHTMsg) {
	dhtNode.lastStab = ""
	// fmt.Println(msg)
	// fmt.Println("src " + msg.Src + " dst : " + dhtNode.bindAddress)
	// src := dhtNode.bindAddress
	key := strings.Split(msg.Data, ";")
	if between([]byte(dhtNode.nodeId), []byte(dhtNode.successor.nodeId), []byte(key[0])) && key[0] != "" && dhtNode.nodeId != key[0] {

		// temp := strings.Split(msg.Data, ";")
		dhtNode.successor.nodeId = key[0]
		dhtNode.successor.bindAddress = key[1]

	}
	// dhtNode.send("notify", dhtNode.successor.bindAddress, dhtNode.successor.nodeId+";"+dhtNode.successor.bindAddress)

}
func (dhtNode *DHTNode) stabilize() {
	dhtNode.send("PredQuery", dhtNode.successor.bindAddress, "", "", dhtNode.nodeId+";"+dhtNode.bindAddress)
	dhtNode.lastStab = dhtNode.successor.bindAddress
	time.Sleep(50 * time.Millisecond)
	if dhtNode.lastStab != "" {
		// redo the stabilize but for the next successor in the succerlist.
	}
}

func (dhtNode *DHTNode) notify(msg *DHTMsg) {
	if dhtNode.predecessor == nil || between([]byte(dhtNode.predecessor.nodeId), []byte(dhtNode.nodeId), []byte(msg.Key)) {

		// fmt.Println(dhtNode.nodeId + ".Notify(" + msg.Key + ")")

		temp := dhtNode.predecessor
		dhtNode.predecessor = makeDHTNode(&msg.Key, msg.Src)
		if dhtNode.successor == nil {
			dhtNode.successor = dhtNode.predecessor
		}

		time.Sleep(50 * time.Millisecond)
		if temp != nil {
			dhtNode.send("PredQueryResponse", temp.bindAddress, "", "", dhtNode.predecessor.nodeId+";"+dhtNode.predecessor.bindAddress)
		}
		// temp := strings.Split(msg.Data, ";")
		// dhtNode.predecessor.nodeId = temp[0]
		// dhtNode.predecessor.bindAddress = temp[1]
		// Notice(dhtNode.predecessor.nodeId + "\t" + dhtNode.nodeId + "\t" + dhtNode.successor.nodeId)
	}
	return
}

func (node *DHTNode) listen() {
	udpAddr, err := net.ResolveUDPAddr("udp", node.bindAddress)
	conn, err := net.ListenUDP("udp", udpAddr)
	node.online = true
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	dec := json.NewDecoder(conn)
	// fmt.Println("Started listening : " + node.bindAdress)
	Error("Started listening : " + node.bindAddress + "\n")

	for {
		if node.online {
			msg := DHTMsg{}
			err = dec.Decode(&msg)
			if err != nil {
				fmt.Println(err.Error())
			}
			node.queue <- &msg
		} else {
			time.Sleep(10000 * time.Millisecond)

		}
	}
}
