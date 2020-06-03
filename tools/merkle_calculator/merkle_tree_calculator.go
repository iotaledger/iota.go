package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/iotaledger/iota.go/merkle"
	"github.com/iotaledger/iota.go/trinary"
)

func main() {

	var depth int
	var securityLevel int
	var seed string
	var outputPath string

	flag.IntVar(&depth, "depth", 0, "Depth of the Merkle tree to create")
	flag.IntVar(&securityLevel, "securityLevel", 0, "Security level of the private key used")
	flag.StringVar(&seed, "seed", "", "Seed for leaves derivation")
	flag.StringVar(&outputPath, "output", "", "Output file path")

	flag.Parse()

	if depth < 1 {
		fmt.Println("'depth' cannot be lower than 1")
		return
	}

	if securityLevel < 1 || securityLevel > 3 {
		fmt.Println("'securityLevel' must be 1, 2 or 3")
		return
	}

	if len(seed) != 81 {
		fmt.Println("'seed' must be a string of 81 trytes")
		return
	}

	if outputPath == "" {
		fmt.Println("'output' is required")
		return
	}

	fmt.Printf("calculating %d addresses...\n", 1<<uint(depth))

	ts := time.Now()

	mt := merkle.CreateMerkleTree(trinary.Hash(seed), securityLevel, depth)

	if err := merkle.StoreMerkleTreeFile(outputPath, mt); err != nil {
		fmt.Println("Error persisting Merkle tree: %v", err)
		return
	}

	fmt.Printf("Merkle tree root: %v\n", mt.Root)

	fmt.Printf("Took %v seconds.\n", time.Since(ts).Truncate(time.Second))

}
