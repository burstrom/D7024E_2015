package main

import (
	"fmt"
	"github.com/burstrom/D7024E_2015/dht"
	"sync"
	"time"
	"runtime"
	"os"
	"bytes"
)

func main() {
	var wg sync.WaitGroup
	
	dht.Notice("\n### TestNet1 Started\n")
	//node := [8]*dht.DHTNode{}
	fmt.Println("Legend:")
	dht.Error("Error ")
	dht.Notice("Notice ")
	dht.Info("Info ")
	dht.Warn("Warn \n")
	node := dht.MakeDHTNode(nil, os.Args[1])
	wg.Add(1)
	go node.StartServer(&wg)
	if (bytes.Compare([]byte(os.Args[1]),[]byte(os.Args[2]))!=0){		
	go node.Send("join", os.Args[2] ,"", "", "")
		}	
	for {
	time.Sleep(time.Second)
	runtime.Gosched()
	go node.PrintAll()	
	}	
}



