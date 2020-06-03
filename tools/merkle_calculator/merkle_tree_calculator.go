package main

import (
	"flag"
	"log"
	"time"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/guards"
	"github.com/iotaledger/iota.go/merkle"
	"github.com/iotaledger/iota.go/trinary"
)

func main() {

	var (
		depth         = flag.Int("depth", 0, "Depth of the Merkle tree to create")
		securityLevel = flag.Int("securityLevel", 0, "Security level of the private key used")
		seed          = flag.String("seed", "", "Seed for leaves derivation")
		outputPath    = flag.String("output", "", "Output file path")
	)

	flag.Parse()

	if *depth < 1 {
		log.Panicf("'depth' cannot be lower than 1")
		return
	}

	if *securityLevel < 1 || *securityLevel > 3 {
		log.Panicf("'securityLevel' must be 1, 2 or 3")
		return
	}

	if !guards.IsTrytesOfExactLength(*seed, consts.HashTrytesSize) {
		log.Panicf("'seed' must be a string of 81 trytes")
		return
	}

	if *outputPath == "" {
		log.Panicf("'output' is required")
		return
	}

	log.Printf("calculating %d addresses...\n", 1<<uint(*depth))

	ts := time.Now()

	mt := merkle.CreateMerkleTree(trinary.Hash(*seed), *securityLevel, *depth)

	if err := merkle.StoreMerkleTreeFile(*outputPath, mt); err != nil {
		log.Panicf("Error persisting Merkle tree: %v", err)
		return
	}

	log.Printf("Merkle tree root: %v\n", mt.Root)

	log.Printf("Took %v seconds.\n", time.Since(ts).Truncate(time.Second))

}
