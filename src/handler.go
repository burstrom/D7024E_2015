package dht

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

//Check if folder already exists
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

//Uploads
func (dhtNode *DHTNode) upload(key string, value string) {
	//stringlist := strings.Split(msg.Data)
	//key := stringlist[0]
	//value := stringlist[1]

	path := "storage/" + dhtNode.nodeId + "/"
	if !exists(path) {
		os.MkdirAll(path, 0777)
	}
	path = "storage/" + dhtNode.nodeId + "/" + key + ".txt"
	createFile(path, value)
}

func (dhtNode *DHTNode) get(w http.ResponseWriter, key string) {
	path := "storage/" + dhtNode.nodeId + "/" + key + ".txt"
	fmt.Println(path)
	bytestring, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	} else {
		str := string(bytestring)
		//fmt.Println(str)
		fmt.Fprintln(w, str)
	}
}

func (dhtNode *DHTNode) put(key string, value string) {
	path := "storage/" + dhtNode.nodeId + "/"
	if !exists(path) {
		os.MkdirAll(path, 0777)
	}
	path = "storage/" + dhtNode.nodeId + "/" + key + ".txt"
	createFile(path, value)

}

func (dhtNode *DHTNode) delete(key string) {
	path := "storage/" + dhtNode.nodeId + "/" + key + ".txt"
	os.Remove(path)
}

func createFile(path string, value string) {
	payload := []byte(value)
	fmt.Println("Data: " + value)
	fmt.Println("Path: " + path)
	err := ioutil.WriteFile(path, payload, 0777)
	check(err)
	//fmt.Print(...)
}

//Error handler
func check(e error) {
	if e != nil {
		fmt.Println(e)
	} else {
		fmt.Print("Successfully created file")
	}
}
