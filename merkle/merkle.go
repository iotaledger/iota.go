// Package merkle provides functions for creating and validating merkle trees.
package merkle

import (
	"errors"

	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/signing/legacy"
	"github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
)

var (
	merkleNullHash = make(Trits, HashTrinarySize)
)

func binaryTreeSize(acc uint64, depth uint64) uint64 {
	return (1 << (depth + 1)) - 1 + acc
}

// MerkleSize computes the size of a merkle tree, e.g. its node number.
// 	leafCount is the number of leaves of the tree
func MerkleSize(leafCount uint64) uint64 {
	var acc uint64 = 1

	if leafCount < 2 {
		return leafCount
	}

	for leafCount >= 2 {
		acc += leafCount
		leafCount = (leafCount + 1) >> 1
	}
	return acc
}

// MerkleDepth computes the depth of a merkle tree.
// 	nodeCount is the number of nodes of the tree
func MerkleDepth(nodeCount uint64) (depth uint64) {
	depth = 0

	for binaryTreeSize(0, depth) < nodeCount {
		depth++
	}

	return depth + 1
}

// merkleNodeIndexTraverse returns the node index of assign location in tree.
// The order of nodes indexes follow depth-first rule.
// 	acc is the number of nodes in the previous counting binary tree
// 	depth is the depth of the node, counting from root
// 	width is the width of the node, counting from left
// 	treeDepth is the depth of whole tree
func merkleNodeIndexTraverse(acc, depth, width, treeDepth uint64) uint64 {
	if treeDepth == 0 {
		return 0
	}

	var depthCursor uint64 = 1
	index := depth + acc
	widthCursor := width
	var widthOfLeaves uint64 = 1 << depth

	for depthCursor <= depth {
		if widthCursor >= (widthOfLeaves >> depthCursor) {
			// add whole binary tree size of the left side binary tree
			index += ((1 << (treeDepth - depthCursor + 1)) - 1)

			// counting the node index of the subtree in which the cursor currently stays in
			widthCursor = widthCursor - (widthOfLeaves >> depthCursor)
		}
		depthCursor++
	}
	return index
}

// MerkleNodeIndex indexes a given node in the tree.
// 	depth is the depth of the node, counting from root
// 	width is the width of the node, counting from left
// 	treeDepth is the depth of whole tree
func MerkleNodeIndex(depth, width, treeDepth uint64) uint64 {
	return merkleNodeIndexTraverse(0, depth, width, treeDepth)
}

// MerkleLeafIndex computes the actual site index of a leaf.
// 	leafIndex is the leaf index
// 	leafCount is the number of leaves
func MerkleLeafIndex(leafIndex, leafCount uint64) uint64 {
	return leafCount - leafIndex - 1
}

// MerkleCreate creates a merkle tree.
//	baseSize is the base size of the tree, e.g. the number of leaves
//	seed is the seed used to generate addresses - Not sent over the network
//	offset is the offset used to generate addresses
//	security is the security used to generate addresses
//	spongeFunc is the optional sponge function to use
func MerkleCreate(baseSize uint64, seed Trytes, offset uint64, security SecurityLevel, spongeFunc ...sponge.SpongeFunction) (Trits, error) {

	// enforcing the tree to be perfect by checking if the base size (number of leaves) is a power of two
	if (baseSize != 0) && (baseSize&(baseSize-1)) != 0 {
		return nil, errors.New("Base size of the merkle tree should be a power of 2")
	}

	treeMerkleSize := MerkleSize(baseSize)
	tree := make(Trits, treeMerkleSize*HashTrinarySize)

	h := sponge.GetSpongeFunc(spongeFunc, curl.NewCurlP27)

	td := MerkleDepth(treeMerkleSize) - 1

	// create base addresses
	for leafIndex := uint64(0); leafIndex < baseSize; leafIndex++ {
		subSeed, err := signing.Subseed(seed, offset+MerkleLeafIndex(leafIndex, baseSize), h)
		if err != nil {
			return nil, err
		}

		key, err := signing.Key(subSeed, security, h)
		if err != nil {
			return nil, err
		}

		keyDigests, err := signing.Digests(key, h)
		if err != nil {
			return nil, err
		}

		address, err := signing.Address(keyDigests, h)
		if err != nil {
			return nil, err
		}

		treeIdx := MerkleNodeIndex(td, leafIndex, td)
		copy(tree[treeIdx*HashTrinarySize:(treeIdx+1)*HashTrinarySize], address)
	}

	// hash tree
	curSize := baseSize
	for depth := td; depth > 0; depth-- {
		for width := uint64(0); width < curSize; width += 2 {

			merkleNodeIdxLeft := MerkleNodeIndex(depth, width, td) * HashTrinarySize
			merkleNodeIdxRight := MerkleNodeIndex(depth, width+1, td) * HashTrinarySize

			if width < curSize-1 {
				// if right hash exists, absorb right hash then left hash
				h.Absorb(tree[merkleNodeIdxRight : merkleNodeIdxRight+HashTrinarySize])
				h.Absorb(tree[merkleNodeIdxLeft : merkleNodeIdxLeft+HashTrinarySize])
			} else {
				// else, absorb the remaining hash then a null hash
				h.Absorb(tree[merkleNodeIdxLeft : merkleNodeIdxLeft+HashTrinarySize])
				h.Absorb(merkleNullHash)
			}

			// squeeze the result in the parent node
			trits, err := h.Squeeze(HashTrinarySize)
			if err != nil {
				return nil, err
			}

			parentIdx := MerkleNodeIndex(depth-1, width/2, td) * HashTrinarySize
			copy(tree[parentIdx:parentIdx+HashTrinarySize], trits)
			h.Reset()
		}
		curSize = (curSize + 1) >> 1
	}

	return tree, nil
}

// MerkleBranch creates the merkle branch to generate back root from index.
//	tree is the merkle tree - Must be allocated
//	siblings is the siblings of the indexed node - Must be allocated
//	treeLength is the length of the tree
//	treeDepth is the depth of the tree
//	leafIndex is the index of the leaf to start the branch from
//	leafCount is the number of leaves of the tree
func MerkleBranch(tree Trits, siblings Trits, treeLength, treeDepth, leafIndex, leafCount uint64) (Trits, error) {

	if tree == nil {
		return nil, errors.New("Null tree")
	}

	if siblings == nil {
		return nil, errors.New("Null sibling")
	}

	if HashTrinarySize*MerkleNodeIndex(treeDepth-1, leafIndex, treeDepth-1) >= treeLength {
		return nil, errors.New("Leaf index out of bounds")
	}

	if treeDepth > MerkleDepth(treeLength/HashTrinarySize) {
		return nil, errors.New("Depth out of bounds")
	}

	var siblingIndex, siteIndex uint64

	siblingIndex = MerkleLeafIndex(leafIndex, leafCount)

	var i int64 = 0
	for depthIndex := treeDepth - 1; depthIndex > 0; depthIndex-- {

		if (siblingIndex & 1) != 0 {
			siblingIndex--
		} else {
			siblingIndex++
		}

		siteIndex = HashTrinarySize * MerkleNodeIndex(depthIndex, siblingIndex, treeDepth-1)
		if siteIndex >= treeLength {
			// if depth width is not even, copy a null hash
			copy(siblings[i*HashTrinarySize:(i+1)*HashTrinarySize], merkleNullHash)
		} else {
			// else copy a sibling
			copy(siblings[i*HashTrinarySize:(i+1)*HashTrinarySize], tree[siteIndex:siteIndex+HashTrinarySize])
		}

		siblingIndex >>= 1
		i++
	}
	return siblings, nil
}

// MerkleRoot generates a merkle root from a hash and his siblings,
//	hash is the hash
//	siblings is the hash siblings
//	siblingsNumber is the number of siblings
//	leafIndex is the node index of the hash
//	spongeFunc is the optional sponge function to use
func MerkleRoot(hash Trits, siblings Trits, siblingsNumber uint64, leafIndex uint64, spongeFunc ...sponge.SpongeFunction) (Trits, error) {
	h := sponge.GetSpongeFunc(spongeFunc, curl.NewCurlP27)

	var j uint64 = 1
	var err error

	for i := uint64(0); i < siblingsNumber; i++ {
		siblingsIdx := i * HashTrinarySize

		if (leafIndex & j) != 0 {
			// if index is a right node, absorb a sibling (left) then the hash
			err = h.Absorb(siblings[siblingsIdx : siblingsIdx+HashTrinarySize])
			err = h.Absorb(hash)
		} else {
			// if index is a left node, absorb the hash then a sibling (right)
			err = h.Absorb(hash)
			err = h.Absorb(siblings[siblingsIdx : siblingsIdx+HashTrinarySize])
		}

		hash, err = h.Squeeze(HashTrinarySize)
		h.Reset()

		j <<= 1
	}
	return hash, err
}
