package merkle_test

import (
	"testing"

	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/merkle"
	"github.com/iotaledger/iota.go/legacy/trinary"
	"github.com/stretchr/testify/assert"
)

var (
	hashToSign = legacy.NullHashTrytes
)

func TestSigning(t *testing.T) {
	var err error
	tree, err := merkle.CreateMerkleTree(seed, securityLevel, depth, merkle.MerkleCreateOptions{Parallelism: 1})
	assert.NoError(t, err)

	tests := []struct {
		name       string
		leaveIndex uint32
	}{
		{name: "leafIndex: 0", leaveIndex: uint32(0)},
		{name: "leafIndex: 1", leaveIndex: uint32(1)},
		{name: "leafIndex: 2", leaveIndex: uint32(2)},
		{name: "max leafIndex", leaveIndex: uint32(1<<depth - 1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path, err := tree.AuditPath(test.leaveIndex)
			assert.NoError(t, err)
			fragments, err := merkle.SignatureFragments(seed, test.leaveIndex, securityLevel, hashToSign)
			assert.NoError(t, err)

			valid, err := merkle.ValidateSignatureFragments(tree.Root, test.leaveIndex, path, fragments, hashToSign)
			assert.NoError(t, err)
			assert.True(t, valid)
		})
	}

	t.Run("valid audit path", func(t *testing.T) {
		path := make([]trinary.Trytes, 32)
		for i := range path {
			path[i] = legacy.NullHashTrytes
		}
		root, err := merkle.MerkleRoot(legacy.NullHashTrytes, 1<<32-1, path)
		assert.NoError(t, err)
		assert.Equal(t, "MDKGSWENCCKHKNSHEZUX9LCCDKDJJR9BXLXXKRVMUGBLOVESSLRKWOPOE9UUZZOTOIOVMTCKQLTDQITPD", root)
	})

	t.Run("audit path too short", func(t *testing.T) {
		path := make([]trinary.Trytes, depth)
		_, err := merkle.MerkleRoot(tree.Root, 1<<depth, path)
		assert.Equal(t, merkle.ErrInvalidAuditPathLength, err)
	})

	t.Run("audit path invalid tryte lengths", func(t *testing.T) {
		path := []trinary.Trytes{""}
		_, err := merkle.MerkleRoot(tree.Root, 0, path)
		assert.Error(t, err)
	})
}
