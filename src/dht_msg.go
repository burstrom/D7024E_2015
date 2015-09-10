package dht

import ("net"
	"encoding/json"
	"fmt"
)

type DHTMsg struct {
	key string `json:"key"`
	src string `json:"src"`
	dst string `json:"dst"`
	// Other?
	stopNode string `json:"stopNode"`

	bytes string `json:"bytes"`
}

type Transport struct {
	node *DHTNode
	bindAdress string
	queue chan *DHTMsg
}

func CreateTransport(dhtNode *DHTNode, bindAdress string) *Transport{
	transport := &Transport{}
	transport.bindAdress = bindAdress
	transport.node = dhtNode
	return transport
}

func CreateMsg(key string, src string, dst string, bytes string) *DHTMsg{
	dhtMsg := &DHTMsg{}
	dhtMsg.key = key
	dhtMsg.src = src
	dhtMsg.dst = dst
	dhtMsg.bytes = bytes
	return dhtMsg
}

func (transport *Transport) listen() {
	fmt.Println("Started listening : " + transport.bindAdress)
	udpAddr, err := net.ResolveUDPAddr("udp", transport.bindAdress)
	conn, err := net.ListenUDP("udp", udpAddr)
	defer conn.Close()
	if err != nil{
		fmt.Println(err.Error())
	}
	dec := json.NewDecoder(conn)

	for{
		msg := DHTMsg{}
		err = dec.Decode(&msg)
		// We got something?
		if err != nil{
			fmt.Println(err.Error())
		}
		transport.queue <- &msg
		fmt.Println(transport.node.nodeId+":> from:" + msg.src + " to: " + msg.dst)
		//fmt.Println(msg.Bytes + " " + "")
	}
}

func (transport *Transport) send(msg *DHTMsg){
	udpAddr, err := net.ResolveUDPAddr("udp", msg.dst)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()
	if err != nil{
		fmt.Println(err.Error())
	}
	res, _ := json.Marshal(msg)
	_, err = conn.Write(res)

	// Todo: JSON Marshalling golang

}
