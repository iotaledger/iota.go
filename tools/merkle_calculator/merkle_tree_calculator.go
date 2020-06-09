package main

import (
	"flag"
	"log"
	"runtime"
	"time"

	"github.com/iotaledger/iota.go/guards"
	"github.com/iotaledger/iota.go/merkle"
)

func main() {

	var (
		depth         = flag.Int("depth", 0, "Depth of the merkle tree to create")
		securityLevel = flag.Int("securityLevel", 0, "Security level of the private key used")
		parallelism   = flag.Int("parallelism", 0, "Amount of concurrent threads used")
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

	if *parallelism == 0 {
		*parallelism = runtime.GOMAXPROCS(0)
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

	progressCallback := func(index uint32) {
		if index%5000 == 0 && index != 0 {
			ratio := float64(index) / float64(count)
			total := time.Duration(float64(time.Since(ts)) / ratio)
			duration := time.Until(ts.Add(total))
			log.Printf("calculated %d/%d (%0.2f%%) addresses. %v left...\n", index, count, ratio*100.0, duration.Truncate(time.Second))
		}
	}

	mt, err := merkle.CreateMerkleTree(*seed, *securityLevel, *depth, merkle.MerkleCreateOptions{ProgressCallback: progressCallback, Parallelism: *parallelism})

	if err != nil {
		log.Panicf("Error creating merkle tree: %v", err)
		return
	}

	if err := merkle.StoreMerkleTreeFile(*outputPath, mt); err != nil {
		log.Panicf("Error persisting merkle tree: %v", err)
		return
	}

	log.Printf("merkle tree root: %v\n", mt.Root)

	log.Printf("successfully created merkle tree (took %v).\n", time.Since(ts).Truncate(time.Second))
}
