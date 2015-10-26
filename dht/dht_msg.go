package dht

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	// "net/http"
	"strings"
	// "time" // added time method
)

type DHTMsg struct {
	//Timestamp int64  `json:"time"`
	MsgId  string `json:"msgid"`
	Key    string `json:"key"`    // Key value
	Src    string `json:"src"`    // Source of message
	Req    string `json:"req"`    //destnation
	Origin string `json:"origin"` // Original Sender of message.
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
	dhtMsg.MsgId = ""
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
			case "cleanupRoot":
				root, err := ioutil.ReadDir(node.path + "root/")
				_ = err
				for _, f := range root {
					bytestring, _ := ioutil.ReadFile(node.path + "clone/" + f.Name())
					if bytestring != nil {
						err := os.Remove(node.path + "root/" + f.Name())
						_ = err
					}
				}
			case "lookupResponse":
				// vNode := makeVNode(&msg.Key, msg.Data)
				// Notice(node.nodeId + " Response: " + msg.Data + ", from " + msg.Key + "\n")
			case "joinResponse":
				node.Predecessor = nil
				node.Successor = MakeDHTNode(&msg.Key, msg.Src)
				node.Send("notify", node.Successor.BindAddress, "", "", "")
			case "join":
				// node.joinRing(msg)
				// msg.Data = "join"
				node.lookup(msg)
			// case "update":
			// node.updateNode(msg)
			case "lookup":
				node.lookup(msg)
			// checks every file in its root, is its predecessor responsible for any of these values? Then it clones all those values.
			case "stabilizeData":

				files, err := ioutil.ReadDir(node.path + "root/")
				_ = err
				for _, f := range files {
					hashedvalue := generateNodeId(f.Name())
					bytestring, err2 := ioutil.ReadFile(node.path + "root/" + f.Name())
					_ = err2
					node.Send("upload", node.Predecessor.BindAddress, "", hashedvalue, f.Name()+";"+string(bytestring))
					// Send the file. and delete it locally (OR) send it to its new clone folder?

					// time.Sleep(50 * time.Millisecond)
				}
				// Clones all the nodes data to it's successor
			case "cloneData":
				files, err := ioutil.ReadDir(node.path + "root/")
				_ = err
				for _, f := range files {
					hashedvalue := generateNodeId(f.Name())
					// fmt.Println("predecessor is responsible for the file")
					bytestring, err2 := ioutil.ReadFile(node.path + "root/" + f.Name())
					_ = err2
					node.Send("upload-forced", node.Successor.BindAddress, "", hashedvalue, f.Name()+";"+string(bytestring))

				}
			case "delete-forced":
				if msg.Src == node.Predecessor.BindAddress {
					path := node.path + "clone/" + msg.Data
					err := os.Remove(path)
					_ = err
				} else {
					//Not correct predecessor, should never happen
					node.Send("stabilize", node.Predecessor.BindAddress, "", "", "")
				}
			case "delete":
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
					node.Send("delete-forced", node.Successor.BindAddress, "", msg.Key, msg.Data)
					node.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", msg.Key, msg.Data)
					return
				} else {
					node.lookup(msg)
				}
			case "deleteResponse":
				w := node.hashMap[msg.MsgId] // Om w inte är null så är det responsewritern
				if w != nil {
					w.WriteHeader(302)
					fmt.Fprint(w, msg.Src+"/"+msg.Key+"\n"+msg.Data)
					delete(node.hashMap, msg.MsgId)
				} else {
					Warnln("Got fetchResponse but no message in hashmap?")
				}
			case "update-forced":
				if msg.Src == node.Predecessor.BindAddress {
					data := strings.Split(msg.Data, ";;")
					path := node.path + "clone/" + data[0]
					err := ioutil.WriteFile(path, []byte(data[1]), 0644)
					_ = err
				} else {
					//Not correct predecessor, should never happen
					node.Send("stabilize", node.Predecessor.BindAddress, "", "", "")
				}
			case "update":
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
					node.Send("update-forced", node.Successor.BindAddress, "", msg.Key, msg.Data)
					node.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", data[0], msg.Data)
				} else {
					node.lookup(msg)
				}
			case "updateResponse":
				w := node.hashMap[msg.MsgId] // Om w inte är null så är det responsewritern
				if w != nil {
					w.WriteHeader(302)
					fmt.Fprint(w, msg.Src+"/"+msg.Key+"\n"+msg.Data)
					delete(node.hashMap, msg.MsgId)
				} else {
					Warnln("Got fetchResponse but no message in hashmap?")
				}
			case "fetch":
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
			case "fetchResponse":
				w := node.hashMap[msg.MsgId] // Om w inte är null så är det responsewritern
				if w != nil {
					w.WriteHeader(302)
					fmt.Fprint(w, msg.Src+"/"+msg.Key+"\n"+msg.Data)
					delete(node.hashMap, msg.MsgId)
				} else {
					Warnln("Got fetchResponse but no message in hashmap?")
				}
			case "upload":
				fmt.Println("\n", msg)
				fmt.Println("\n -> " + node.BindAddress)
				// hashedValue := generateNodeId(msg.Key)
				if node.responsible(msg.Key) {
					// This code is only accessed if you contact the correct node directly.
					data := strings.Split(msg.Data, ";")
					node.upload(node.path+"root/", data[0], data[1])
					fmt.Println("\n Filename: " + data[0] + "\n" + data[1])
					node.Send("upload-forced", node.Successor.BindAddress, node.BindAddress, msg.Key, msg.Data)
					if msg.Origin != node.Successor.BindAddress {
						node.Send2(msg.MsgId, msg.Req+"Response", msg.Origin, "", msg.Key, msg.Data)
					}
				} else {
					if msg.Origin != node.Successor.BindAddress {
						node.lookup(msg)
					}
				}
			case "uploadResponse":
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

			// Replication which is sent from the nodes predecessor.
			// if it is the predecessor it replicates the data
			case "upload-forced":
				fmt.Println("Node: " + node.BindAddress + " forced upload")
				if msg.Src == node.Predecessor.BindAddress {
					data := strings.Split(msg.Data, ";")
					// path+clone/ represents the cloned data path.
					node.upload(node.path+"clone/", data[0], data[1])
					// node.upload(msg.Key, msg.Data)
				} else {
					//Not correct predecessor, should never happen
					node.Send("stabilize", node.Predecessor.BindAddress, "", "", "")
				}
			case "fingerQuery":
				node.fingerQuery(msg)
			case "fingerSetup":
				node.setupFingers()
			case "fingerResponse":
				node.fingerResponse(msg)
			case "printAll":
				// fmt.Println(node.Predecessor.nodeId + "\t" + node.nodeId + "\t" + node.Successor.nodeId)
				node.printQuery(msg)
			case "getPredecessor":
				node.getPredecessor(msg)
			case "StabilizeResponse":
				node.StabilizeResponse(msg)
			case "notify":
				node.notify(msg)
			case "notifyResponse":
				// Warnln(node.BindAddress + " gets Successor " + msg.Src)
				temp := node.Successor
				node.Successor = MakeDHTNode(&msg.Key, msg.Src)
				if node.Successor != temp {
					// node.Send("cloneData", node.BindAddress, "", "", "")
				}

				// if node.Predecessor == nil {
				// node.Predecessor = node.Successor
				// }
				// node.Successor = MakeDHTNode(&msg.Key, msg.Src)
			case "stabilize":
				node.stabilize(msg)
			case "kill":
				fmt.Println("Kill all connections/threads related to this node?")
			}

		}
	}
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
	if dhtNode.Predecessor == nil || between([]byte(dhtNode.Predecessor.nodeId), []byte(dhtNode.nodeId), []byte(msg.Key)) {

		// fmt.Println(dhtNode.nodeId + ".Notify(" + msg.Key + ")")

		temp := dhtNode.Predecessor
		dhtNode.Predecessor = MakeDHTNode(&msg.Key, msg.Src)
		if dhtNode.Successor == nil {
			dhtNode.Successor = dhtNode.Predecessor
		}
		dhtNode.Send("notifyResponse", dhtNode.Predecessor.BindAddress, "", "", "")

		// time.Sleep(50 * time.Millisecond)
		if temp != nil {
			dhtNode.Send("PredQueryResponse", temp.BindAddress, "", "", dhtNode.Predecessor.nodeId+";"+dhtNode.Predecessor.BindAddress)
		}
		// temp := strings.Split(msg.Data, ";")
		// dhtNode.Predecessor.nodeId = temp[0]
		// dhtNode.Predecessor.BindAddress = temp[1]
		// Notice(dhtNode.Predecessor.nodeId + "\t" + dhtNode.nodeId + "\t" + dhtNode.Successor.nodeId)
	}
	return
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
