package dht

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNet1(t *testing.T) {
	var wg sync.WaitGroup
	// tQueue := makeTasker()
	Notice("\n### TestNet1 Started\n")
	node := [8]*DHTNode{}
	fmt.Println("Legend:")
	Error("Error ")
	Notice("Notice ")
	Info("Info ")
	Warn("Warn \n")

	fmt.Print("")
	id0 := "00"
	id1 := "01"
	id2 := "02"
	id3 := "03"
	id4 := "04"
	id5 := "05"
	id6 := "06"
	id7 := "07"
	node[0] = makeDHTNode(&id0, "localhost:1110")
	node[1] = makeDHTNode(&id1, "localhost:1111")
	node[2] = makeDHTNode(&id2, "localhost:1112")
	node[3] = makeDHTNode(&id3, "localhost:1113")
	node[4] = makeDHTNode(&id4, "localhost:1114")
	node[5] = makeDHTNode(&id5, "localhost:1115")
	node[6] = makeDHTNode(&id6, "localhost:1116")
	node[7] = makeDHTNode(&id7, "localhost:1117")
	wg.Add(8)
	go node[0].startServer(&wg)
	go node[1].startServer(&wg)
	go node[2].startServer(&wg)
	go node[3].startServer(&wg)
	go node[4].startServer(&wg)
	go node[5].startServer(&wg)
	go node[6].startServer(&wg)
	go node[7].startServer(&wg)
	wg.Wait()

	// go node[0].startweb()
	// go node[1].startweb()
	// go node[2].startweb()
	// go node[3].startweb()
	// go node[4].startweb()
	// go node[5].startweb()
	// go node[6].startweb()
	// go node[7].startweb()

	fmt.Println("All nodes joining eachother!")
	node[1].send("join", node[0].bindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[2].send("join", node[0].bindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[3].send("join", node[0].bindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[4].send("join", node[0].bindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[5].send("join", node[0].bindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[6].send("join", node[0].bindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[7].send("join", node[0].bindAddress, "", "", "")
	time.Sleep(450 * time.Millisecond)
	// go node[1].printAll()
	time.Sleep(200 * time.Millisecond)

	// Iterates and updates each nodes fingers.
	fmt.Println("All nodes joined!")
	// go node[0].setupFingers()
	// fmt.Print("\nSetting up fingers: ")
	for i := 0; i < 8; i++ {
		node[i].fingerResponses = 0
		// fmt.Println(node[i].nodeId+"  successor: ", node[i].successor)
		// node[i].queue <- CreateMsg("fingerSetup", node[i].bindAddress, node[i].bindAddress, "", "")
		// time.Sleep(50 * time.Millisecond)
	}
	// node[1].stabilize()

	// fmt.Println("\nFinger setup completed")
	// fmt.Println("Waiting for fingers to setup correctly..")
	// fmt.Println(node[0].fingerResponses, ": Responses so far")
	time.Sleep(900 * time.Millisecond)
	// fmt.Println(node[0].fingerResponses, ": Responses so far")
	// time.Sleep(400 * time.Millisecond)
	// fmt.Println(node[0].fingerResponses, ": Responses so far")
	// fmt.Println(node[0].predecessor.bindAddress + " - " + node[0].bindAddress + " - " + node[0].successor.bindAddress)
	// fmt.Println(node[0].FingersToString())

	time.Sleep(50 * time.Millisecond)
	//
	// node[0].printAll()
	// time.Sleep(50 * time.Millisecond)
	// fmt.Println("Trying a lookup!")
	// time.Sleep(time.Millisecond * 5000)

	// time.Sleep(15000 * time.Millisecond)
	// go node[0].printAll()
	// time.Sleep(5000 * time.Millisecond)
	// node[1].send("lookup", node[1].bindAddress, "", "50", "")
	//Keeps the web servers alive
	for {
		Error("\nPre \t\t Cur \t\t Suc\t\t\t")
		Warnln("Fingers")
		for i := 0; i < 8; i++ {
			if node[i].predecessor == nil && node[i].successor == nil {
				Notice("\tnil\t- " + node[i].bindAddress + " - nil\t-")
			} else if node[i].predecessor == nil {
				Notice("\tnil\t- " + node[i].bindAddress + " - " + node[i].successor.bindAddress)
			} else if node[i].successor == nil {
				Notice(node[i].predecessor.bindAddress + "\t- " + node[i].bindAddress + " -\t nil")
			} else {
				Notice(node[i].predecessor.bindAddress + "\t- " + node[i].bindAddress + " - " + node[i].successor.bindAddress)
			}
			Warn("\t" + node[i].FingersToString() + "\n")
			time.Sleep(20 * time.Millisecond)
		}
		time.Sleep(time.Millisecond * 4000)
		// node[0].printAll()
	}

	//key string, src string, dst string, bytes string
}
