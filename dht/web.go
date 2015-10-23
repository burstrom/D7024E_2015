package dht

import (
	// Standard library packages
	"fmt"
	//"html/template"
	"net/http"

	// Third party packages
	"github.com/julienschmidt/httprouter"
)

func (dhtNode *DHTNode) startweb() {
	fmt.Println("Node #" + dhtNode.nodeId + " , started listening to : " + dhtNode.BindAddress)
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
		dhtNode.Send("upload", dhtNode.BindAddress, "", key, value)
	})

	r.GET("/storage/:key", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// key := p.ByName("key")
		// dhtNode.get(w, key)
	})

	r.PUT("/storage/:key", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r.ParseForm()
		// key := p.ByName("key")
		// value := r.Form["value"][0]
		//dhtNode.lookup(CreateMsg("lookup", dhtNode.nodeId, dhtNode.BindAddress, generateNodeId(key), ""))
		fmt.Fprint(w, "This should return the payload of the key!\n")
		// dhtNode.hashMap(key[value])
		// for {
		// select {
		// case httpresponse := <-dhtNode.httpResponse:
		// fmt.Println(httpresponse)
		// stringlist := string.Split(httpresponse, ",")
		// }
		// }
	})

	r.DELETE("/storage/:key", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		key := p.ByName("key")
		dhtNode.delete(key)
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
