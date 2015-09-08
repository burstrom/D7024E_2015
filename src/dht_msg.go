package dht
import ("net"
	"encoding/json"
	"net"
	"fmt"
)

type dhtMsg struct {
	Key string
	Src	string
	Dst string
	// Other?
	Bytes string //?
}

type Transport struct {
	bindAdress string
}

func (transport *Transport) listen() {
	udpAddr, err := net.ResolveUDPAddr("udp", transport.bindAdress)
	conn, err := net.ListenUDP("udp", udpAddr)
	defer conn.Close()
	dec := json.NewDecoder(conn)
	for{
		msg := dhtMsg{}
		err = dec.Decode(&msg)
		// We got something?
	}
}

func (transport *Transport) send(msg *dhtMsg){
	udpAddr, err := net.ResolveUDPAddr("udp", dhtMsg.Dst)

	conn, err := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()

	_, err = conn.Write(msg.Bytes()) // wat?

	// Todo: JSON Marshalling golang

}
