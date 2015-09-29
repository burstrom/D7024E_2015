package dht

import (
	"encoding/json"
	"fmt"
	"net"
)

type DHTMsg struct {
	Key string `json:"key"`
	Src string `json:"src"`
	Dst string `json:"dst"`
	// Other?
	Req  string `json:"req"`
	Opt  string `json:"opt"`
	Data string `json:"data"`
	//Origin string `json:"origin"`
	//Bytes string `json:"bytes"`
}

func CreateMsg(key string, src string, dst string, req string, opt string, data string) *DHTMsg {
	dhtMsg := &DHTMsg{}
	dhtMsg.Key = key
	dhtMsg.Src = src
	dhtMsg.Dst = dst
	dhtMsg.Req = req
	dhtMsg.Opt = opt
	dhtMsg.Data = data
	return dhtMsg
}

func (node *DHTNode) handler() {
	for {
		select {
		case msg := <-node.queue:
			switch msg.Req {
			case "join":
				node.joinRing(msg)
			case "update":
				node.updateNode(msg)
			case "lookupResp":
				//fmt.Println("Response: " + msg.Data)
				//transport.nodest.setupFingers()
			case "fingerQuery":
				//fmt.Println(node.nodeId+":\t", msg)

				if node.responsible(msg.Opt) {
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
			case "printring":
				if msg.Opt != node.nodeId {
					node.send("printring", node.predecessor, msg.Opt, msg.Data+"->"+node.nodeId)
				} else {
					//fmt.Println(msg.Data)
				}
			case "printall":
				fingers := node.FingersToString()
				if msg.Opt != node.nodeId {
					node.send("printall", node.successor, msg.Opt, msg.Data+"\n"+node.predecessor.nodeId+"\t"+node.nodeId+"\t"+node.successor.nodeId+"\t"+fingers)
				} else {
					str := "Pre.\tNode\tSucc.\n" + msg.Data + "\n" + node.predecessor.nodeId + "\t" + node.nodeId + "\t" + node.successor.nodeId + "\t" + fingers + "\n"
					Noticeln(str)
					//fmt.Print(")
					//fmt.Print(msg.Data+"\n"+node.predecessor.nodeId+"\t"+node.nodeId+"\t"+node.successor.nodeId+"\t"+fingers+"\n", "")
				}
			}

		}
	}
}

func (node *DHTNode) listen() {
	udpAddr, err := net.ResolveUDPAddr("udp", node.bindAdress)
	conn, err := net.ListenUDP("udp", udpAddr)
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	dec := json.NewDecoder(conn)
	// fmt.Println("Started listening : " + node.bindAdress)
	Error("Started listening : " + node.bindAdress + "\n")

	for {
		msg := DHTMsg{}
		err = dec.Decode(&msg)
		if err != nil {
			fmt.Println(err.Error())
		}
		node.queue <- &msg
	}
}
