package dht

type Task struct {
	node   *DHTNode
	method string
	msg    *DHTMsg
}

type Taskqueue struct {
	queue chan *Task
}

func (tQueue *Taskqueue) queueTask(node *DHTNode, method string) {
	task := new(Task)
	task.node = node
	task.method = method
	tQueue.queue <- task
}

func makeTasker() *Taskqueue {
	tqueue := new(Taskqueue)
	tqueue.queue = make(chan *Task)
	return tqueue
}

/*
func (t *Taskqueue) taskHandler() {
	for {
		select {
		case msg := <-t.queue:
			switch msg.Req {
			case "lookupResponse":
				// vNode := makeVNode(&msg.Key, msg.Data)
				Notice(node.nodeId + " Response: " + msg.Data + ", from " + msg.Key + "\n")
			case "join":
				node.joinRing(msg)
			case "update":
				node.updateNode(msg)
			case "lookup":
				node.lookup(msg)
			case "fingerQuery":
				node.fingerQuery(msg)
			case "fingerSetup":
				node.setupFingers()
			case "fingerResponse":
				node.fingerResponse(msg)
			case "printAll":
				fmt.Println(node.predecessor.nodeId + "\t" + node.nodeId + "\t" + node.successor.nodeId)
				node.printQuery(msg)
			}

		}
	}
}
*/
