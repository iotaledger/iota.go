package kerl_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/stretchr/testify/assert"
)

func TestKerlBytesZeroLastTrit(t *testing.T) {
	t.Run("0 → 0", func(t *testing.T) {
		bs := bytes.Repeat([]byte{0}, consts.HashBytesSize)
		kerl.KerlBytesZeroLastTrit(bs)
		assert.Equal(t, bytes.Repeat([]byte{0}, consts.HashBytesSize), bs)
	})

	t.Run("⌊3²⁴² / 2⌋ → ⌊3²⁴² / 2⌋", func(t *testing.T) {
		// in: ⌊3²⁴² / 2⌋
		in := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"
		// expected:  ⌊3²⁴² / 2⌋
		expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

		bs, err := hex.DecodeString(in)
		assert.NoError(t, err)
		kerl.KerlBytesZeroLastTrit(bs)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("-⌊3²⁴² / 2⌋ → -⌊3²⁴² / 2⌋", func(t *testing.T) {
		// in: -⌊3²⁴² / 2⌋
		in := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"
		// expected:  -⌊3²⁴² / 2⌋
		expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

		bs, err := hex.DecodeString(in)
		assert.NoError(t, err)
		kerl.KerlBytesZeroLastTrit(bs)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("-⌊3²⁴² / 2⌋ - 1 → ⌊3²⁴² / 2⌋", func(t *testing.T) {
		// in: -⌊3²⁴² / 2⌋ - 1
		in := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769b"
		// expected:  -⌊3²⁴² / 2⌋ - 1 + 3²⁴²
		expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

		bs, err := hex.DecodeString(in)
		assert.NoError(t, err)
		kerl.KerlBytesZeroLastTrit(bs)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("⌊3²⁴² / 2⌋ + 1 → -⌊3²⁴² / 2⌋", func(t *testing.T) {
		// in: ⌊3²⁴² / 2⌋ + 1
		in := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8965"
		// expected:  ⌊3²⁴² / 2⌋ + 1 - 3²⁴²
		expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

		bs, err := hex.DecodeString(in)
		assert.NoError(t, err)
		kerl.KerlBytesZeroLastTrit(bs)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("⌊3²⁴² / 2⌋ + 1 → -⌊3²⁴² / 2⌋", func(t *testing.T) {
		// in: ⌊3²⁴² / 2⌋ + 1
		in := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8965"
		// expected:  ⌊3²⁴² / 2⌋ + 1 - 3²⁴²
		expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

		bs, err := hex.DecodeString(in)
		assert.NoError(t, err)
		kerl.KerlBytesZeroLastTrit(bs)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("2³⁸³ - 1 → 2³⁸³ - 1 - 3²⁴²", func(t *testing.T) {
		// in: 2³⁸³ - 1
		in := "7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
		// expected:  2³⁸³ - 1 - 3²⁴²
		expected := "c32c2820af00a840abf2aff4ac12fee96e88a3930cb671fb5ed0c551e7b76f238a43fe4dd6f75f60c1ff132cb462ed36"

		bs, err := hex.DecodeString(in)
		assert.NoError(t, err)
		kerl.KerlBytesZeroLastTrit(bs)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("-2³⁸³ → -2³⁸³ + 3²⁴²", func(t *testing.T) {
		// in: -2³⁸³
		in := "800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
		// expected:  -2³⁸³ + 3²⁴²
		expected := "3cd3d7df50ff57bf540d500b53ed011691775c6cf3498e04a12f3aae184890dc75bc01b22908a09f3e00ecd34b9d12c9"

		bs, err := hex.DecodeString(in)
		assert.NoError(t, err)
		kerl.KerlBytesZeroLastTrit(bs)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})
}

func TestKerlTritsToBytes(t *testing.T) {
	t.Run("should return bytes for 0", func(t *testing.T) {
		// in: balanced 243-trit representation of 0
		trits := make(trinary.Trits, consts.HashTrinarySize)
		// expected: unsigned 384-bit representation of 0
		expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

		bs, err := kerl.KerlTritsToBytes(trits)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for 1", func(t *testing.T) {
		// in: balanced 243-trit representation of 1
		trits, _ := trinary.PadTrits(trinary.Trits{1}, consts.HashTrinarySize)
		// expected: unsigned 384-bit representation of 1
		expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"

		bs, err := kerl.KerlTritsToBytes(trits)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for -1", func(t *testing.T) {
		// in: balanced 243-trit representation of -1
		trits, _ := trinary.PadTrits(trinary.Trits{-1}, consts.HashTrinarySize)
		// expected: unsigned 384-bit representation of -1
		expected := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

		bs, err := kerl.KerlTritsToBytes(trits)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for all '1's", func(t *testing.T) {
		// in: balanced 243-trit representation of ⌊3²⁴³ / 2⌋
		trits := trinary.MustTrytesToTrits("MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM")
		// expected: unsigned 384-bit representation of ⌊3²⁴² / 2⌋
		expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

		bs, err := kerl.KerlTritsToBytes(trits)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for all '-1's", func(t *testing.T) {
		// in: balanced 243-trit representation of -⌊3²⁴³ / 2⌋
		trits := trinary.MustTrytesToTrits("NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN")
		// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
		expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

		bs, err := kerl.KerlTritsToBytes(trits)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for the uint32 chunk size", func(t *testing.T) {
		// in: balanced 243-trit representation of 27⁶
		trits, _ := trinary.PadTrits(trinary.IntToTrits(0x17179149), consts.HashTrinarySize)
		// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
		expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000017179149"

		bs, err := kerl.KerlTritsToBytes(trits)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for min number with positive last trit", func(t *testing.T) {
		// in: balanced 243-trit representation of ⌊3²⁴² / 2⌋ + 1
		trits := trinary.MustTrytesToTrits("NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNE")
		// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
		expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

		bs, err := kerl.KerlTritsToBytes(trits)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for max number with negative last trit", func(t *testing.T) {
		// in: balanced 243-trit representation of -⌊3²⁴² / 2⌋ - 1
		trits := trinary.MustTrytesToTrits("MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMV")
		// expected: unsigned 384-bit representation of ⌊3²⁴² / 2⌋
		expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

		bs, err := kerl.KerlTritsToBytes(trits)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should handle internal carries", func(t *testing.T) {
		// the following trytes correspond to 2^320 leading to additions with many carries
		trits := trinary.MustTrytesToTrits("NNNNNNNNNNNNIPWAK9KOEYFFRZLJXRFLFLBRBFQATTA9TLIDNFNIEMCSPPUHKUGISALJSLL9PSXBQXEPW")
		expected := "a3ab5256e3121af155f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

		bs, err := kerl.KerlTritsToBytes(trits)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for valid trits", func(t *testing.T) {
		trits := trinary.MustTrytesToTrits("9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX")
		expected := "c88581022f0df1dd6289b737d911363a2390e2d379a2940a77ca15203024629b02fd392859dc58d3774ef615792ce00f"

		bs, err := kerl.KerlTritsToBytes(trits)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return an error for invalid trits slice length", func(t *testing.T) {
		_, err := kerl.KerlTritsToBytes(trinary.Trits{1, 1})
		assert.Error(t, err)
	})
}

func TestKerlTrytesToBytes(t *testing.T) {
	t.Run("should return bytes for 0", func(t *testing.T) {
		// in: balanced 81-tryte representation of 0
		trytes := trinary.IntToTrytes(0, consts.HashTrytesSize)
		// expected: unsigned 384-bit representation of 0
		expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

		bs, err := kerl.KerlTrytesToBytes(trytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for 1", func(t *testing.T) {
		// in: balanced 81-tryte representation of 1
		trytes := trinary.IntToTrytes(1, consts.HashTrytesSize)
		// expected: unsigned 384-bit representation of 1
		expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"

		bs, err := kerl.KerlTrytesToBytes(trytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for -1", func(t *testing.T) {
		// in: balanced 81-tryte representation of -1
		trytes := trinary.IntToTrytes(-1, consts.HashTrytesSize)
		// expected: unsigned 384-bit representation of -1
		expected := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

		bs, err := kerl.KerlTrytesToBytes(trytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for all 'M's", func(t *testing.T) {
		// in: balanced 81-tryte representation of ⌊3²⁴³ / 2⌋
		trytes := "MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM"
		// expected: unsigned 384-bit representation of ⌊3²⁴² / 2⌋
		expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

		bs, err := kerl.KerlTrytesToBytes(trytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for all 'N's", func(t *testing.T) {
		// in: balanced 81-tryte representation of -⌊3²⁴³ / 2⌋
		trytes := "NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN"
		// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
		expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

		bs, err := kerl.KerlTrytesToBytes(trytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for the uint32 chunk size", func(t *testing.T) {
		// in: balanced 81-tryte representation of 27⁶
		trytes := trinary.IntToTrytes(0x17179149, consts.HashTrytesSize)
		// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
		expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000017179149"

		bs, err := kerl.KerlTrytesToBytes(trytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for min number with positive last trit", func(t *testing.T) {
		// in: balanced 81-tryte representation of ⌊3²⁴² / 2⌋ + 1
		trytes := "NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNE"
		// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
		expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

		bs, err := kerl.KerlTrytesToBytes(trytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for max number with negative last trit", func(t *testing.T) {
		// in: balanced 81-tryte representation of -⌊3²⁴² / 2⌋ - 1
		trytes := "MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMV"
		// expected: unsigned 384-bit representation of ⌊3²⁴² / 2⌋
		expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

		bs, err := kerl.KerlTrytesToBytes(trytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should handle internal carries", func(t *testing.T) {
		// the following trytes correspond to 2^320 leading to additions with many carries
		trytes := "NNNNNNNNNNNNIPWAK9KOEYFFRZLJXRFLFLBRBFQATTA9TLIDNFNIEMCSPPUHKUGISALJSLL9PSXBQXEPW"
		expected := "a3ab5256e3121af155f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

		bs, err := kerl.KerlTrytesToBytes(trytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return bytes for valid trytes", func(t *testing.T) {
		trytes := "9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX"
		expected := "c88581022f0df1dd6289b737d911363a2390e2d379a2940a77ca15203024629b02fd392859dc58d3774ef615792ce00f"

		bs, err := kerl.KerlTrytesToBytes(trytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, hex.EncodeToString(bs))
	})

	t.Run("should return an error for invalid trytes slice length", func(t *testing.T) {
		_, err := kerl.KerlTrytesToBytes("99")
		assert.Error(t, err)
	})
}

func TestKerlBytesToTrits(t *testing.T) {
	t.Run("should return trits for all '0x00's", func(t *testing.T) {
		bytes, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		expected, _ := trinary.PadTrits(trinary.IntToTrits(0), consts.HashTrinarySize)

		trits, err := kerl.KerlBytesToTrits(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trits)
	})

	t.Run("should return trits for max", func(t *testing.T) {
		bytes, _ := hex.DecodeString("800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		expected := trinary.MustTrytesToTrits("VTPNBFOMDQVXGNPDQGHWKHKXLWMYDWXHLVJIQZVJGGAIPXJ9MAAAGXXEQBCBCXBWEWRNODU9NGUQWNHXC")

		trits, err := kerl.KerlBytesToTrits(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trits)
	})

	t.Run("should return trits for min", func(t *testing.T) {
		bytes, _ := hex.DecodeString("7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
		expected := trinary.MustTrytesToTrits("DGKMYULNWJECTMKWJTSDPSPCODNBWDCSOEQRJAEQTTZRKCQ9NZZZTCCVJYXYXCYDVDIMLWF9MTFJDMSCX")

		trits, err := kerl.KerlBytesToTrits(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trits)
	})

	t.Run("should return trytes for ⌊3²⁴² / 2⌋ + 1", func(t *testing.T) {
		bytes, _ := hex.DecodeString("5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8965")
		expected := trinary.MustTrytesToTrits("NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNW")

		trits, err := kerl.KerlBytesToTrits(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trits)
	})

	t.Run("should return trytes for -⌊3²⁴² / 2⌋ - 1", func(t *testing.T) {
		bytes, _ := hex.DecodeString("a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769b")
		expected := trinary.MustTrytesToTrits("MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMD")

		trits, err := kerl.KerlBytesToTrits(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trits)
	})

	t.Run("should return trits for valid bytes", func(t *testing.T) {
		bytes, _ := hex.DecodeString("c88581022f0df1dd6289b737d911363a2390e2d379a2940a77ca15203024629b02fd392859dc58d3774ef615792ce00f")
		expected := trinary.MustTrytesToTrits("9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX")

		trits, err := kerl.KerlBytesToTrits(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trits)
	})

	t.Run("should return an error for invalid bytes slice length", func(t *testing.T) {
		_, err := kerl.KerlBytesToTrits([]byte{1, 45, 62})
		assert.Error(t, err)
	})
}

func TestKerlBytesToTrytes(t *testing.T) {
	t.Run("should return trytes for all '0x00's", func(t *testing.T) {
		bytes, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		expected := trinary.IntToTrytes(0, consts.HashTrytesSize)

		trytes, err := kerl.KerlBytesToTrytes(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trytes)
	})

	t.Run("should return trytes for max", func(t *testing.T) {
		bytes, _ := hex.DecodeString("800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		expected := "VTPNBFOMDQVXGNPDQGHWKHKXLWMYDWXHLVJIQZVJGGAIPXJ9MAAAGXXEQBCBCXBWEWRNODU9NGUQWNHXC"

		trytes, err := kerl.KerlBytesToTrytes(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trytes)
	})

	t.Run("should return trytes for min", func(t *testing.T) {
		bytes, _ := hex.DecodeString("7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
		expected := "DGKMYULNWJECTMKWJTSDPSPCODNBWDCSOEQRJAEQTTZRKCQ9NZZZTCCVJYXYXCYDVDIMLWF9MTFJDMSCX"

		trytes, err := kerl.KerlBytesToTrytes(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trytes)
	})

	t.Run("should return trytes for ⌊3²⁴² / 2⌋ + 1", func(t *testing.T) {
		bytes, _ := hex.DecodeString("5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8965")
		expected := "NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNW"

		trytes, err := kerl.KerlBytesToTrytes(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trytes)
	})

	t.Run("should return trytes for -⌊3²⁴² / 2⌋ - 1", func(t *testing.T) {
		bytes, _ := hex.DecodeString("a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769b")
		expected := "MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMD"

		trytes, err := kerl.KerlBytesToTrytes(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trytes)
	})

	t.Run("should return trytes for valid bytes", func(t *testing.T) {
		bytes, _ := hex.DecodeString("c88581022f0df1dd6289b737d911363a2390e2d379a2940a77ca15203024629b02fd392859dc58d3774ef615792ce00f")
		expected := "9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX"

		trytes, err := kerl.KerlBytesToTrytes(bytes)
		assert.NoError(t, err)
		assert.Equal(t, expected, trytes)
	})

	t.Run("should return an error for invalid bytes slice length", func(t *testing.T) {
		_, err := kerl.KerlBytesToTrytes([]byte{1, 45, 62})
		assert.Error(t, err)
	})
}
