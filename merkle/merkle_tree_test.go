package merkle_test

import (
	"testing"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/merkle"
	"github.com/iotaledger/iota.go/signing"
	"github.com/iotaledger/iota.go/signing/key"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/stretchr/testify/assert"
)

const (
	seed          = "ABCDEFGHIJKLMNOPQRSTUVWXYZ9ABCDEFGHIJKLMNOPQRSTUVWXYZ9ABCDEFGHIJKLMNOPQRSTUVWXYZ9"
	securityLevel = 2
	depth         = 7
	expectedRoot  = "WFQ9CDKRMAKHKUEPPRGF9GWXCLLHRUGYNYDNWFQI9QDPLCQISLASULACYZLXGG9GGFGNRRSXEFSDHTBLW"
)

func TestCreateMerkleTree(t *testing.T) {
	// Using Depth 7
	merkleTree, err := CreateMerkleTree(seed, securityLevel, depth)
	assert.NoError(t, err)

	t.Run("creates a correctly-sized tree", func(t *testing.T) {
		assert.Equal(t, 7, merkleTree.Layers[7].Level)
		assert.Equal(t, 3, merkleTree.Layers[3].Level)
		assert.Len(t, merkleTree.Layers, 8)
	})

	t.Run("does not use Kerl KDF", func(t *testing.T) {
		assert.NotEqual(t, "VERHESGRVSUWWZJNCKMQREASXZOIW9BBYGHV9QCLVCIGJYZOEIODSIHRCBZAFNNAJSTSC9LRHKKBLJPDB", merkleTree.Root)
	})

	t.Run("leaves are computed using Shake KDF", func(t *testing.T) {
		leavesCount := 1 << uint(depth)
		leaves := merkleTree.Layers[depth].Hashes
		for index := 0; index < leavesCount; index++ {
			subSeedTrits, _ := signing.Subseed(seed, uint64(index), kerl.NewKerl())
			keyTrits, _ := key.Shake(subSeedTrits, consts.SecurityLevel(securityLevel))
			digestsTrits, _ := signing.Digests(keyTrits, kerl.NewKerl())
			addressTrits, _ := signing.Address(digestsTrits, kerl.NewKerl())
			address, _ := trinary.TritsToTrytes(addressTrits)
			assert.Equal(t, address, leaves[index])
		}
	})

	t.Run("each node is the hash of the corresponding two children using Kerl sponge", func(t *testing.T) {
		layers := merkleTree.Layers
		for d := 1; d <= depth; d++ {
			for pair := 0; pair < 1<<uint(d); pair += 2 {
				sponge := kerl.NewKerl()
				sponge.MustAbsorbTrytes(layers[d].Hashes[pair])
				sponge.MustAbsorbTrytes(layers[d].Hashes[pair+1])
				assert.Equal(t, sponge.MustSqueezeTrytes(consts.HashTrinarySize), layers[d-1].Hashes[pair/2])
			}
		}
	})

	t.Run("match root", func(t *testing.T) {
		assert.Equal(t, merkleTree.Root, merkleTree.Layers[0].Hashes[0])
		assert.Equal(t, expectedRoot, merkleTree.Root)
	})
}
