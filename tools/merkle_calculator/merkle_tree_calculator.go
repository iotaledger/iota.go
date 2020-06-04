package main

import (
	"flag"
	"log"
	"time"

	"github.com/iotaledger/iota.go/guards"
	"github.com/iotaledger/iota.go/merkle"
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

	if !guards.IsTransactionHash(*seed) {
		log.Panicf("'seed' must be a string of 81 trytes")
		return
	}

	if *outputPath == "" {
		log.Panicf("'output' is required")
		return
	}

	count := 1 << uint(*depth)

	log.Printf("calculating %d addresses...\n", count)

	ts := time.Now()

	mt, err := merkle.CreateMerkleTree(*seed, *securityLevel, *depth,
		func(index int) {
			if index%5000 == 0 && index != 0 {
				ratio := float64(index) / float64(count)
				total := time.Duration(float64(time.Since(ts)) / ratio)
				duration := time.Until(ts.Add(total))
				log.Printf("calculated %d/%d (%0.2f%%) addresses. %v left...\n", index, count, ratio*100.0, duration.Truncate(time.Second))
			}
		})

	if err != nil {
		log.Panicf("Error creating Merkle tree: %v", err)
		return
	}

	if err := merkle.StoreMerkleTreeFile(*outputPath, mt); err != nil {
		log.Panicf("Error persisting Merkle tree: %v", err)
		return
	}

	log.Printf("Merkle tree root: %v\n", mt.Root)

	log.Printf("Took %v seconds.\n", time.Since(ts).Truncate(time.Second))

}
