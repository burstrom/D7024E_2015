package dht
import "fmt"

type Contact struct {
	ip   string
	port string
}

// go test -test.run <functionsnamn>

type DHTNode struct {
	nodeId      string
	successor   *DHTNode
	predecessor *DHTNode
	contact     Contact
	fingers [bits]*DHTNode
}

func makeDHTNode(nodeId *string, ip string, port string) *DHTNode {
	dhtNode := new(DHTNode)
	dhtNode.contact.ip = ip
	dhtNode.contact.port = port

	if nodeId == nil {
		genNodeId := generateNodeId()
		dhtNode.nodeId = genNodeId
	} else {
		dhtNode.nodeId = *nodeId
	}

	dhtNode.successor = nil
	dhtNode.predecessor = nil

	return dhtNode
}

func (dhtNode *DHTNode) addToRing(newDHTNode *DHTNode) {
	// Default case, if there is only two nodes but neither one is connected.
	if dhtNode.successor == nil && dhtNode.predecessor == nil{
		// Set the dhtNodes successor to the new node, and change predecessor
		dhtNode.predecessor = newDHTNode
		dhtNode.successor = newDHTNode
		newDHTNode.predecessor = dhtNode
		newDHTNode.successor = dhtNode
	}else {
		if(between([]byte(dhtNode.nodeId),[]byte(dhtNode.successor.nodeId),[]byte(newDHTNode.nodeId))){
			// It should be between dhtNode and dhtNode successor.
			newDHTNode.successor = dhtNode.successor
			newDHTNode.predecessor = dhtNode
			// Update the node1 successor, and node 2 predecessor.
			dhtNode.successor.predecessor = newDHTNode
			dhtNode.successor = newDHTNode
		}else {
			// Recursively call add to ring with next value
			dhtNode.successor.addToRing(newDHTNode)
		}
	}
	stabalizeRing(dhtNode,dhtNode)
}

func (dhtNode *DHTNode) lookup(key string) *DHTNode {
	// If the node is responsible for the given key
	if(dhtNode.responsible(key)){
		return dhtNode
	}
	return dhtNode.successor.lookup(key)
}

func (dhtNode *DHTNode) acceleratedLookupUsingFingers(key string) *DHTNode {
	// TODO
	return dhtNode // XXX This is not correct obviously
}

func (dhtNode *DHTNode) responsible(key string) bool {
	// Is the key the same key as the node?
	if(dhtNode.nodeId == key){
		return true
	}
	if(dhtNode.predecessor.nodeId == key){
		return false
	}
	// Is the key value between
	return between([]byte(dhtNode.predecessor.nodeId),[]byte(dhtNode.nodeId),[]byte(key))
	return false
}

func (dhtNode *DHTNode) printRing() {
	activeNode := dhtNode.successor
	fmt.Print(dhtNode.nodeId + " ")
	for activeNode != dhtNode{
		fmt.Print(activeNode.nodeId + " ")
		activeNode = activeNode.successor
	}
	fmt.Println()
}

func (dhtNode *DHTNode) testCalcFingers(m int, bits int) {
	/* idBytes, _ := hex.DecodeString(dhtNode.nodeId)
	fingerHex, _ := calcFinger(idBytes, m, bits)
	fingerSuccessor := dhtNode.lookup(fingerHex)
	fingerSuccessorBytes, _ := hex.DecodeString(fingerSuccessor.nodeId)
	fmt.Println("successor    " + fingerSuccessor.nodeId)

	dist := distance(idBytes, fingerSuccessorBytes, bits)
	fmt.Println("distance     " + dist.String()) */
}


func (dhtNode *DHTNode) printFingers(){
	fmt.Print("#"+dhtNode.nodeId+ " :> ")
	for k:=0; k < len(dhtNode.fingers); k++{
		fmt.Print(dhtNode.fingers[k].nodeId+" ")
	}
	fmt.Println()
}
// nodeXX.calcFinger
/*func (dhtNode *DHTNode) printFinger(n []byte, k int, m int) (string []byte){



	return dhtNode
}*/