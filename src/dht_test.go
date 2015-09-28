package dht

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNet1(t *testing.T) {
	var wg sync.WaitGroup
	fmt.Println("## Testing started")
	id0 := "00"
	id1 := "01"
	id2 := "02"
	id3 := "03"
	id4 := "04"
	id5 := "05"
	id6 := "06"
	id7 := "07"
	node0 := makeDHTNode(&id0, "localhost", "1111")
	node1 := makeDHTNode(&id1, "localhost", "1112")
	node2 := makeDHTNode(&id2, "localhost", "1113")
	node3 := makeDHTNode(&id3, "localhost", "1114")
	node4 := makeDHTNode(&id4, "localhost", "1115")
	node5 := makeDHTNode(&id5, "localhost", "1116")
	node6 := makeDHTNode(&id6, "localhost", "1117")
	node7 := makeDHTNode(&id7, "localhost", "1118")
	wg.Add(8)
	go node0.startServer(&wg)
	go node1.startServer(&wg)
	go node2.startServer(&wg)
	go node3.startServer(&wg)
	go node4.startServer(&wg)
	go node5.startServer(&wg)
	go node6.startServer(&wg)
	go node7.startServer(&wg)
	wg.Wait()
	go node1.send("join", node0, "", "")
	time.Sleep(400 * time.Millisecond)
	go node2.send("join", node1, "", "")
	time.Sleep(400 * time.Millisecond)
	go node3.send("join", node1, "", "")
	time.Sleep(400 * time.Millisecond)
	go node4.send("join", node2, "", "")
	time.Sleep(400 * time.Millisecond)
	go node5.send("join", node2, "", "")
	time.Sleep(400 * time.Millisecond)
	go node6.send("join", node3, "", "")
	time.Sleep(400 * time.Millisecond)
	go node7.send("join", node3, "", "")
	time.Sleep(400 * time.Millisecond)
	go node1.printAll()
	time.Sleep(400 * time.Millisecond)

	//node1.setupFingers()
	go node1.setupFingers()
	time.Sleep(15000 * time.Millisecond)
	//go node5.transport.send(CreateMsg("bajs", node5.transport.bindAdress, "localhost:1112", "join"))

	//key string, src string, dst string, bytes string
}
