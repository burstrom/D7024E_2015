package dht

import (
	"bytes"
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
	fingers         [bits]*VNode
	bindAddress     string
	queue           chan *DHTMsg
	fingerResponses int
}

type VNode struct {
	nodeId      string
	bindAddress string
	fingerIndex int
}

func makeDHTNode(nodeId *string, bindAddress string) *DHTNode {
	// Defines the node, and adds the tuple values of IP and Port.
	dhtNode := new(DHTNode)
	//dhtNode.transport = CreateTransport(dhtNode, ip+":"+port)
	dhtNode.bindAddress = bindAddress
	dhtNode.queue = make(chan *DHTMsg)
	//No ID? Let's generate one.
	// dhtNode.nodeId = generateNodeId(bindAddress)
	if nodeId == nil {
		genNodeId := generateNodeId("0")
		dhtNode.nodeId = genNodeId
	} else {
		dhtNode.nodeId = *nodeId
	}
	dhtNode.successor = nil
	dhtNode.predecessor = nil
	dhtNode.fingerResponses = 0
	// Added manually:
	go dhtNode.handler()
	// fmt.Println("Node: " + bindAddress)

	return dhtNode
}

func makeVNode(nodeId *string, bindAddress string) *VNode {
	vNode := new(VNode)
	vNode.bindAddress = bindAddress
	vNode.nodeId = *nodeId
	return vNode
}

func (dhtNode *DHTNode) startServer(wg *sync.WaitGroup) {
	wg.Done()
	go dhtNode.startweb()
	dhtNode.listen()
}

func (dhtNode *DHTNode) updateNode(msg *DHTMsg) {
	// Source of msg becomes predecessor, data has nodeID, ip & port for successor)
	preNode := makeDHTNode(&msg.Key, msg.Src)
	dhtNode.predecessor = preNode
	if msg.Data != "" {
		successor := strings.Split(msg.Data, ";")
		dhtNode.successor = makeDHTNode(&successor[0], successor[1])
	}
	// fmt.Println("[UPDT]\t" + dhtNode.predecessor.nodeId + "\t" + dhtNode.nodeId + "\t" + dhtNode.successor.nodeId)

}

func (node *DHTNode) printQuery(msg *DHTMsg) {
	// fmt.Println("Node " + node.nodeId + " got [PRINT]")
	// fts := node.FingersToString()

	if msg.Origin != node.bindAddress {
		msg.Data = msg.Data + node.predecessor.nodeId + "\t" + node.nodeId + "\t" + node.successor.nodeId + "\t\n"
		node.send("printAll", node.successor.bindAddress, msg.Origin, msg.Key, msg.Data)

	} else {
		str := "Pre.\tNode\tSucc.\n" + msg.Data + node.predecessor.nodeId + "\t" + node.nodeId + "\t" + node.successor.nodeId + "\t\n"
		Noticeln(str)
		//fmt.Print(str)
		//fmt.Print(msg.Data+"\n"+node.predecessor.nodeId+"\t"+node.nodeId+"\t"+node.successor.nodeId+"\t"+fingers+"\n", "")
	}
}


/* 	When a node gets a finger query, it splits the data to get the origin node info and also which index value it should be pointed to.
Then it sends a response to the origin node with its own ID, Binddress (in the data) comma separated.
*/
func (dhtNode *DHTNode) fingerQuery(msg *DHTMsg) {
	if dhtNode.responsible(msg.Key) {
		// fmt.Println(dhtNode.nodeId + " is responsible for :" + msg.Key)
		go dhtNode.send("fingerResponse", msg.Origin, "", "", msg.Data)
	} else if dhtNode.bindAddress != msg.Origin {
		// Framtiden köra en accelerated forward?
		// fmt.Println(dhtNode.nodeId + " isn't responsible for : " + msg.Key)
		go dhtNode.sendFrwd(msg, dhtNode.successor)
	}
	//time.Sleep(200 * time.Millisecond)
	// fmt.Println(msg)

	// go dhtNode.send("fingerResponse", newNode, data[3], dhtNode.nodeId+":"+dhtNode.bindAdress)
	//time.Sleep(200 * time.Millisecond}
}

func (dhtNode *DHTNode) fingerResponse(msg *DHTMsg) {
	// Source of msg becomes predecessor, data has nodeID, ip & port for successor)
	dhtNode.fingerResponses++
	newNode := makeVNode(&msg.Key, msg.Src)
	fIndex, err := strconv.Atoi(msg.Data)
	if err != nil {
		fmt.Println(err)
	}
	dhtNode.fingers[fIndex] = newNode
}

func (dhtNode *DHTNode) joinRing(msg *DHTMsg) {
	newDHTNode := makeDHTNode(&msg.Key, msg.Src)
	if dhtNode.successor == nil && dhtNode.predecessor == nil {
		// fmt.Println("#" + dhtNode.bindAddress + " - join- " + msg.Key)

		dhtNode.predecessor = newDHTNode
		dhtNode.successor = newDHTNode
		dhtNode.send("update", newDHTNode.bindAddress, "", "", dhtNode.nodeId+";"+dhtNode.bindAddress)
		// Infoln("[JOIN]\t" + dhtNode.predecessor.nodeId + "\t" + dhtNode.nodeId + "\t" + dhtNode.successor.nodeId)

	} else {
		// Is the node between dhtNode and dhtNode successor?
		if between([]byte(dhtNode.nodeId), []byte(dhtNode.successor.nodeId), []byte(newDHTNode.nodeId)) {
			newDHTNode.successor = dhtNode.successor
			newDHTNode.predecessor = dhtNode
			//dhtNode.send("join", newDHTNode, "response", "")

			dhtNode.successor = newDHTNode
			// Update successor node (only predecessor)
			newDHTNode.send("update", newDHTNode.successor.bindAddress, "", "", "")

			// Update the new nodes value, with dhtNode as predecessor and the data-string as successor
			dhtNode.send("update", newDHTNode.bindAddress, "", "", newDHTNode.successor.nodeId+";"+newDHTNode.successor.bindAddress)
			// Infoln("[JOIN]\t" + dhtNode.predecessor.nodeId + "\t" + dhtNode.nodeId + "\t" + dhtNode.successor.nodeId)
			// fmt.Println("#" + dhtNode.bindAddress + " - join- " + msg.Key)
		} else {
			// Should use fingers?
			newDHTNode.send("join", dhtNode.successor.bindAddress, "", "", "")
		}
	}
	// Create a call or something to update fingertables??
	time.Sleep(200 * time.Millisecond)

	//dhtNode.setupFingers()
	//tempNode := dhtNode
}

func (dhtNode *DHTNode) printAll() {
	if dhtNode.successor != nil {
		dhtNode.send("printAll", dhtNode.successor.bindAddress, "", "", "")
	}

}

/* Returns the node which is responsible for key as a Response
Input MSG = {
	Key = The key which is looked up,
	Origin = Original sender.
	Data = Which type of lookup, is it for a join or normal lookup?
}

Response MSG = {
	Key = Node's ID,
	Data = Node bindadress.
}
*/
func (dhtNode *DHTNode) lookup(msg *DHTMsg) {
	fmt.Println(dhtNode.fingers)

	// If the node is between Node <- <- Node.successor?
	if dhtNode.predecessor.nodeId == msg.Key {
		dhtNode.predecessor.send("lookupResponse", msg.Origin, "", "", msg.Key)
		return
	}
	if dhtNode.nodeId == msg.Key || between([]byte(dhtNode.predecessor.nodeId), []byte(dhtNode.nodeId), []byte(msg.Key)) {
		dhtNode.send("lookupResponse", msg.Origin, "", "", msg.Key)
		return
	}
	if dhtNode.successor.nodeId == msg.Key {
		dhtNode.successor.send("lookupResponse", msg.Origin, "", "", msg.Key)
		return
	}

	if dhtNode.fingerResponses != bits {
		dhtNode.send("lookup", dhtNode.successor.bindAddress, msg.Origin, msg.Key, msg.Data)
	} else if between([]byte(dhtNode.nodeId), []byte(dhtNode.fingers[bits-1].nodeId), []byte(msg.Key)) {
		fmt.Println(dhtNode.nodeId + " got from: " + msg.Src + " with key: " + msg.Key)
		for k := bits - 1; k >= 0; k-- {
			if dhtNode.fingers[k] == nil {

			} else if between([]byte(dhtNode.nodeId), []byte(dhtNode.fingers[k].nodeId), []byte(msg.Key)) == false {
				dhtNode.send("lookup", dhtNode.fingers[k].bindAddress, msg.Origin, msg.Key, msg.Data)
				return
			}
		}
		go dhtNode.send("lookup", dhtNode.successor.bindAddress, msg.Origin, msg.Key, msg.Data)

	} else {
		dhtNode.send("lookup", dhtNode.fingers[len(dhtNode.fingers)-1].bindAddress, msg.Origin, msg.Key, msg.Data)
	}

}

/*
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
}*/

// Prints all unique fingers
func (dhtNode *DHTNode) FingersToString() string {
	//fmt.Print("#" + dhtNode.nodeId + " :> ")
	// returnval := ""
	fingers := make([]string, bits)
	for k := 0; k < bits; k++ {
		if dhtNode.fingers[k] != nil {
			fingers[k] = dhtNode.fingers[k].bindAddress
		}
	}
	// newList :=
	strList := removeDuplicatesUnordered(fingers)
	return strings.Join(strList, " ")
	// returnval = "{"+returnval + "}"
	// return returnval
}

func (dhtNode *DHTNode) setupFingers() {
	// fmt.Println("From node " + dhtNode.nodeId)
	//kString := ""
	// fmt.Print("Node " + dhtNode.nodeId + " Fing:")
	dhtNode.fingerResponses = 0
	for k := 0; k < bits; k++ {
		idBytes, _ := hex.DecodeString(dhtNode.nodeId)
		//fingerID, _ := calcFinger(idBytes, k+1, bits)
		fingerHex, _ := calcFinger(idBytes, k, bits)
		//fmt.Print(" ' " + fingerHex)
		kstr := strconv.Itoa(k)
		// fmt.Println(idBytes, " finger search for "+kstr)
		dhtNode.send("fingerQuery", dhtNode.successor.bindAddress, "", fingerHex, kstr)
	}
	// fmt.Println("")

}

func (dhtNode *DHTNode) send(req, dst, origin, key, data string) {
	// If the origin is empty, then it becomes the DHTNodes bind adress since it was the one who sent the first.
	/*if req == "LookupResponse" {
		fmt.Println("dst: " + dst + ", origin: " + origin + ", key: " + key + ", data: " + data)
	}*/

	if key == "" {
		key = dhtNode.nodeId
	}
	if origin == "" {
		origin = dhtNode.bindAddress
	}
	msg := CreateMsg(req, dhtNode.bindAddress, origin, key, data)

	udpAddr, err := net.ResolveUDPAddr("udp", dst)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	res, _ := json.Marshal(msg)
	_, err = conn.Write(res) // wat?
}

// Forwards the message from the origin node (which is sent in the data file)
func (dhtNode *DHTNode) sendFrwd(msg *DHTMsg, dstNode *DHTNode) {
	msg.Src = dhtNode.bindAddress
	dhtNode.send(msg.Req, dstNode.bindAddress, msg.Origin, msg.Key, msg.Data)
}

// func (dhtNode *DHTNode) printFingers() string {
// 	output := "{"
// 	fingerList := string[bits]
// 	finger := dhtNode.fingers[0]
// 	for k := 0; k < bits; k++ {
// 		finger = dhtNode.fingers[k]
// 		// output = output + "," + finger.nodeId + " " + finger.bindAddress
// 		fingerList[k] = finger.bindAddress
// 	}
// 	output = dhtNode.nodeId + ": " + output + "}"
// 	return output
// }

func (dhtNode *DHTNode) responsible(key string) bool {
	// key == dhtnode?
	if bytes.Compare([]byte(dhtNode.nodeId), []byte(key)) == 0 {
		return true
	}
	// If key > predecessor or <= dhtnode
	if bytes.Compare([]byte(dhtNode.predecessor.nodeId), []byte(key)) == -1 || bytes.Compare([]byte(dhtNode.nodeId), []byte(key)) >= 0 {
		return between([]byte(dhtNode.predecessor.nodeId), []byte(dhtNode.nodeId), []byte(key))
	}
	return false

}

func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key, _ := range encountered {
		result = append(result, key)
	}
	return result
}
