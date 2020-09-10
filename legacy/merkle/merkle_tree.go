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

	"github.com/iotaledger/iota.go/guards"
	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/kerl"
	"github.com/iotaledger/iota.go/legacy/signing"
	"github.com/iotaledger/iota.go/legacy/signing/key"
	sponge "github.com/iotaledger/iota.go/legacy/signing/utils"
	"github.com/iotaledger/iota.go/legacy/trinary"
)

var (
	// ErrDepthTooSmall is returned when the depth for creating the Merkle tree is too low.
	ErrDepthTooSmall = errors.New("depth must be positive")
	// ErrDepthTooLarge is returned when the depth for creating the Merkle tree is too high.
	ErrDepthTooLarge = errors.New("largest depth is 32")
	// ErrInvalidAuditPathLength is returned when the length of the Merkle audit path does not match the leaf index.
	ErrInvalidAuditPathLength = errors.New("invalid length of the audit path")
	// ErrInvalidLeafIndex is returned for invalid leaf indices.
	ErrInvalidLeafIndex = errors.New("invalid leaf index")
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

// MerkleCreateOptions is used to pass optional creation options to CreateMerkleTree.
type MerkleCreateOptions struct {
	// CalculateAddressesStartCallback will be called at the start of leaf generation, with the total count of the leaves.
	CalculateAddressesStartCallback func(count uint32)
	// CalculateAddressesCallback will be called after each leaf generation, with the corresponding index.
	CalculateAddressesCallback func(index uint32)
	// CalculateAddressesFinishedCallback will be called after leaf generation is finished, with the total count of the leaves.
	CalculateAddressesFinishedCallback func(count uint32)
	// CalculateLayersCallback will be called before each layer generation, with the corresponding index.
	CalculateLayersCallback func(index uint32)
	// The number of parallel threads used by the generation routines.
	Parallelism int
}

// calculateAllAddresses calculates all addresses that are used for the Merkle tree of the coordinator.
func calculateAllAddresses(seed trinary.Hash, securityLvl legacy.SecurityLevel, count uint32, opts ...MerkleCreateOptions) []trinary.Hash {

	var progressStartCallback func(uint32) = nil
	var progressCallback func(uint32) = nil
	var progressFinishedCallback func(uint32) = nil
	parallelism := runtime.GOMAXPROCS(0)

	if len(opts) > 0 {
		if opts[0].CalculateAddressesStartCallback != nil {
			progressStartCallback = opts[0].CalculateAddressesStartCallback
		}
		if opts[0].CalculateAddressesCallback != nil {
			progressCallback = opts[0].CalculateAddressesCallback
		}
		if opts[0].CalculateAddressesFinishedCallback != nil {
			progressFinishedCallback = opts[0].CalculateAddressesFinishedCallback
		}
		if opts[0].Parallelism > 0 {
			parallelism = opts[0].Parallelism
		}
	}

	if progressStartCallback != nil {
		progressStartCallback(count)
	}

	result := make([]trinary.Hash, count)

	wg := sync.WaitGroup{}
	wg.Add(parallelism)

	// calculate all addresses in parallel.
	input := make(chan uint32)
	for i := 0; i < parallelism; i++ {
		go func() {
			defer wg.Done()

			for index := range input {
				address, err := computeAddress(seed, index, securityLvl)
				if err != nil {
					panic(err)
				}
				result[index] = address
			}
		}()
	}

	for index := uint32(0); index < count; index++ {
		input <- index

		if progressCallback != nil {
			progressCallback(index)
		}
	}

	close(input)
	wg.Wait()

	if progressFinishedCallback != nil {
		progressFinishedCallback(count)
	}

	return result
}

// calculateAllLayers calculates all layers of the Merkle tree used for coordinator signatures.
func calculateAllLayers(addresses []trinary.Hash, opts ...MerkleCreateOptions) [][]trinary.Hash {
	depth := bits.Len(uint(len(addresses))) - 1

	var progressCallback func(uint32) = nil

	if len(opts) > 0 {
		if opts[0].CalculateLayersCallback != nil {
			progressCallback = opts[0].CalculateLayersCallback
		}
	}

	// depth+1 because it has to include the Root at [0]
	layers := make([][]trinary.Hash, depth+1)

	layers[depth] = addresses

	for i := depth - 1; i >= 0; i-- {
		if progressCallback != nil {
			progressCallback(uint32(i))
		}
		layers[i] = calculateNextLayer(layers[i+1], opts...)
	}

	return layers
}

// calculateNextLayer calculates a single layer of the Merkle tree used for coordinator signatures.
func calculateNextLayer(lastLayer []trinary.Hash, opts ...MerkleCreateOptions) []trinary.Hash {

	parallelism := runtime.GOMAXPROCS(0)

	if len(opts) > 0 {
		if opts[0].Parallelism > 0 {
			parallelism = opts[0].Parallelism
		}
	}

	result := make([]trinary.Hash, len(lastLayer)/2)

	wg := sync.WaitGroup{}
	wg.Add(parallelism)

	// calculate all nodes in parallel.
	input := make(chan int)
	for i := 0; i < parallelism; i++ {
		go func() {
			defer wg.Done()

			for index := range input {
				sp := kerl.NewKerl()

				// Merkle trees are calculated layer by layer by hashing two corresponding nodes of the last layer.
				// https://en.wikipedia.org/wiki/Merkle_tree
				sp.MustAbsorbTrytes(lastLayer[index*2])
				sp.MustAbsorbTrytes(lastLayer[index*2+1])

				result[index] = sp.MustSqueezeTrytes(legacy.HashTrinarySize)
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

func computeKey(seed trinary.Hash, index uint32, securityLvl legacy.SecurityLevel, spongeFunc sponge.SpongeFunction) (trinary.Trits, error) {
	subSeedTrits, err := signing.Subseed(seed, uint64(index), spongeFunc)
	if err != nil {
		return nil, err
	}

	return key.Shake(subSeedTrits, securityLvl)
}

// computeAddress generates an address deterministically, according to the given seed, subseed index and security level;
// a modified key derivation function is used to avoid the M-bug.
func computeAddress(seed trinary.Hash, index uint32, securityLvl legacy.SecurityLevel) (trinary.Hash, error) {
	k := kerl.NewKerl()

	keyTrits, err := computeKey(seed, index, securityLvl, k)
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
// deriving subseeds from the provided seed. An optional MerkleCreateOptions struct can be
// passed to specify function's parallelism and progress callback.
func CreateMerkleTree(seed trinary.Hash, securityLvl legacy.SecurityLevel, depth int, opts ...MerkleCreateOptions) (*MerkleTree, error) {
	if depth < 1 {
		return nil, ErrDepthTooSmall
	}
	if depth > 32 {
		return nil, ErrDepthTooLarge
	}

	if !guards.IsHash(seed) {
		return nil, legacy.ErrInvalidSeed
	}

	addresses := calculateAllAddresses(seed, securityLvl, 1<<uint(depth), opts...)
	layers := calculateAllLayers(addresses, opts...)

	mt := &MerkleTree{Depth: depth}
	// depth+1 because it has to include the Root at [0]
	mt.Layers = make([]*MerkleTreeLayer, depth+1)

	for i, layer := range layers {
		mt.Layers[i] = &MerkleTreeLayer{Level: i, Hashes: layer}
	}

	mt.Root = mt.Layers[0].Hashes[0]

	return mt, nil
}

// AuditPath returns the Merkle audit path for the provided leaf.
// The audit path is the slice of missing hashes required to compute the nodes from the leaf to the root of the tree.
func (mt *MerkleTree) AuditPath(leafIndex uint32) ([]trinary.Hash, error) {
	if uint64(leafIndex|1) > uint64(len(mt.Layers[mt.Depth].Hashes)) {
		return nil, ErrInvalidLeafIndex
	}

	path := make([]trinary.Hash, 0, mt.Depth)
	for currentLayerIndex := mt.Depth; currentLayerIndex > 0; currentLayerIndex-- {
		layer := mt.Layers[currentLayerIndex]

		if leafIndex%2 == 0 {
			// even
			path = append(path, layer.Hashes[leafIndex+1])
		} else {
			// odd
			path = append(path, layer.Hashes[leafIndex-1])
		}

		leafIndex /= 2
	}
	return path, nil
}
