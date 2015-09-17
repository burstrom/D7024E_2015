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

func CreateMsg(key string, src string, dst string, req string) *DHTMsg {
	dhtMsg := &DHTMsg{}
	dhtMsg.Key = key
	dhtMsg.Src = src
	dhtMsg.Dst = dst
	dhtMsg.Req = req
	return dhtMsg
}

func (transport *Transport) handler() {
	for {
		select {
		case msg := <-transport.queue:
			switch msg.Req {
			case "lookup":
				fmt.Println("Message lookup received!")
				fmt.Println(msg)
			case "join":
				fmt.Println("Message join received")
				fmt.Println(msg)
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

func (transport *Transport) send(msg *DHTMsg) {
	udpAddr, err := net.ResolveUDPAddr("udp", msg.Dst)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	res, _ := json.Marshal(msg)
	_, err = conn.Write(res) // wat?

	// Todo: JSON Marshalling golang

}
