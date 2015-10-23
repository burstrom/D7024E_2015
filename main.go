package dht

import (
	"fmt"
	"github.com/burstrom/D7024E_2015/dht"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	// tQueue := makeTasker()
	dht.Notice("\n### TestNet1 Started\n")
	node := [8]*dht.DHTNode{}
	fmt.Println("Legend:")
	dht.Error("Error ")
	dht.Notice("Notice ")
	dht.Info("Info ")
	dht.Warn("Warn \n")

	fmt.Print("")
	id0 := "00"
	id1 := "01"
	id2 := "02"
	id3 := "03"
	id4 := "04"
	id5 := "05"
	id6 := "06"
	id7 := "07"
	node[0] = dht.MakeDHTNode(&id0, "localhost:1110")
	node[1] = dht.MakeDHTNode(&id1, "localhost:1111")
	node[2] = dht.MakeDHTNode(&id2, "localhost:1112")
	node[3] = dht.MakeDHTNode(&id3, "localhost:1113")
	node[4] = dht.MakeDHTNode(&id4, "localhost:1114")
	node[5] = dht.MakeDHTNode(&id5, "localhost:1115")
	node[6] = dht.MakeDHTNode(&id6, "localhost:1116")
	node[7] = dht.MakeDHTNode(&id7, "localhost:1117")
	wg.Add(8)
	go node[0].StartServer(&wg)
	go node[1].StartServer(&wg)
	go node[2].StartServer(&wg)
	go node[3].StartServer(&wg)
	go node[4].StartServer(&wg)
	go node[5].StartServer(&wg)
	go node[6].StartServer(&wg)
	go node[7].StartServer(&wg)
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
	node[1].Send("join", node[0].BindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[2].Send("join", node[0].BindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[3].Send("join", node[0].BindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[4].Send("join", node[0].BindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[5].Send("join", node[0].BindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[6].Send("join", node[0].BindAddress, "", "", "")
	time.Sleep(150 * time.Millisecond)
	node[7].Send("join", node[0].BindAddress, "", "", "")
	time.Sleep(450 * time.Millisecond)
	// go node[1].printAll()
	time.Sleep(200 * time.Millisecond)

	// Iterates and updates each nodes fingers.
	fmt.Println("All nodes joined!")
	// go node[0].setupFingers()
	// fmt.Print("\nSetting up fingers: ")
	for i := 0; i < 8; i++ {
		node[i].FingerResponses = 0
		// fmt.Println(node[i].nodeId+"  Successor: ", node[i].Successor)
		// node[i].queue <- CreateMsg("fingerSetup", node[i].BindAddress, node[i].BindAddress, "", "")
		// time.Sleep(50 * time.Millisecond)
	}
	// node[1].stabilize()

	// fmt.Println("\nFinger setup completed")
	// fmt.Println("Waiting for fingers to setup correctly..")
	// fmt.Println(node[0].FingerResponses, ": Responses so far")
	time.Sleep(900 * time.Millisecond)
	// fmt.Println(node[0].FingerResponses, ": Responses so far")
	// time.Sleep(400 * time.Millisecond)
	// fmt.Println(node[0].FingerResponses, ": Responses so far")
	// fmt.Println(node[0].Predecessor.BindAddress + " - " + node[0].BindAddress + " - " + node[0].Successor.BindAddress)
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
	// node[1].Send("lookup", node[1].BindAddress, "", "50", "")
	//Keeps the web servers alive
	for {
		dht.Error("\nPre \t\t Cur \t\t Suc\t\t\t")
		dht.Warnln("Fingers")
		for i := 0; i < 8; i++ {
			if node[i].Predecessor == nil && node[i].Successor == nil {
				dht.Notice("\tnil\t- " + node[i].BindAddress + " - nil\t-")
			} else if node[i].Predecessor == nil {
				dht.Notice("\tnil\t- " + node[i].BindAddress + " - " + node[i].Successor.BindAddress)
			} else if node[i].Successor == nil {
				dht.Notice(node[i].Predecessor.BindAddress + "\t- " + node[i].BindAddress + " -\t nil")
			} else {
				dht.Notice(node[i].Predecessor.BindAddress + "\t- " + node[i].BindAddress + " - " + node[i].Successor.BindAddress)
			}
			dht.Warn("\t" + node[i].FingersToString() + "\n")
			time.Sleep(20 * time.Millisecond)
		}
		time.Sleep(time.Millisecond * 4000)
		// node[0].printAll()
	}

	//key string, src string, dst string, bytes string
}
