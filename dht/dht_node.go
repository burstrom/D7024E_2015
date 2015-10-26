package dht

import (
	// "bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
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
		* go dhtNode.web() which handles all http messages.
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
	hashMap         map[string]http.ResponseWriter
	fileMap         map[string]string
	Connection      *net.UDPConn
	path            string
}

type VNode struct {
	nodeId      string
	BindAddress string
	fingerIndex int
}

func MakeDHTNode(nodeId *string, BindAddress string) *DHTNode {
	dhtNode := new(DHTNode)
	dhtNode.BindAddress = BindAddress
	dhtNode.queue = make(chan *DHTMsg)
	dhtNode.Connection = nil
	dhtNode.nodeId = generateNodeId(BindAddress)
	dhtNode.Successor = nil
	dhtNode.Predecessor = nil
	dhtNode.FingerResponses = 0
	dhtNode.online = false
	dhtNode.hashMap = make(map[string]http.ResponseWriter)
	dhtNode.fileMap = make(map[string]string)
	vPath := strings.Split(BindAddress, ":")
	dhtNode.path = "storage/" + vPath[1] + "/"
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

func (dhtNode *DHTNode) lookup(msg *DHTMsg) {
	// fmt.Println(dhtNode.BindAddress + " got lookup of type: " + requestType + " from " + msg.Src)
	if dhtNode.Predecessor != nil && dhtNode.Predecessor.nodeId == msg.Key {
		dhtNode.Send2(msg.MsgId, msg.Req, dhtNode.Predecessor.BindAddress, msg.Origin, msg.Key, msg.Data)
		return
	}

	if (dhtNode.nodeId == msg.Key) || (dhtNode.Predecessor == nil && dhtNode.Successor == nil) {
		dhtNode.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", msg.Key, msg.Data)
		if msg.Req == "upload" {
			data := strings.Split(msg.Data, ";")
			dhtNode.upload(dhtNode.path+"root/", data[0], data[1])
			dhtNode.Send("replicate", dhtNode.Successor.BindAddress, dhtNode.BindAddress, msg.Key, msg.Data)
		}
		return
	}
	// fmt.Println("Trying to debug the problem ", dhtNode.Predecessor)
	// If the key is equal to its prdecessor

	if dhtNode.Predecessor != nil && between([]byte(dhtNode.Predecessor.nodeId), []byte(dhtNode.nodeId), []byte(msg.Key)) { // if the key is between the nodes Predecessor and itself.
		dhtNode.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", msg.Key, msg.Data)
		if msg.Req == "upload" {
			data := strings.Split(msg.Data, ";")
			dhtNode.upload(dhtNode.path+"root/", data[0], data[1])
			dhtNode.Send("replicate", dhtNode.Successor.BindAddress, dhtNode.BindAddress, msg.Key, msg.Data)
		}
		return
	}
	if between([]byte(dhtNode.nodeId), []byte(dhtNode.Successor.nodeId), []byte(msg.Key)) {
		dhtNode.Send2(msg.MsgId, msg.Req, dhtNode.Successor.BindAddress, msg.Origin, msg.Key, msg.Data)
		return
	}
	if dhtNode.FingerResponses != bits {
		dhtNode.Send2(msg.MsgId, msg.Req, dhtNode.Successor.BindAddress, msg.Origin, msg.Key, msg.Data)
	} else if between([]byte(dhtNode.nodeId), []byte(dhtNode.fingers[bits-1].nodeId), []byte(msg.Key)) {
		// fmt.Println(dhtNode.nodeId + " got from: " + msg.Src + " with key: " + msg.Key)
		for k := bits - 1; k >= 0; k-- {
			if dhtNode.fingers[k] == nil {

			} else if between([]byte(dhtNode.nodeId), []byte(dhtNode.fingers[k].nodeId), []byte(msg.Key)) == false {
				dhtNode.Send(msg.Req, dhtNode.fingers[k].BindAddress, msg.Origin, msg.Key, msg.Data)
				return
			}
		}
		go dhtNode.Send2(msg.MsgId, msg.Req, dhtNode.Successor.BindAddress, msg.Origin, msg.Key, msg.Data)

	} else {
		dhtNode.Send2(msg.MsgId, msg.Req, dhtNode.fingers[len(dhtNode.fingers)-1].BindAddress, msg.Origin, msg.Key, msg.Data)
	}

}

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

func (dhtNode *DHTNode) Send2(msgId, req, dst, origin, key, data string) {
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
	msg.MsgId = msgId

	udpAddr, err := net.ResolveUDPAddr("udp", dst)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	res, _ := json.Marshal(msg)
	_, err = conn.Write(res) // wat?
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
	// fmt.Println("Responsible: " + key + " N: " + dhtNode.nodeId + ", P: " + dhtNode.Predecessor.nodeId)
	// Checks if the key is equals to my predecessor?
	if dhtNode.Predecessor != nil && dhtNode.Predecessor.nodeId == key {
		return false
	}
	// checks if the key is equals to my key?
	if dhtNode.nodeId == key {
		return true
	}
	if dhtNode.Predecessor != nil && between([]byte(dhtNode.Predecessor.nodeId), []byte(dhtNode.nodeId), []byte(key)) { // if the key is between the nodes Predecessor and itself.
		return true
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

func (node *DHTNode) cleanupRoot() {
	root, err := ioutil.ReadDir(node.path + "root/")
	_ = err
	for _, f := range root {
		bytestring, _ := ioutil.ReadFile(node.path + "clone/" + f.Name())
		if bytestring != nil {
			err := os.Remove(node.path + "root/" + f.Name())
			_ = err
		}
	}
}

func (node *DHTNode) joinResponse(msg *DHTMsg) {
	node.Predecessor = nil
	node.Successor = MakeDHTNode(&msg.Key, msg.Src)
	node.Send("notify", node.Successor.BindAddress, "", "", "")
}

func (node *DHTNode) stabilizeData() {

	files, err := ioutil.ReadDir(node.path + "root/")
	_ = err
	for _, f := range files {
		hashedvalue := generateNodeId(f.Name())
		bytestring, err2 := ioutil.ReadFile(node.path + "root/" + f.Name())
		_ = err2
		node.Send("upload", node.Predecessor.BindAddress, "", hashedvalue, f.Name()+";"+string(bytestring))
		// Send the file. and delete it locally (OR) send it to its new clone folder?
	}
	time.Sleep(50 * time.Duration(len(files)) * time.Millisecond)
	node.Send("cleanupRoot", node.Successor.BindAddress, "", "", "")
	// Clones all the nodes data to it's successor
}

func (node *DHTNode) cloneData() {
	files, err := ioutil.ReadDir(node.path + "root/")
	_ = err
	for _, f := range files {
		hashedvalue := generateNodeId(f.Name())
		// fmt.Println("predecessor is responsible for the file")
		bytestring, err2 := ioutil.ReadFile(node.path + "root/" + f.Name())
		_ = err2
		node.Send("uploadForced", node.Successor.BindAddress, "", hashedvalue, f.Name()+";"+string(bytestring))
	}
}

func (node *DHTNode) deleteForced(msg *DHTMsg) {
	if msg.Src == node.Predecessor.BindAddress {
		path := node.path + "clone/" + msg.Data
		err := os.Remove(path)
		_ = err
	} else {
		//Not correct predecessor, should never happen
		node.Send("stabilize", node.Predecessor.BindAddress, "", "", "")
	}
}

func (node *DHTNode) deleteFile(msg *DHTMsg) {
	if node.responsible(msg.Key) {
		// This code is only accessed if you contact the correct node directly.
		path := node.path + "root/" + msg.Data
		// filedata := []byte(msg.Data[1])
		err := os.Remove(path)
		if err != nil {
			delete(node.fileMap, msg.Data)
			node.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", "", err.Error())
			return
		}
		node.Send("deleteForced", node.Successor.BindAddress, "", msg.Key, msg.Data)
		node.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", msg.Key, msg.Data)
		return
	} else {
		node.lookup(msg)
	}
}

func (node *DHTNode) deleteResponse(msg *DHTMsg) {
	w := node.hashMap[msg.MsgId] // Om w inte är null så är det responsewritern
	if w != nil {
		w.WriteHeader(302)
		fmt.Fprint(w, msg.Src+"/"+msg.Key+"\n"+msg.Data)
		delete(node.hashMap, msg.MsgId)
	} else {
		Warnln("Got fetchResponse but no message in hashmap?")
	}
}

func (node *DHTNode) updateForced(msg *DHTMsg) {
	if msg.Src == node.Predecessor.BindAddress {
		data := strings.Split(msg.Data, ";;")
		path := node.path + "clone/" + data[0]
		err := ioutil.WriteFile(path, []byte(data[1]), 0644)
		_ = err
	} else {
		//Not correct predecessor, should never happen
		node.Send("stabilize", node.Predecessor.BindAddress, "", "", "")
	}
}

func (node *DHTNode) update(msg *DHTMsg) {
	if node.responsible(msg.Key) {
		// This code is only accessed if you contact the correct node directly.
		data := strings.Split(msg.Data, ";;")
		path := node.path + "root/" + data[0]
		// filedata := []byte(msg.Data[1])
		err := ioutil.WriteFile(path, []byte(data[1]), 0644)
		if err != nil || len(data) < 2 {
			node.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", "", err.Error())
			return
		}
		node.Send("updateForced", node.Successor.BindAddress, "", msg.Key, msg.Data)
		node.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", data[0], msg.Data)
	} else {
		node.lookup(msg)
	}
}

func (node *DHTNode) updateResponse(msg *DHTMsg) {
	w := node.hashMap[msg.MsgId] // Om w inte är null så är det responsewritern
	if w != nil {
		w.WriteHeader(302)
		fmt.Fprint(w, msg.Src+"/"+msg.Key+"\n"+msg.Data)
		delete(node.hashMap, msg.MsgId)
	} else {
		Warnln("Got fetchResponse but no message in hashmap?")
	}
}

func (node *DHTNode) fetch(msg *DHTMsg) {
	if node.responsible(msg.Key) {
		// This code is only accessed if you contact the correct node directly.
		path := node.path + "root/" + msg.Data
		bytestring, err := ioutil.ReadFile(path)
		if err == nil {
			node.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", msg.Data, string(bytestring))
		} else {
			node.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", msg.Data, err.Error())
		}
	} else {
		node.lookup(msg)
	}
}

func (node *DHTNode) fetchResponse(msg *DHTMsg) {
	w := node.hashMap[msg.MsgId] // Om w inte är null så är det responsewritern
	if w != nil {
		w.WriteHeader(302)
		fmt.Fprint(w, msg.Src+"/"+msg.Key+"\n"+msg.Data)
		delete(node.hashMap, msg.MsgId)
	} else {
		Warnln("Got fetchResponse but no message in hashmap?")
	}
}

func (node *DHTNode) uploadHandler(msg *DHTMsg) {
	// fmt.Println("\n", msg)
	// fmt.Println("\n -> " + node.BindAddress)
	// hashedValue := generateNodeId(msg.Key)
	if node.responsible(msg.Key) {
		// This code is only accessed if you contact the correct node directly.
		data := strings.Split(msg.Data, ";")
		filedata := ""
		for _, s := range data[1:] {
			filedata = filedata + s
		}
		node.upload(node.path+"root/", data[0], filedata)
		fmt.Println("\n Filename: " + data[0] + "\n" + data[1])
		node.Send("uploadForced", node.Successor.BindAddress, node.BindAddress, msg.Key, msg.Data)
		if msg.Origin != node.Successor.BindAddress || msg.Origin == node.BindAddress {
			node.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", msg.Key, msg.Data)
		}
	} else {
		if msg.Origin != node.Successor.BindAddress {
			node.lookup(msg)
		}
	}
}

func (node *DHTNode) uploadResponse(msg *DHTMsg) {
	w := node.hashMap[msg.MsgId] // Om w inte är null så är det responsewritern
	if w != nil {
		// fmt.Println("\nSending upload response!")
		w.WriteHeader(302)
		data := strings.Split(msg.Data, ";")
		fmt.Fprint(w, "Uploaded on address "+msg.Src+"/"+data[0]+"\n")
		// http.Redirect(w, http.Get)
		delete(node.hashMap, msg.MsgId)
	} else {
		Warnln("Got uploadResponse but no message in hashmap?")
	}
}

func (node *DHTNode) uploadForced(msg *DHTMsg) {
	fmt.Println("Node: " + node.BindAddress + " forced upload")
	if msg.Src == node.Predecessor.BindAddress {
		data := strings.Split(msg.Data, ";")
		filedata := ""
		for _, s := range data[1:] {
			filedata = filedata + s
		}
		node.upload(node.path+"root/", data[0], filedata)
		// path+clone/ represents the cloned data path.
		node.upload(node.path+"clone/", data[0], data[1])
		// node.upload(msg.Key, msg.Data)
	} else {
		//Not correct predecessor, should never happen
		node.Send("stabilize", node.Predecessor.BindAddress, "", "", "")
	}
}

func (node *DHTNode) notifyResponse(msg *DHTMsg) {
	// Warnln(node.BindAddress + " gets Successor " + msg.Src)
	node.Successor = MakeDHTNode(&msg.Key, msg.Src)
	// if node.Predecessor == nil {
	// node.Predecessor = node.Successor
	// }
	// node.Successor = MakeDHTNode(&msg.Key, msg.Src)
}

func (dhtNode *DHTNode) getPredecessor(msg *DHTMsg) {
	// fmt.Println("src " + msg.Src + " dst : " + dhtNode.BindAddress)
	if dhtNode.Predecessor == nil {
		dhtNode.Send("StabilizeResponse", msg.Origin, "", "", "")
	} else {
		dhtNode.Send("StabilizeResponse", msg.Origin, "", dhtNode.Predecessor.nodeId, dhtNode.Predecessor.BindAddress)
	}

}

func (dhtNode *DHTNode) StabilizeResponse(msg *DHTMsg) {
	// dhtNode.lastStab = ""
	// fmt.Println(msg)
	/*if dhtNode.Predecessor == nil {
		fmt.Println(dhtNode.BindAddress + " suc: " + dhtNode.Successor.BindAddress + " has Predecessor: " + msg.Data)
	}*/
	// fmt.Println("src " + msg.Src + " dst : " + dhtNode.BindAddress)
	// src := dhtNode.BindAddress
	// key := strings.Split(msg.Data, ";")
	if (between([]byte(dhtNode.nodeId), []byte(dhtNode.Successor.nodeId), []byte(msg.Key)) && dhtNode.nodeId != msg.Key) || msg.Key == "" {

		// temp := strings.Split(msg.Data, ";")
		dhtNode.Successor.nodeId = msg.Key
		dhtNode.Successor.BindAddress = msg.Data

	} else {
		// fmt.Println(node0)
		dhtNode.Send("getPredecessor", msg.Data, "", "", "")
	}
	dhtNode.Send("notify", dhtNode.Successor.BindAddress, "", "", "")
	// dhtNode.Send("notify", dhtNode.Successor.BindAddress, dhtNode.Successor.nodeId+";"+dhtNode.Successor.BindAddress)

}
func (dhtNode *DHTNode) stabilize(msg *DHTMsg) {
	// fmt.Println(dhtNode.BindAddress + " stabilize, Successor : " + dhtNode.Successor.BindAddress)
	dhtNode.Send("getPredecessor", dhtNode.Successor.BindAddress, "", "", "")
	// dhtNode.lastStab = dhtNode.Successor.BindAddress
	// time.Sleep(50 * time.Millisecond)
	// if dhtNode.lastStab != "" {
	// redo the stabilize but for the next Successor in the succerlist.
	// }
}

func (dhtNode *DHTNode) notify(msg *DHTMsg) {
	if dhtNode.Predecessor != nil && dhtNode.Predecessor.nodeId == msg.Key {
		// do nothing
	} else if dhtNode.Predecessor == nil || between([]byte(dhtNode.Predecessor.nodeId), []byte(dhtNode.nodeId), []byte(msg.Key)) {

		// fmt.Println(dhtNode.nodeId + ".Notify(" + msg.Key + ")")

		temp := dhtNode.Predecessor
		dhtNode.Predecessor = MakeDHTNode(&msg.Key, msg.Src)
		if dhtNode.Successor == nil {
			dhtNode.Successor = dhtNode.Predecessor
		}

		dhtNode.Send("notifyResponse", dhtNode.Predecessor.BindAddress, "", "", "")
		// Sends a response to the old predecessor. if it has beome changed.
		if temp != nil {
			// IF THERE WAS A CHANGE BETWEEN PREDECESSOR AND SUCH.
			if temp.nodeId != dhtNode.Predecessor.nodeId {
				// Triggers the step by step data stabilization.
				dhtNode.Send("newPredecessorEvent", dhtNode.BindAddress, "", "", "")
				// Informs the old predecessor that it should stabilize to find its new successor.
				dhtNode.Send("PredQueryResponse", temp.BindAddress, "", "", dhtNode.Predecessor.nodeId+";"+dhtNode.Predecessor.BindAddress)
			}

		}
		// temp := strings.Split(msg.Data, ";")
		// dhtNode.Predecessor.nodeId = temp[0]
		// dhtNode.Predecessor.BindAddress = temp[1]
		// Notice(dhtNode.Predecessor.nodeId + "\t" + dhtNode.nodeId + "\t" + dhtNode.Successor.nodeId)
	}
}

func (node *DHTNode) newPredecessor() {
	files, err := ioutil.ReadDir(node.path + "clone/")
	_ = err
	for _, f := range files {
		bytestring, err2 := ioutil.ReadFile(node.path + "clone/" + f.Name()) // Reads the file
		_ = err2
		node.Send("cloneReplication", node.Predecessor.BindAddress, "", "", f.Name()+";"+string(bytestring))
		// Send the file. and delete it locally (OR) send it to its new clone folder?

	}
	time.Sleep(50 * time.Duration(len(files)) * time.Millisecond)
	numberOfFiles := "0"
	if files != nil {
		numberOfFiles = strconv.Itoa(len(files))
	}
	node.Send("cloneReplicationEOF", node.Predecessor.BindAddress, "", "", numberOfFiles)
	// Clones all the nodes data to it's successor
}

func (node *DHTNode) cloneReplication(msg *DHTMsg) {
	data := strings.Split(msg.Data, ";")
	filedata := ""
	for _, s := range data[1:] {
		filedata = filedata + s
	}
	node.upload(node.path+"clone/", data[0], filedata)
}

func (node *DHTNode) cloneReplicationEOF(msg *DHTMsg) {
	files, _ := ioutil.ReadDir(node.path + "clone/")
	numberOfFiles, _ := strconv.Atoi(msg.Data)
	if files != nil {
		if len(files) != numberOfFiles { // Missing some files. resend all files again.
			node.Send("newPredecessorEvent", node.Successor.BindAddress, "", "", "")
			return
		}
	}
	if node.Predecessor != nil {
		node.Send("stabilizeData", node.Predecessor.BindAddress, "", "", "")
	}
	// Returns that it is done with all the file downloads. now it can clear its clone folder. Potential problem: We should have a check for consistency in this case. (Before removing the folder on the successr, does the numbers still match?)
	node.Send("cleanupClone", node.Successor.BindAddress, "", "", "")
}

func (node *DHTNode) cleanupClone(msg *DHTMsg) {
	os.RemoveAll(node.path + "clone/")
	os.Mkdir(node.path+"clone/", 0755)
	node.Send("stabilizeData", node.BindAddress, "", "", "")
}

func (node *DHTNode) listen() {
	udpAddr, err := net.ResolveUDPAddr("udp", node.BindAddress)
	conn, err := net.ListenUDP("udp", udpAddr)
	node.online = true
	node.Connection = conn
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	dec := json.NewDecoder(conn)
	// fmt.Println("Started listening : " + node.bindAdress)
	// Error("Started listening : " + node.BindAddress + "\n")
	for {
		if node.online {
			msg := DHTMsg{}
			err = dec.Decode(&msg)
			if err != nil {
				fmt.Println(err.Error())
			}
			node.queue <- &msg
		} else {
			break
		}
	}
}

func (node *DHTNode) timerSomething() {
	k := 0
	for {
		if node.Successor != nil {
			// node.Send("notify", node.Successor.BindAddress, "", "", "")
			node.Send("getPredecessor", node.Successor.BindAddress, "", "", "")
			// node.Send("cleanupRoot", node.BindAddress, "", "", "")
			/*node.Send("PredQuery", node.Successor.BindAddress, "", "", node.nodeId+";"+node.BindAddress)*/
		}
		if k == 3 {
			node.Send("fingerSetup", node.BindAddress, "", "", "")
			k = 0
		}
		time.Sleep(1000 * time.Millisecond)
		k++
	}
}
