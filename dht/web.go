package dht

import (
	// Standard library packages
	"fmt"
	//"html/template"
	"net/http"
	// "strconv"
	"time"
	// Third party packages
	"github.com/julienschmidt/httprouter"
)

func (dhtNode *DHTNode) startweb() {
	fmt.Println("Node #" + dhtNode.nodeId + " , started listening to : " + dhtNode.BindAddress)
	timeoutValue := 1000
	// Instantiate a new router
	r := httprouter.New()
	dhtNodeIP := dhtNode.BindAddress
	//r.GET("/", Index)
	r.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprint(w, "Welcome to "+dhtNodeIP+"!\n")
	})

	r.POST("/storage", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		r.ParseForm()
		key := r.Form["key"][0]
		value := r.Form["value"][0]
		msgId := generateNodeId(time.Now().String() + r.Host)
		dhtNode.Send2(msgId, "upload", dhtNode.BindAddress, "", generateNodeId(key), key+";"+value)
		dhtNode.hashMap[msgId] = w
		time.Sleep(timeoutValue * time.Millisecond)
		// fmt.Println("\n ## " + msgId + " -> " + dhtNode.BindAddress + " with data: " + key + "," + value)
		rw := dhtNode.hashMap[msgId]
		if rw != nil {
			w.WriteHeader(418)
			fmt.Fprint(w, "Couldn't upload file, TIMEOUT")
		}
	})

	r.GET("/storage/:key", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.ParseForm()
		key := p.ByName("key")
		msgId := generateNodeId(time.Now().String() + r.Host)
		dhtNode.Send2(msgId, "fetch", dhtNode.BindAddress, "", generateNodeId(key), key)
		dhtNode.hashMap[msgId] = w
		time.Sleep(timeoutValue * time.Millisecond)
		// fmt.Println("\n ## " + msgId + " -> " + dhtNode.BindAddress + " with data: " + key + "," + value)
		rw := dhtNode.hashMap[msgId]
		if rw != nil {
			w.WriteHeader(418)
			fmt.Fprint(w, "Couldn't get file, TIMEOUT")
		}
	})

	r.PUT("/storage/:key", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.ParseForm()
		key := p.ByName("key")
		value := r.Form["value"][0]
		msgId := generateNodeId(time.Now().String() + r.Host)
		dhtNode.Send2(msgId, "update", dhtNode.BindAddress, "", generateNodeId(key), key+";;"+value)
		dhtNode.hashMap[msgId] = w
		time.Sleep(timeoutValue * time.Millisecond)
		rw := dhtNode.hashMap[msgId]
		if rw != nil {
			w.WriteHeader(418)
			fmt.Fprint(w, "Couldn't update file, TIMEOUT")
		} // fmt.Fprint(w, "This should return the payload of the key!\n")
	})

	r.DELETE("/storage/:key", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.ParseForm()
		key := p.ByName("key")
		msgId := generateNodeId(time.Now().String() + r.Host)
		dhtNode.Send2(msgId, "delete", dhtNode.BindAddress, "", generateNodeId(key), key)
		dhtNode.hashMap[msgId] = w
		time.Sleep(timeoutValue * time.Millisecond)
		// fmt.Println("\n ## " + msgId + " -> " + dhtNode.BindAddress + " with data: " + key + "," + value)
		rw := dhtNode.hashMap[msgId]
		if rw != nil {
			w.WriteHeader(418)
			fmt.Fprint(w, "Couldn't delete file, TIMEOUT")
		}
	})

	r.POST("/kill", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		r.ParseForm()
		key := r.Form["key"][0]
		value := r.Form["value"][0]
		dhtNode.put(key, value)
		fmt.Fprintln(w, "Key:"+key+" Value:"+value)
	})

	// Fire up the server
	http.ListenAndServe(dhtNodeIP, r)
}
