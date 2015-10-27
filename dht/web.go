package dht

import (
	// Standard library packages
	"fmt"
	//"html/template"
	"net/http"
	// "strconv"
	// "sync"
	"time"
	// Third party packages
	"github.com/julienschmidt/httprouter"
)

func (dhtNode *DHTNode) startweb() {
	fmt.Println("Node #" + dhtNode.nodeId + " , started listening to : " + dhtNode.BindAddress)
	timeoutValue := time.Duration(2000)
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
		dhtNode.Send(msgId, "upload", dhtNode.BindAddress, "", generateNodeId(key), key+";"+value)
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
		dhtNode.Send(msgId, "fetch", dhtNode.BindAddress, "", generateNodeId(key), key)
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
		dhtNode.Send(msgId, "update", dhtNode.BindAddress, "", generateNodeId(key), key+";;"+value)
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
		dhtNode.Send(msgId, "delete", dhtNode.BindAddress, "", generateNodeId(key), key)
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
		// r.ParseForm()
		// key := r.Form["key"][0]
		// value := r.Form["value"][0]
		// dhtNode.put(key, value)
		fmt.Fprintln(w, "Killed node "+dhtNode.BindAddress)
		dhtNode.online = false
		// dhtNode.Send("", "kill", dhtNode.BindAddress, "", "", "")

		// go dhtNode.Connection.Close()
	})

	r.POST("/exit", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// r.ParseForm()
		// key := r.Form["key"][0]
		// value := r.Form["value"][0]
		// dhtNode.put(key, value)
		// fmt.Fprintln(w, "Killed node "+dhtNode.BindAddress)
		// go dhtNode.Connection.Close()
		dhtNode.Send("", "exit", dhtNode.BindAddress, "", "", "")
	})

	r.POST("/restart", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// r.ParseForm()
		// key := r.Form["key"][0]
		// value := r.Form["value"][0]
		// dhtNode.put(key, value)
		fmt.Fprintln(w, "Started node "+dhtNode.BindAddress)
		dhtNode.Send("", "restart", dhtNode.BindAddress, "", "", "")
	})

	r.POST("/stabilizedata", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.ParseForm()
		dhtNode.Send("", "stabilizeData", dhtNode.BindAddress, "", "", "")
		fmt.Fprintln(w, "newPredecessor method")
	})

	r.POST("/clonedata", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.ParseForm()
		dhtNode.Send("", "cloneData", dhtNode.BindAddress, "", "", "")
		fmt.Fprintln(w, "newPredecessor method")
	})

	r.POST("/join/:key", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.ParseForm()
		key := p.ByName("key")
		dhtNode.Send("", "join", key, "", "", "")
		fmt.Fprintln(w, "Joining with node: "+key)
	})

	r.GET("/getData", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.ParseForm()
		dataValue := "\nStatus\t\tPre \t\t Cur \t\t Suc\t\t Fingers\n"
		if dhtNode.online {
			dataValue += "[ONLINE]\t"
		} else {
			dataValue += "[OFFLINE]\t"
		}
		if dhtNode.Predecessor != nil {
			dataValue += dhtNode.Predecessor.BindAddress + "\t"
		} else {
			dataValue += "\tnil\t"
		}
		dataValue += dhtNode.BindAddress + "\t"
		if dhtNode.Successor != nil {
			dataValue += dhtNode.Successor.BindAddress + "\t"
		} else {
			dataValue += "\tnil\t"
		}

		fingers := dhtNode.FingersToString()
		if fingers == "" {
			dataValue += "nil\t"
		} else {
			dataValue += fingers + "\t"
		}
		successorList := dhtNode.SuccessorListToString()
		if successorList == "[ ]" {
			dataValue += "nil\n"
		} else {
			dataValue += successorList + "\n"
		}
		fmt.Fprintln(w, dataValue)
	})

	// Fire up the server
	http.ListenAndServe(dhtNodeIP, r)
}
