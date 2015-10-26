package dht

import (
	"encoding/json"
	"fmt"
	// "io/ioutil"
	"net"
	// "os"
	// "net/http"
	// "strings"
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

	return dhtMsg
}

// Main handler which makes every msg synchronously, so we don't get problems with many operations using the same functions
func (node *DHTNode) handler() {
	for {
		select {
		case msg := <-node.queue:
			switch msg.Req {
			// Cleanup Root removes all files in the root file which also exist in the clone dir.
			case "cleanupRoot":
				node.cleanupRoot()

			// This response is sent from the node who is responsible for the key, then this node tries to join in front of this node.
			case "joinResponse":
				node.joinResponse(msg)

			// Node received a join request from a node who wants to join the network.
			case "join":
				// node.joinRing(msg)
				// msg.Data = "join"
				node.lookup(msg)

			// Sends a lookup and the node who is responsible for the msg.Key value returns a lookupResponse
			case "lookup":
				node.lookup(msg)

			// tries to upload all files to it's predecessor. Which means all files that are supposed to be in the /clone folder will be cloned automatically.
			case "stabilizeData":
				node.stabilizeData()

			// clones all data to the nodes successor (the nodes root folder becomes uploaded to the successors clone folder)
			case "cloneData":
				node.cloneData()

			// delete-forced is a method used by the predecessor, if its sent from another node then it doesn't do anything (well it returns a stabilize)
			case "delete-forced":
				node.deleteForced(msg)

			// Deletes the file if it exists, if it doesn't exist it returns the error as a "deleteResponse"
			case "delete":
				node.deleteFile(msg)

			// Response for the delete request. returns it to the correct http request.
			case "deleteResponse":
				node.deleteResponse(msg)

			// Update-forced similar to the other forced methods. only runnable by the predecessor of the node
			case "update-forced":
				node.updateForced(msg)

			// update/patch file, the node gets a request to update a files data (if it exists)
			case "update":
				node.update(msg)

			// the response from the node who owns the file with data if it succeeded or failed
			case "updateResponse":
				node.updateResponse(msg)

			// Fetch the data which corresponds to the key + filename (which is inside the data)
			case "fetch":
				node.fetch(msg)

			// Response from the node who is responsible for the key, and if the file exists its been sent in the data. This node then forwards the message back to the matching msgId
			case "fetchResponse":
				node.fetchResponse(msg)

			// Uploads a file to the correct node.
			case "upload":
				node.uploadHandler(msg)

			// Response from the node which is responsible for the key, then it contains if it was successful or an error if there was an error.
			case "uploadResponse":
				node.uploadResponse(msg)

			// Replication which is sent from the nodes predecessor.
			// if it is the predecessor it replicates the data
			case "upload-forced":
				node.uploadForced(msg)

			// This is used when updating fingers, fingerquery checks if the node is respnsible for given key or not.
			case "fingerQuery":
				node.fingerQuery(msg)

			// Setupfinger method which sends fingerqueries for all keys the finger wants to be responsible for
			case "fingerSetup":
				node.setupFingers()

			// Response from the nodes which is responsible for the given value k.
			case "fingerResponse":
				node.fingerResponse(msg)

			// PrintAll methods iterates through the ring to give info how the ring is connected.
			case "printAll":
				node.printQuery(msg)

			// It sends the nodes predecessor to match it when doing the notify/stabilize.
			case "getPredecessor":
				node.getPredecessor(msg)

			// StabilizeResponse is a part of the check to verify if n' immidieate successor and tell the successor about n
			case "StabilizeResponse":
				node.StabilizeResponse(msg)

			// Notify means that the msg.key (node) thinks it is this nodes predecessor
			case "notify":
				node.notify(msg)

			// Step 2 in the notify process
			case "notifyResponse":
				node.notifyResponse(msg)

			// Verifies n' immidiate succesor.
			case "stabilize":
				node.stabilize(msg)

			// Kills a node permanently, this is created for debugging purposes
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
