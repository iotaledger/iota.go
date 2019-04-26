package merkle_test

import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/merkle"
	. "github.com/iotaledger/iota.go/trinary"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	leafCount = 32
)

var _ = Describe("Merkle", func() {

	Context("Merkle()", func() {
		It("Merkle", func() {
			var size, depth uint64
			var err error
			var merkleTree, siblings, hash Trits

			size = MerkleSize(leafCount)
			Expect(size).To(Equal(uint64(63)))

			depth = MerkleDepth(size)
			Expect(depth).To(Equal(uint64(6)))

			merkleTree = make(Trits, size*HashTrinarySize)
			siblings = make(Trits, (depth-1)*HashTrinarySize)
			hash = make(Trits, HashTrinarySize)

			merkleTree, err = MerkleCreate(leafCount, "ABCDEFGHIJKLMNOPQRSTUVWXYZ9ABCDEFGHIJKLMNOPQRSTUVWXYZ9ABCDEFGHIJKLMNOPQRSTUVWXYZ9", 7, SecurityLevel(2), NewCurlP81())
			Expect(err).ToNot(HaveOccurred())

			for i := uint64(0); i < leafCount; i++ {
				siblings, err = MerkleBranch(merkleTree, siblings, size*HashTrinarySize, depth, i, leafCount)
				Expect(err).ToNot(HaveOccurred())

				merkleTreeIdx := MerkleNodeIndex(depth-1, MerkleLeafIndex(i, leafCount), depth-1) * HashTrinarySize

				copy(hash, merkleTree[merkleTreeIdx:merkleTreeIdx+HashTrinarySize])

				hash, err = MerkleRoot(hash, siblings, depth-1, i, NewCurlP81())
				Expect(err).ToNot(HaveOccurred())

				Expect(merkleTree[:HashTrinarySize]).To(Equal(hash))
			}
		})
	})
})
