package main

import (
	// Standard library packages
	// "fmt"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	// "sync"
	"time"
	// Third party packages
	// "github.com/julienschmidt/httprouter"
)

func createContainer()

type Page struct {
	Title string
	Body  []byte
}

func createFile(path string, value string) {
	payload := []byte(value)
	err := ioutil.WriteFile(path, payload, 0777)
	_ = err
}

func main() {
	fs := http.FileServer(http.Dir(""))
	out2, err2 := exec.Command("sh", "-c", "docker ps").Output()
	if err2 != nil {
		log.Printf(err2.Error())
	}
	createFile("./data/activeContainers", "Lastupdated:"+time.Now().String()+"\n"+string(out2))
	http.HandleFunc("/createContainer", func(w http.ResponseWriter, r *http.Request) {
		port := r.FormValue("PORT")
		_ = port
		name := r.FormValue("NAME")
		// command := "docker run  -d -i -t -p " + port + ":" + port + " --net=host --name " + name + " burstrom/reposky localhost:" + port
		// log.Println("sh", "-c", "docker run -d -i -t -p "+port+":"+port+" --net=host --name "+name+" burstrom/reposky", "localhost:"+port)
		// cmd := exec.Command("sh -c " + command)
		cmd := exec.Command("sh", "-c", "docker run -d -i -t -p "+port+":"+port+" --net=host --name "+name+" burstrom/reposky localhost:"+port)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.Println(fmt.Sprint(err) + ": " + stderr.String())
			return
		}
		// log.Println("Result: " + out.String())
		log.Printf("CMD output1: %s\n", out)
		if err != nil {
			log.Printf(err.Error())
		}
		out2, err2 := exec.Command("sh", "-c", "docker ps").Output()
		if err2 != nil {
			log.Printf(err2.Error())
		}
		createFile("./data/activeContainers", "Lastupdated:"+time.Now().String()+"\n"+string(out2))
		http.Redirect(w, r, "/", http.StatusMovedPermanently)

	})

	http.HandleFunc("/deleteContainer", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("CONTAINER_ID")
		// command := "docker run  -d -i -t -p " + port + ":" + port + " --net=host --name " + name + " burstrom/reposky localhost:" + port
		// log.Println("sh", "-c", "docker run -d -i -t -p "+port+":"+port+" --net=host --name "+name+" burstrom/reposky", "localhost:"+port)
		// cmd := exec.Command("sh -c " + command)
		cmd := exec.Command("sh", "-c", "docker rm -f "+name)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.Println(fmt.Sprint(err) + ": " + stderr.String())
			return
		}
		// log.Println("Result: " + out.String())
		log.Printf("CMD output1: %s\n", out)
		if err != nil {
			log.Printf(err.Error())
		}
		out2, err2 := exec.Command("sh", "-c", "docker ps").Output()
		if err2 != nil {
			log.Printf(err2.Error())
		}
		createFile("./data/activeContainers", "Lastupdated:"+time.Now().String()+"\n"+string(out2))
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	})
	http.Handle("/", fs)
	log.Println("starting webserver on port 8800")
	go updateOuput()
	http.ListenAndServe(":8800", nil)
}

func updateOuput() {
	for {
		time.Sleep(5000 * time.Millisecond)
		out, err := exec.Command("sh", "-c", "docker ps").Output()
		if err != nil {
			log.Printf(err.Error())
		}
		createFile("./data/activeContainers", "Lastupdated:"+time.Now().String()+"\n"+string(out))

	}
}
