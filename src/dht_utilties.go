package dht

import (
	"bytes"
	"crypto/sha1"
	// "encoding/hex"
	"fmt"
	// "github.com/nu7hatch/gouuid"
	"math/big"
)

const bits = 160

func distance(a, b []byte, bitsvar int) *big.Int {
	var ring big.Int
	ring.Exp(big.NewInt(2), big.NewInt(int64(bitsvar)), nil)

	var a_int, b_int big.Int
	(&a_int).SetBytes(a)
	(&b_int).SetBytes(b)

	var dist big.Int
	(&dist).Sub(&b_int, &a_int)

	(&dist).Mod(&dist, &ring)
	return &dist
}

func lst(id1, id2 string) bool {
	bid1 := []byte(id1)
	bid2 := []byte(id2)
	if bytes.Compare(bid1, bid2) == -1 {
		return true
	}
	return false
}

func between(id1, id2, key []byte) bool {
	// 0 if a==b, -1 if a < b, and +1 if a > b
	if bytes.Compare(key, id1) == 0 { // key == id1
		return true
	}
	if bytes.Compare(id2, id1) == 1 { // id2 > id1
		if bytes.Compare(key, id2) == -1 && bytes.Compare(key, id1) == 1 { // key < id2 && key > id1
			return true
		} else {
			return false
		}
	} else { // id1 > id2
		if bytes.Compare(key, id1) == 1 || bytes.Compare(key, id2) == -1 { // key > id1 || key < id2
			return true
		} else {
			return false
		}
	}
}

// (n + 2^(k-1)) mod (2^m)
func calcFinger(n []byte, k int, m int) (string, []byte) {
	//fmt.Println("calculating result = (n+2^(k-1)) mod (2^m)")

	// convert the n to a bigint
	nBigInt := big.Int{}
	nBigInt.SetBytes(n)

	//fmt.Printf("n            %s\n", nBigInt.String())
	//fmt.Printf("k            %d\n", k)
	//fmt.Printf("m            %d\n", m)

	// get the right addend, i.e. 2^(k-1)
	two := big.NewInt(2)
	addend := big.Int{}
	addend.Exp(two, big.NewInt(int64(k-1)), nil)

	//fmt.Printf("2^(k-1)      %s\n", addend.String())

	// calculate sum
	sum := big.Int{}
	sum.Add(&nBigInt, &addend)

	//fmt.Printf("(n+2^(k-1))  %s\n", sum.String())

	// calculate 2^m
	ceil := big.Int{}
	ceil.Exp(two, big.NewInt(int64(m)), nil)

	//fmt.Printf("2^m          %s\n", ceil.String())

	// apply the mod
	result := big.Int{}
	result.Mod(&sum, &ceil)

	//fmt.Printf("result       %s\n", result.String())

	resultBytes := result.Bytes()
	resultHex := fmt.Sprintf("%x", resultBytes)

	//fmt.Printf("result (hex) %s\n", resultHex)

	return resultHex, resultBytes
}

func generateNodeId(name string) string {
	// old
	/*u, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}*/

	// calculate sha-1 hash
	hasher := sha1.New()
	hasher.Write([]byte(name))

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

/* Two values to have recursive solution which is more easily understandable.
* Method stabalizes and updates all finger-tables (which is essentially only a list of nodes.)
 */
/*
func (dhtNode *DHTNode) stabilizeRing(stopNode string) {
	for k := 0; k < bits; k++ {
		idBytes, _ := hex.DecodeString(dhtNode.nodeId)
		fingerID, _ := calcFinger(idBytes, k+1, bits)
		dhtNode.fingers[k] = dhtNode.lookup(fingerID)
	}
	if dhtNode.successor.nodeId != stopNode {
		dhtNode.successor.stabilizeRing(stopNode)
	}
}
*/
