package merklehasher_test

import (
	"bytes"
	"crypto"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/merklehasher"

	// import implementation.
	_ "golang.org/x/crypto/blake2b"
)

func TestMerkleHasher(t *testing.T) {
	var includedBlocks iotago.BlockIDs

	// https://github.com/Wollac/iota-crypto-demo/tree/master/examples/merkle

	includedBlocks = append(includedBlocks, iotago.MustSlotIdentifierFromHexString("0x52fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c6490000000000000000"))
	includedBlocks = append(includedBlocks, iotago.MustSlotIdentifierFromHexString("0x81855ad8681d0d86d1e91e00167939cb6694d2c422acd208a0072939487f69990000000000000000"))
	includedBlocks = append(includedBlocks, iotago.MustSlotIdentifierFromHexString("0xeb9d18a44784045d87f3c67cf22746e995af5a25367951baa2ff6cd471c483f10000000000000000"))
	includedBlocks = append(includedBlocks, iotago.MustSlotIdentifierFromHexString("0x5fb90badb37c5821b6d95526a41a9504680b4e7c8b763a1b1d49d4955c8486210000000000000000"))
	includedBlocks = append(includedBlocks, iotago.MustSlotIdentifierFromHexString("0x6325253fec738dd7a9e28bf921119c160f0702448615bbda08313f6a8eb668d20000000000000000"))
	includedBlocks = append(includedBlocks, iotago.MustSlotIdentifierFromHexString("0x0bf5059875921e668a5bdf2c7fc4844592d2572bcd0668d2d6c52f5054e2d0830000000000000000"))
	includedBlocks = append(includedBlocks, iotago.MustSlotIdentifierFromHexString("0x6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f0000000000000000"))

	hasher := merklehasher.NewHasher[iotago.BlockID](crypto.BLAKE2b_256)
	hash, err := hasher.HashValues(includedBlocks)
	require.NoError(t, err)

	expectedHash, err := iotago.DecodeHex("0x1f943b38f71a984d154b7ffef0ff0481c644b3c796d3a451fbe7ebd909b1022c")
	require.NoError(t, err)
	require.True(t, bytes.Equal(hash, expectedHash))

	for i := 0; i < len(includedBlocks); i++ {
		path, err := hasher.ComputeProofForIndex(includedBlocks, i)
		require.NoError(t, err)

		require.True(t, bytes.Equal(hash, path.Hash(hasher)))
		isProof, err := path.ContainsValue(includedBlocks[i])
		require.NoError(t, err)
		require.True(t, isProof)

		jsonPath, err := json.Marshal(path)
		require.NoError(t, err)

		pathFromJSON := new(merklehasher.Proof[iotago.BlockID])
		err = json.Unmarshal(jsonPath, pathFromJSON)
		require.NoError(t, err)
		require.True(t, bytes.Equal(hash, pathFromJSON.Hash(hasher)))
	}
}
