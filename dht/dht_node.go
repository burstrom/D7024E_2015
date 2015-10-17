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
	Successor   *DHTNode
	Predecessor *DHTNode
	// Added manually:
	fingers         [bits]*VNode
	BindAddress     string
	queue           chan *DHTMsg
	FingerResponses int
	online          bool
	lastStab        string
}

type VNode struct {
	nodeId      string
	BindAddress string
	fingerIndex int
}

func MakeDHTNode(nodeId *string, BindAddress string) *DHTNode {
	// Defines the node, and adds the tuple values of IP and Port.
	dhtNode := new(DHTNode)
	//dhtNode.transport = CreateTransport(dhtNode, ip+":"+port)
	dhtNode.BindAddress = BindAddress
	dhtNode.queue = make(chan *DHTMsg)
	//No ID? Let's generate one.
	dhtNode.nodeId = generateNodeId(BindAddress)
	// if nodeId == nil {
	// genNodeId := generateNodeId(*nodeId)
	// dhtNode.nodeId = genNodeId
	// } else {
	// dhtNode.nodeId = *nodeId
	// }
	dhtNode.Successor = nil
	dhtNode.Predecessor = nil
	dhtNode.FingerResponses = 0
	// Added manually:

	// fmt.Println("Node: " + BindAddress)
	dhtNode.online = false
	return dhtNode
}

func makeVNode(nodeId *string, BindAddress string) *VNode {
	vNode := new(VNode)
	vNode.BindAddress = BindAddress
	vNode.nodeId = *nodeId
	return vNode
}

func (dhtNode *DHTNode) StartServer(wg *sync.WaitGroup) {
	wg.Done()
	go dhtNode.handler()
	go dhtNode.startweb()
	go dhtNode.timerSomething()
	dhtNode.listen()
}

func (dhtNode *DHTNode) updateNode(msg *DHTMsg) {
	// Source of msg becomes Predecessor, data has nodeID, ip & port for Successor)
	preNode := MakeDHTNode(&msg.Key, msg.Src)
	dhtNode.Predecessor = preNode
	if msg.Data != "" {
		Successor := strings.Split(msg.Data, ";")
		dhtNode.Successor = MakeDHTNode(&Successor[0], Successor[1])
	}
	// fmt.Println("[UPDT]\t" + dhtNode.Predecessor.nodeId + "\t" + dhtNode.nodeId + "\t" + dhtNode.Successor.nodeId)

}

func (node *DHTNode) printQuery(msg *DHTMsg) {
	// fmt.Println("Node " + node.nodeId + " got [PRINT]")
	// fts := node.FingersToString()
	Infoln(node.Predecessor.BindAddress + " - " + node.BindAddress + " - " + node.Successor.BindAddress)
	if msg.Origin != node.BindAddress {
		// msg.Data = msg.Data + node.Predecessor.BindAddress + "\t" + node.BindAddress + "\t" + node.Successor.BindAddress + "\t\n"
		node.Send("printAll", node.Successor.BindAddress, msg.Origin, msg.Key, msg.Data)
	}
	// } else {
	// str := "Pre.\tNode\tSucc.\n" + msg.Data + node.Predecessor.BindAddress + "\t" + node.BindAddress + "\t" + node.Successor.BindAddress + "\t\n"
	// Noticeln(str)
	//fmt.Print(str)
	//fmt.Print(msg.Data+"\n"+node.Predecessor.nodeId+"\t"+node.nodeId+"\t"+node.Successor.nodeId+"\t"+fingers+"\n", "")
	// }
}

/* 	When a node gets a finger query, it splits the data to get the origin node info and also which index value it should be pointed to.
Then it Sends a response to the origin node with its own ID, Binddress (in the data) comma separated.
*/
func (dhtNode *DHTNode) fingerQuery(msg *DHTMsg) {
	if dhtNode.responsible(msg.Key) {
		// fmt.Println(dhtNode.nodeId + " is responsible for :" + msg.Key)
		go dhtNode.Send("fingerResponse", msg.Origin, "", "", msg.Data)
	} else if dhtNode.BindAddress != msg.Origin {
		// Framtiden köra en accelerated forward?
		// fmt.Println(dhtNode.nodeId + " isn't responsible for : " + msg.Key)
		go dhtNode.SendFrwd(msg, dhtNode.Successor)
	}
	//time.Sleep(200 * time.Millisecond)
	// fmt.Println(msg)

	// go dhtNode.Send("fingerResponse", newNode, data[3], dhtNode.nodeId+":"+dhtNode.bindAdress)
	//time.Sleep(200 * time.Millisecond}
}

func (dhtNode *DHTNode) fingerResponse(msg *DHTMsg) {
	// Source of msg becomes Predecessor, data has nodeID, ip & port for Successor)
	dhtNode.FingerResponses++
	newNode := makeVNode(&msg.Key, msg.Src)
	fIndex, err := strconv.Atoi(msg.Data)
	if err != nil {
		fmt.Println(err)
	}
	dhtNode.fingers[fIndex] = newNode
	if dhtNode.FingerResponses == bits {
		// fmt.Println(dhtNode.nodeId + ".setupFingers() [COMPLETED]")
	}
}

func (dhtNode *DHTNode) joinRing(msg *DHTMsg) {
	newDHTNode := MakeDHTNode(&msg.Key, msg.Src)
	if dhtNode.Successor == nil && dhtNode.Predecessor == nil {
		// fmt.Println("#" + dhtNode.BindAddress + " - join- " + msg.Key)

		dhtNode.Predecessor = newDHTNode
		dhtNode.Successor = newDHTNode
		dhtNode.Send("update", newDHTNode.BindAddress, "", "", dhtNode.nodeId+";"+dhtNode.BindAddress)
		// Infoln("[JOIN]\t" + dhtNode.Predecessor.nodeId + "\t" + dhtNode.nodeId + "\t" + dhtNode.Successor.nodeId)

	} else {
		// Is the node between dhtNode and dhtNode Successor?
		if between([]byte(dhtNode.nodeId), []byte(dhtNode.Successor.nodeId), []byte(newDHTNode.nodeId)) {
			newDHTNode.Successor = dhtNode.Successor
			newDHTNode.Predecessor = dhtNode
			//dhtNode.Send("join", newDHTNode, "response", "")

			dhtNode.Successor = newDHTNode
			// Update Successor node (only Predecessor)
			newDHTNode.Send("update", newDHTNode.Successor.BindAddress, "", "", "")

			// Update the new nodes value, with dhtNode as Predecessor and the data-string as Successor
			dhtNode.Send("update", newDHTNode.BindAddress, "", "", newDHTNode.Successor.nodeId+";"+newDHTNode.Successor.BindAddress)
			// Infoln("[JOIN]\t" + dhtNode.Predecessor.nodeId + "\t" + dhtNode.nodeId + "\t" + dhtNode.Successor.nodeId)
			// fmt.Println("#" + dhtNode.BindAddress + " - join- " + msg.Key)
		} else {
			// TODO: use fingers if the fingertable is updated/setup?
			newDHTNode.Send("join", dhtNode.Successor.BindAddress, "", "", "")
		}
	}
	// Create a call or something to update fingertables??
	time.Sleep(200 * time.Millisecond)

	//dhtNode.setupFingers()
	//tempNode := dhtNode
}

func (dhtNode *DHTNode) printAll() {
	if dhtNode.Successor != nil {
		Infoln("Pre \t\t Cur \t\t Suc")
		dhtNode.Send("printAll", dhtNode.Successor.BindAddress, "", "", "")
	}

}

/* Returns the node which is responsible for key as a Response
Input MSG = {
	Key = The key which is looked up,
	Origin = Original Sender.
	Data = Which type of lookup, is it for a join or normal lookup?
}

Response MSG = {
	Key = Node's ID,
	Data = Node bindadress.
}
*/
func (dhtNode *DHTNode) lookup(msg *DHTMsg) {
	// fmt.Println(dhtNode.BindAddress + " got lookup of type: " + requestType + " from " + msg.Src)
	if dhtNode.Predecessor != nil && dhtNode.Predecessor.nodeId == msg.Key {
		dhtNode.Send(msg.Req, dhtNode.Predecessor.BindAddress, msg.Origin, msg.Key, msg.Data)
		return
	}

	if (dhtNode.nodeId == msg.Key) || (dhtNode.Predecessor == nil && dhtNode.Successor == nil) {
		dhtNode.Send(msg.Data+"Response", msg.Origin, "", "", msg.Key)
		return
	}
	// fmt.Println("Trying to debug the problem ", dhtNode.Predecessor)
	// If the key is equal to its prdecessor

	if dhtNode.Predecessor != nil && between([]byte(dhtNode.Predecessor.nodeId), []byte(dhtNode.nodeId), []byte(msg.Key)) { // if the key is between the nodes Predecessor and itself.
		dhtNode.Send(msg.Data+"Response", msg.Origin, "", "", msg.Key)
		return
	}
	if between([]byte(dhtNode.nodeId), []byte(dhtNode.Successor.nodeId), []byte(msg.Key)) {
		dhtNode.Send(msg.Req, dhtNode.Successor.BindAddress, msg.Origin, msg.Key, msg.Data)
		return
	}
	if dhtNode.FingerResponses != bits {
		dhtNode.Send(msg.Req, dhtNode.Successor.BindAddress, msg.Origin, msg.Key, msg.Data)
	} else if between([]byte(dhtNode.nodeId), []byte(dhtNode.fingers[bits-1].nodeId), []byte(msg.Key)) {
		// fmt.Println(dhtNode.nodeId + " got from: " + msg.Src + " with key: " + msg.Key)
		for k := bits - 1; k >= 0; k-- {
			if dhtNode.fingers[k] == nil {

			} else if between([]byte(dhtNode.nodeId), []byte(dhtNode.fingers[k].nodeId), []byte(msg.Key)) == false {
				dhtNode.Send(msg.Req, dhtNode.fingers[k].BindAddress, msg.Origin, msg.Key, msg.Data)
				return
			}
		}
		go dhtNode.Send(msg.Req, dhtNode.Successor.BindAddress, msg.Origin, msg.Key, msg.Data)

	} else {
		dhtNode.Send(msg.Req, dhtNode.fingers[len(dhtNode.fingers)-1].BindAddress, msg.Origin, msg.Key, msg.Data)
	}

}

/*
func (dhtNode *DHTNode) acceleratedLookupUsingFingers(key string) *DHTNode {
	// If the node or it's Successor is responsible for the key?
	// Uses fingers to achieve a logarithmic lookup instead of linear.

	counter = counter + 1
	if dhtNode.responsible(key) {
		return dhtNode
	}
	fmt.Println("Is Successor responsible?")
	if dhtNode.Successor.responsible(key) {
		return dhtNode.Successor
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
			fingers[k] = dhtNode.fingers[k].BindAddress
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
	// fmt.Println(dhtNode.nodeId + ".setupFingers()")
	dhtNode.FingerResponses = 0
	for k := 0; k < bits; k++ {
		idBytes, _ := hex.DecodeString(dhtNode.nodeId)
		//fingerID, _ := calcFinger(idBytes, k+1, bits)
		fingerHex, _ := calcFinger(idBytes, k, bits)
		//fmt.Print(" ' " + fingerHex)
		kstr := strconv.Itoa(k)
		// time.Sleep(50 * time.Millisecond)
		// fmt.Println(idBytes, " finger search for "+kstr)
		if dhtNode.Successor != nil {
			dhtNode.Send("fingerQuery", dhtNode.Successor.BindAddress, "", fingerHex, kstr)
		}

	}
	// fmt.Println("")

}

func (dhtNode *DHTNode) Send(req, dst, origin, key, data string) {
	// If the origin is empty, then it becomes the DHTNodes bind adress since it was the one who sent the first.
	/*if req == "LookupResponse" {
		fmt.Println("dst: " + dst + ", origin: " + origin + ", key: " + key + ", data: " + data)
	}*/

	if key == "" {
		key = dhtNode.nodeId
	}
	if origin == "" {
		origin = dhtNode.BindAddress
	}
	msg := CreateMsg(req, dhtNode.BindAddress, origin, key, data)

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
func (dhtNode *DHTNode) SendFrwd(msg *DHTMsg, dstNode *DHTNode) {
	msg.Src = dhtNode.BindAddress
	dhtNode.Send(msg.Req, dstNode.BindAddress, msg.Origin, msg.Key, msg.Data)
}

// func (dhtNode *DHTNode) printFingers() string {
// 	output := "{"
// 	fingerList := string[bits]
// 	finger := dhtNode.fingers[0]
// 	for k := 0; k < bits; k++ {
// 		finger = dhtNode.fingers[k]
// 		// output = output + "," + finger.nodeId + " " + finger.BindAddress
// 		fingerList[k] = finger.BindAddress
// 	}
// 	output = dhtNode.nodeId + ": " + output + "}"
// 	return output
// }

func (dhtNode *DHTNode) responsible(key string) bool {
	// key == dhtnode?
	if bytes.Compare([]byte(dhtNode.nodeId), []byte(key)) == 0 {
		return true
	}
	// If key > Predecessor or <= dhtnode
	if bytes.Compare([]byte(dhtNode.Predecessor.nodeId), []byte(key)) == -1 || bytes.Compare([]byte(dhtNode.nodeId), []byte(key)) >= 0 {
		return between([]byte(dhtNode.Predecessor.nodeId), []byte(dhtNode.nodeId), []byte(key))
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

func (node *DHTNode) timerSomething() {
	k := 0
	for {
		if node.Successor != nil {
			// node.Send("notify", node.Successor.BindAddress, "", "", "")
			node.Send("getPredecessor", node.Successor.BindAddress, "", "", "")
			/*node.Send("PredQuery", node.Successor.BindAddress, "", "", node.nodeId+";"+node.BindAddress)*/
		}
		if k == 3 {
			// Warnln("Timer reset...")
			node.Send("fingerSetup", node.BindAddress, "", "", "")
			k = 0
		}
		time.Sleep(1000 * time.Millisecond)
		k++
	}
}
