package dht

import (
	"encoding/json"
	"fmt"
	"net"
	//"time" // added time method
)

type DHTMsg struct {
	//Timestamp int64  `json:"time"`
	Key    string `json:"key"` // Key value
	Src    string `json:"src"` // Source of message
	Req    string `json:"req"`
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
	dhtMsg.Req = req
	dhtMsg.Origin = origin
	dhtMsg.Data = data
	return dhtMsg
}

func (node *DHTNode) handler() {
	for {
		select {
		case msg := <-node.queue:
			switch msg.Req {
			case "LookupResponse":
				vNode := makeVNode(&msg.Key, msg.Data)
				fmt.Println(vNode)
				//Notice(node.nodeId + " Response: " + msg.Data + ", from " + msg.Key + "\n")
			case "join":
				node.joinRing(msg)
			case "update":
				node.updateNode(msg)
			case "lookup":
				node.lookup(msg)
			case "fingerQuery":
				//fmt.Println(node.nodeId+":\t", msg)
				if node.responsible(msg.Key) {
					// fmt.Println(msg)
					node.fingerQuery(msg)
				} else {
					// Framtiden köra en accelerated forward?
					node.sendFrwd(msg, node.successor)
				}
				// Returnerna sig själv till source om typen är request. av typen fingerresponse
				// Vid typen response, lägg till den som finger.
			case "fingerResponse":
				node.fingerResponse(msg)
			case "printall":
				fingers := node.FingersToString()
				if msg.Origin != node.bindAddress {
					msg.Data = msg.Data + node.predecessor.nodeId + "\t" + node.nodeId + "\t" + node.successor.nodeId + "\t" + fingers + "\n"
					node.send("printall", node.successor.bindAddress, msg.Origin, msg.Key, msg.Data)

				} else {
					str := "Pre.\tNode\tSucc.\n" + msg.Data + node.predecessor.nodeId + "\t" + node.nodeId + "\t" + node.successor.nodeId + "\t" + fingers + "\n"
					Noticeln(str)
					//fmt.Print(")
					//fmt.Print(msg.Data+"\n"+node.predecessor.nodeId+"\t"+node.nodeId+"\t"+node.successor.nodeId+"\t"+fingers+"\n", "")
				}
			}

		}
	}
}

func (node *DHTNode) listen() {
	udpAddr, err := net.ResolveUDPAddr("udp", node.bindAddress)
	conn, err := net.ListenUDP("udp", udpAddr)
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	dec := json.NewDecoder(conn)
	// fmt.Println("Started listening : " + node.bindAdress)
	Error("Started listening : " + node.bindAddress + "\n")

	for {
		msg := DHTMsg{}
		err = dec.Decode(&msg)
		if err != nil {
			fmt.Println(err.Error())
		}
		node.queue <- &msg
	}
}
