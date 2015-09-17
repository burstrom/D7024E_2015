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
	Req string `json:"req"`
	Opt string `json:"opt"`
	//Bytes string `json:"bytes"`
}

type Transport struct {
	node       *DHTNode
	bindAdress string
	queue      chan *DHTMsg
}

func CreateTransport(dhtNode *DHTNode, bindAdress string) *Transport {
	transport := &Transport{}
	transport.bindAdress = bindAdress
	transport.node = dhtNode
	transport.queue = make(chan *DHTMsg)
	go transport.handler()
	return transport
}

func CreateMsg(key string, src string, dst string, req string, opt string) *DHTMsg {
	dhtMsg := &DHTMsg{}
	dhtMsg.Key = key
	dhtMsg.Src = src
	dhtMsg.Dst = dst
	dhtMsg.Req = req
	dhtMsg.Opt = opt
	return dhtMsg
}

func (transport *Transport) handler() {
	for {
		select {
		case msg := <-transport.queue:
			switch msg.Req {
			case "update":
				fmt.Println("\\e[96m[UPDATE]\\e[39m received!")
				fmt.Println(msg)
			case "join":

				//fmt.Println("[" + transport.node.nodeId + "]: <-[JOIN]:[" + msg.Src + "]")
				//fmt.Println("SRC:" + msg.Src)
				//fmt.Println("DST:" + msg.Dst)
				//fmt.Println(msg)
				transport.node.joinRing(msg)
				//transport.node.addToRing(newDHTNode)

			}
			//fmt.Println(msg.Dst + " DEST")
		}
	}
}

func (transport *Transport) listen() {
	udpAddr, err := net.ResolveUDPAddr("udp", transport.bindAdress)
	conn, err := net.ListenUDP("udp", udpAddr)
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	dec := json.NewDecoder(conn)
	fmt.Println("Started listening : " + transport.bindAdress)
	for {
		msg := DHTMsg{}
		err = dec.Decode(&msg)
		// We got something?
		if err != nil {
			fmt.Println(err.Error())
		}
		//fmt.Println(msg)
		//fmt.Println(transport.node.nodeId + ":> from:" + msg.Src + " to: " + msg.Dst)
		transport.queue <- &msg
		//fmt.Println(msg.Bytes + " " + "")
	}
}

func (dhtNode *DHTNode) send(req string, dstNode *DHTNode, opt string) {

	msg := CreateMsg(dhtNode.nodeId, dhtNode.transport.bindAdress, dstNode.transport.bindAdress, req, opt)
	udpAddr, err := net.ResolveUDPAddr("udp", msg.Dst)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	res, _ := json.Marshal(msg)
	_, err = conn.Write(res) // wat?

}
