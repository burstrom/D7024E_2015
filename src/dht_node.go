package dht

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// go test -test.run <functionsnamn>

/*
	<ACTIVE THREADS>
		* go dhtNode.handler() which loops and handles all DHTMessages which is added to the queue
		* go dhtNode.listen() which loops and listens on a specific port.
*/

var counter = 0

type DHTNode struct {
	nodeId      string
	successor   *DHTNode
	predecessor *DHTNode
	// Added manually:
	fingers    [bits]*DHTNode
	bindAdress string
	queue      chan *DHTMsg
}

func makeDHTNode(nodeId *string, ip string, port string) *DHTNode {
	// Defines the node, and adds the tuple values of IP and Port.
	dhtNode := new(DHTNode)
	//dhtNode.transport = CreateTransport(dhtNode, ip+":"+port)
	dhtNode.bindAdress = ip + ":" + port
	dhtNode.queue = make(chan *DHTMsg)
	//No ID? Let's generate one.
	if nodeId == nil {
		genNodeId := generateNodeId()
		dhtNode.nodeId = genNodeId
	} else {
		dhtNode.nodeId = *nodeId
	}
	dhtNode.successor = nil
	dhtNode.predecessor = nil
	// Added manually:
	go dhtNode.handler()
	return dhtNode
}

func (dhtNode *DHTNode) startServer(wg *sync.WaitGroup) {
	wg.Done()
	dhtNode.listen()
}

func (dhtNode *DHTNode) updateNode(msg *DHTMsg) {
	// Source of msg becomes predecessor, data has nodeID, ip & port for successor)
	ip_port := strings.Split(msg.Src, ":")

	preNode := makeDHTNode(&msg.Key, ip_port[0], ip_port[1])
	dhtNode.predecessor = preNode
	successor := strings.Split(msg.Data, ":")
	if msg.Data != "" {
		dhtNode.successor = makeDHTNode(&successor[0], successor[1], successor[2])
	}
}

/* 	When a node gets a finger query, it splits the data to get the origin node info and also which index value it should be pointed to.
Then it sends a response to the origin node with its own ID, Binddress (in the data) comma separated.
*/
func (dhtNode *DHTNode) fingerQuery(msg *DHTMsg) {
	data := strings.Split(msg.Data, ":")
	newNode := makeDHTNode(&data[0], data[1], data[2])
	//time.Sleep(200 * time.Millisecond)
	go dhtNode.send("fingerResponse", newNode, data[3], dhtNode.nodeId+":"+dhtNode.bindAdress)
	//time.Sleep(200 * time.Millisecond}
}
func (dhtNode *DHTNode) fingerResponse(msg *DHTMsg) {
	// Source of msg becomes predecessor, data has nodeID, ip & port for successor)
	data := strings.Split(msg.Data, ":")
	newNode := makeDHTNode(&data[0], data[1], data[2])
	fIndex, err := strconv.Atoi(msg.Opt)
	if err != nil {
		fmt.Println(err)
	}
	dhtNode.fingers[fIndex] = newNode
	// fmt.Println("Node #" + dhtNode.nodeId + " Got a finger RESPONSE from: " + msg.Src + " K Value = " + msg.Opt)

	//dhtNode.fingers[]
}

func (dhtNode *DHTNode) joinRing(msg *DHTMsg) {
	ip_port := strings.Split(msg.Src, ":")
	nodeid := msg.Key
	newDHTNode := makeDHTNode(&nodeid, ip_port[0], ip_port[1])

	if dhtNode.successor == nil && dhtNode.predecessor == nil {
		dhtNode.predecessor = newDHTNode
		dhtNode.successor = newDHTNode
		dhtNode.send("update", newDHTNode, "", dhtNode.nodeId+":"+dhtNode.bindAdress)

	} else {
		// Is the node between dhtNode and dhtNode successor?
		if between([]byte(dhtNode.nodeId), []byte(dhtNode.successor.nodeId), []byte(newDHTNode.nodeId)) {
			newDHTNode.successor = dhtNode.successor
			newDHTNode.predecessor = dhtNode
			//dhtNode.send("join", newDHTNode, "response", "")

			dhtNode.successor = newDHTNode
			// Update successor node (only predecessor)
			newDHTNode.send("update", newDHTNode.successor, "", "")

			// Update the new nodes value, with dhtNode as predecessor and the data-string as successor
			dhtNode.send("update", newDHTNode, "", newDHTNode.successor.nodeId+":"+newDHTNode.successor.bindAdress)

		} else {
			// Should use fingers?
			newDHTNode.send("join", dhtNode.successor, "", "")
		}
	}
	// Create a call or something to update fingertables??
	time.Sleep(200 * time.Millisecond)

	//dhtNode.setupFingers()
	//tempNode := dhtNode
}

func (dhtNode *DHTNode) printAll() {
	dhtNode.send("printall", dhtNode.successor, dhtNode.nodeId, "")
}

func (dhtNode *DHTNode) lookup(key string) *DHTNode {
	//Linear lookup function, go through everynode until responsible is found
	//fmt.Println("Checking for key: " + key + " Against" + dhtNode.nodeId)
	if dhtNode.responsible(key) {
		fmt.Println("FINISHED??")
		return dhtNode
	}
	/*activeNode := dhtNode.successor
	i := 1
	for activeNode.nodeId != dhtNode.nodeId {
		fmt.Println("loop: ", i)
		i = i + 1
		if activeNode.responsible(key) {
			return activeNode
		}
		//fmt.Println("activenode succ:" + activeNode.successor.nodeId + " stopnode: " + dhtNode.nodeId)
		activeNode = activeNode.successor
	}*/
	//return dhtNode
	return dhtNode.successor.lookup(key)
}

func (dhtNode *DHTNode) acceleratedLookupUsingFingers(key string) *DHTNode {
	// If the node or it's successor is responsible for the key?
	// Uses fingers to achieve a logarithmic lookup instead of linear.

	counter = counter + 1
	if dhtNode.responsible(key) {
		return dhtNode
	}
	fmt.Println("Is successor responsible?")
	if dhtNode.successor.responsible(key) {
		return dhtNode.successor
	}
	// Is the key within the interval node1 - node1.fingers[last] ?
	if between([]byte(dhtNode.nodeId), []byte(dhtNode.fingers[len(dhtNode.fingers)-1].nodeId), []byte(key)) {
		// if the key is within the interval, decrease the value k, check if the key still is in the interval?
		for k := 1; k <= bits; k++ {
			if !between([]byte(dhtNode.nodeId), []byte(dhtNode.fingers[len(dhtNode.fingers)-k].nodeId), []byte(key)) {
				return dhtNode.fingers[len(dhtNode.fingers)-k].acceleratedLookupUsingFingers(key)
			}
		}
	}
	// If the key isn't within the interval, it must be within another interval calculated from the last finger of node1
	return dhtNode.fingers[len(dhtNode.fingers)-1].acceleratedLookupUsingFingers(key)
}

func (dhtNode *DHTNode) responsible(key string) bool {
	// Is the key the same key as the node?
	if dhtNode.nodeId == key {
		//fmt.Println("Node " + dhtNode.nodeId + "[TRUE]")
		return true
	}
	if dhtNode.predecessor.nodeId == key {
		//fmt.Println("Node+" + dhtNode.nodeId + ".pre " + dhtNode.predecessor.nodeId + "[FALSE]")
		return false
	}
	// Is the key value between
	//fmt.Println("Return between: " + dhtNode.predecessor.nodeId + " - " + dhtNode.nodeId + " Key: " + key)
	return between([]byte(dhtNode.predecessor.nodeId), []byte(dhtNode.nodeId), []byte(key))
}

func (dhtNode *DHTNode) FingersToString() string {
	//fmt.Print("#" + dhtNode.nodeId + " :> ")
	returnval := "{"
	for k := 0; k < len(dhtNode.fingers); k++ {
		if dhtNode.fingers[k] != nil {
			returnval = returnval + dhtNode.fingers[k].nodeId + " "
		}
	}
	returnval = returnval + "}"
	return returnval
}

func (dhtNode *DHTNode) setupFingers() {
	// fmt.Println("From node " + dhtNode.nodeId)
	//kString := ""
	for k := 0; k < bits; k++ {
		idBytes, _ := hex.DecodeString(dhtNode.nodeId)
		//fingerID, _ := calcFinger(idBytes, k+1, bits)
		fingerHex, _ := calcFinger(idBytes, k+1, bits)
		kstr := strconv.Itoa(k)
		dhtNode.send("fingerQuery", dhtNode.successor, fingerHex, dhtNode.nodeId+":"+dhtNode.bindAdress+":"+kstr)

	}

}

func (dhtNode *DHTNode) send(req string, dstNode *DHTNode, opt, data string) {

	msg := CreateMsg(dhtNode.nodeId, dhtNode.bindAdress, dstNode.bindAdress, req, opt, data)
	// fmt.Println("Message:", msg)
	udpAddr, err := net.ResolveUDPAddr("udp", msg.Dst)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	res, _ := json.Marshal(msg)
	_, err = conn.Write(res) // wat?

}

func (dhtNode *DHTNode) sendFrwd(msg *DHTMsg, dstNode *DHTNode) {
	originNodeVal := strings.Split(msg.Data, ":")
	originNode := makeDHTNode(&originNodeVal[0], originNodeVal[1], originNodeVal[2])
	if originNode.nodeId != dhtNode.nodeId {
		originNode.send(msg.Req, dstNode, msg.Opt, msg.Data)
	}

}

func (dhtNode *DHTNode) printFingers() string {
	output := "{"
	finger := dhtNode.fingers[0]
	for k := 0; k < bits; k++ {
		finger = dhtNode.fingers[k]
		output = output + "," + finger.nodeId + " " + finger.bindAdress
	}
	output = dhtNode.nodeId + ": " + output + "}"
	return output
}
