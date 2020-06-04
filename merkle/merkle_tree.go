// Ported from https://github.com/gohornet/hornet codebase.
// Original authors: muXxer <mux3r@web.de>
//                   Alexander Sporn <github@alexsporn.de>
//                   Thoralf-M <46689931+Thoralf-M@users.noreply.github.com>

// Package merkle provides functions and types to deal with the creation and storage of
// Merkle trees, using the secure SHAKE256 KDF implemented in the signing/key package:
// thus not being affected by the the infamous M-Bug.
package merkle

import (
	"errors"
	"math/bits"
	"runtime"
	"sync"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl"
	"github.com/iotaledger/iota.go/signing"
	"github.com/iotaledger/iota.go/signing/key"
	"github.com/iotaledger/iota.go/trinary"
)

var (
	errInvDepth = errors.New("invalid depth")
)

// MerkleTree contains the Merkle tree used for the coordinator signatures.
type MerkleTree struct {
	// The depth of the Merkle tree.
	Depth int
	// The root address of the Merkle tree.
	Root trinary.Hash
	// Merkle tree layers indexed by their level.
	Layers []*MerkleTreeLayer
}

// MerkleTreeLayer contains the nodes of a layer of a Merkle tree.
type MerkleTreeLayer struct {
	// The level of the layer in the tree.
	Level int
	// The nodes of the layer.
	Hashes []trinary.Hash
}

// milestoneIndex represents an index of a milestone.
type milestoneIndex uint32

// calculateAllAddresses calculates all addresses that are used for the Merkle tree of the coordinator.
func calculateAllAddresses(seed trinary.Hash, securityLvl int, count int, progressCallback ...func(int)) []trinary.Hash {
	result := make([]trinary.Hash, count)

	wg := sync.WaitGroup{}
	wg.Add(runtime.NumCPU())

	// calculate all addresses in parallel.
	input := make(chan milestoneIndex)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()

			for index := range input {
				address, err := computeAddress(seed, index, securityLvl)
				if err != nil {
					panic(err)
				}
				result[int(index)] = address
			}
		}()
	}

	for index := 0; index < count; index++ {
		input <- milestoneIndex(index)

		if len(progressCallback) > 0 {
			progressCallback[0](index)
		}
	}

	close(input)
	wg.Wait()

	return result
}

// calculateAllLayers calculates all layers of the Merkle tree used for coordinator signatures.
func calculateAllLayers(addresses []trinary.Hash) [][]trinary.Hash {
	depth := bits.Len(uint(len(addresses))) - 1

	layers := make([][]trinary.Hash, depth+1)

	layers[depth] = addresses

	for i := depth - 1; i >= 0; i-- {
		layers[i] = calculateNextLayer(layers[i+1])
	}

	return layers
}

// calculateNextLayer calculates a single layer of the Merkle tree used for coordinator signatures.
func calculateNextLayer(lastLayer []trinary.Hash) []trinary.Hash {

	result := make([]trinary.Hash, len(lastLayer)/2)

	wg := sync.WaitGroup{}
	wg.Add(runtime.NumCPU())

	// calculate all nodes in parallel.
	input := make(chan int)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()

			for index := range input {
				sp := kerl.NewKerl()

				// Merkle trees are calculated layer by layer by hashing two corresponding nodes of the last layer.
				// https://en.wikipedia.org/wiki/Merkle_tree
				sp.AbsorbTrytes(lastLayer[index*2])
				sp.AbsorbTrytes(lastLayer[index*2+1])

				result[index] = sp.MustSqueezeTrytes(consts.HashTrinarySize)
			}
		}()
	}

	for index := 0; index < len(lastLayer)/2; index++ {
		input <- index
	}

	close(input)
	wg.Wait()

	return result
}

// computeAddress generates an address deterministically, according to the given seed, milestone index and security level;
// a modified key derivation function is used to avoid the M-bug.
func computeAddress(seed trinary.Hash, index milestoneIndex, securityLvl int) (trinary.Hash, error) {

	k := kerl.NewKerl()

	subSeedTrits, err := signing.Subseed(seed, uint64(index), k)
	if err != nil {
		return "", err
	}

	keyTrits, err := key.Shake(subSeedTrits, consts.SecurityLevel(securityLvl))
	if err != nil {
		return "", err
	}

	digestsTrits, err := signing.Digests(keyTrits, k)
	if err != nil {
		return "", err
	}

	addressTrits, err := signing.Address(digestsTrits, k)
	if err != nil {
		return "", err
	}

	address, err := trinary.TritsToTrytes(addressTrits)
	if err != nil {
		return "", err
	}

	return address, nil
}

// CreateMerkleTree creates a MerkleTree structure of the specified depth,
// using a SHAKE256 key of the the length specified by the supplied securitylevel,
// deriving subseeds from the provided seed.
func CreateMerkleTree(seed trinary.Hash, securityLvl int, depth int, progressCallback ...func(int)) (*MerkleTree, error) {

	if depth < 1 {
		return nil, errInvDepth
	}

	if err := trinary.ValidTrytes(seed); err != nil {
		return nil, err
	} else if len(seed) != consts.HashTrinarySize/consts.TrinaryRadix {
		return nil, consts.ErrInvalidSeed
	}

	addresses := calculateAllAddresses(seed, securityLvl, 1<<uint(depth), progressCallback...)
	layers := calculateAllLayers(addresses)

	mt := &MerkleTree{Depth: depth}
	mt.Layers = make([]*MerkleTreeLayer, depth+1)

	for i, layer := range layers {
		mt.Layers[i] = &MerkleTreeLayer{Level: i, Hashes: layer}
	}

	mt.Root = mt.Layers[0].Hashes[0]

	return mt, nil
}
