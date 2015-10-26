package dht

import (
// "encoding/json"
// "fmt"
// "io/ioutil"
// "net"
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
			if node.online {
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
				case "deleteForced":
					node.deleteForced(msg)

				// Deletes the file if it exists, if it doesn't exist it returns the error as a "deleteResponse"
				case "delete":
					node.deleteFile(msg)

				// Response for the delete request. returns it to the correct http request.
				case "deleteResponse":
					node.deleteResponse(msg)

				// Update-forced similar to the other forced methods. only runnable by the predecessor of the node
				case "updateForced":
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
				case "uploadForced":
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

				// This method will send all the files in the clone directory to its new predecessor since it is the one who is responsible for that data now. Ending with <EOF> object and a number. which the new predecessor must respond to check if it matches (so that all files has been uploaded correctly)
				case "newPredecessorEvent":
					node.newPredecessor()

				// This method catches all the files which is sent from its successor "newPredecessorEvent"
				case "cloneReplication":
					node.cloneReplication(msg)

				// Last message which checks the amount of files so that all files got uploaded correctly (has a delay of 50ms*amount of files)
				case "cloneReplicationEOF":
					node.cloneReplicationEOF(msg)

				case "cleanupClone":
					node.cleanupClone(msg)

				// Notify means that the msg.key (node) thinks it is this nodes predecessor
				case "notify":
					node.notify(msg)

				// Step 2 in the notify process
				case "notifyResponse":
					node.notifyResponse(msg)

				// Verifies n' immidiate succesor.
				case "stabilize":
					node.stabilize(msg)

				case "heartbeat":
					node.Send(msg.MsgId, "heartBeatResponse", msg.Origin, "", "", "")
				case "heartBeatResponse":
					node.heartBeatResponse(msg)
				// Kills a node permanently, this is created for debugging purposes
				case "kill":
					node.online = false
					if node.Predecessor != nil {
						node.Predecessor.nodeId = ""
						node.Predecessor.BindAddress = ""
					}

					// node.kill()
					// fmt.Println("Kill all connections/threads related to this node?")
				case "exit":
					node.exit()
				case "successorSet":
					node.successorSet(msg)
				case "successorSetResponse":
					node.successorSetResponse(msg)
				}

			} else if msg.Req == "restart" {
				node.online = true
			}

		}
	}
}
