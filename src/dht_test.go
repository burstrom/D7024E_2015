package dht

import (
	"fmt"
	"testing"
)

func TestDHT1(t *testing.T) {
	id0 := "00"
	id1 := "01"
	id2 := "02"
	id3 := "03"
	id4 := "04"
	id5 := "05"
	id6 := "06"
	id7 := "07"

	node0b := makeDHTNode(&id0, "localhost", "1111")
	node1b := makeDHTNode(&id1, "localhost", "1112")
	node2b := makeDHTNode(&id2, "localhost", "1113")
	node3b := makeDHTNode(&id3, "localhost", "1114")
	node4b := makeDHTNode(&id4, "localhost", "1115")
	node5b := makeDHTNode(&id5, "localhost", "1116")
	node6b := makeDHTNode(&id6, "localhost", "1117")
	node7b := makeDHTNode(&id7, "localhost", "1118")

	node0b.addToRing(node1b)
	node1b.addToRing(node2b)
	node1b.addToRing(node3b)
	node1b.addToRing(node4b)
	node4b.addToRing(node5b)
	node3b.addToRing(node6b)
	node3b.addToRing(node7b)

	fmt.Println("-> ring structure")
	node1b.printRing()

	node3b.testCalcFingers(0, 3)
	node3b.testCalcFingers(1, 3)
	node3b.testCalcFingers(2, 3)
	node3b.testCalcFingers(3, 3)
}

func TestDHT2(t *testing.T) {
	node1 := makeDHTNode(nil, "localhost", "1111")
	node2 := makeDHTNode(nil, "localhost", "1112")
	node3 := makeDHTNode(nil, "localhost", "1113")
	node4 := makeDHTNode(nil, "localhost", "1114")
	node5 := makeDHTNode(nil, "localhost", "1115")
	node6 := makeDHTNode(nil, "localhost", "1116")
	node7 := makeDHTNode(nil, "localhost", "1117")
	node8 := makeDHTNode(nil, "localhost", "1118")
	node9 := makeDHTNode(nil, "localhost", "1119")

	key1 := "2b230fe12d1c9c60a8e489d028417ac89de57635"
	key2 := "87adb987ebbd55db2c5309fd4b23203450ab0083"
	key3 := "74475501523a71c34f945ae4e87d571c2c57f6f3"

	fmt.Println("TEST: " + node1.lookup(key1).nodeId + " is responsible for " + key1)
	fmt.Println("TEST: " + node1.lookup(key2).nodeId + " is responsible for " + key2)
	fmt.Println("TEST: " + node1.lookup(key3).nodeId + " is responsible for " + key3)

	node1.addToRing(node2)
	node1.addToRing(node3)
	node1.addToRing(node4)
	node4.addToRing(node5)
	node3.addToRing(node6)
	node3.addToRing(node7)
	node3.addToRing(node8)
	node7.addToRing(node9)

	fmt.Println("-> ring structure")
	node1.printRing()

	nodeForKey1 := node1.lookup(key1)
	fmt.Println("dht node " + nodeForKey1.nodeId + " running at " + nodeForKey1.contact.ip + ":" + nodeForKey1.contact.port + " is responsible for " + key1)

	nodeForKey2 := node1.lookup(key2)
	fmt.Println("dht node " + nodeForKey2.nodeId + " running at " + nodeForKey2.contact.ip + ":" + nodeForKey2.contact.port + " is responsible for " + key2)

	nodeForKey3 := node1.lookup(key3)
	fmt.Println("dht node " + nodeForKey3.nodeId + " running at " + nodeForKey3.contact.ip + ":" + nodeForKey3.contact.port + " is responsible for " + key3)
}
