package ed25519_test

import (
	"compress/gzip"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"os"
	"path"
	"testing"

	"github.com/iotaledger/iota.go/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignVerify(t *testing.T) {
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)

	message := []byte("test message")
	sig := ed25519.Sign(privateKey, message)
	assert.True(t, ed25519.Verify(publicKey, message, sig), "valid signature rejected")

	wrongMessage := []byte("wrong message")
	assert.False(t, ed25519.Verify(publicKey, wrongMessage, sig), "signature of different message accepted")
}

func TestEqual(t *testing.T) {
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	assert.Equalf(t, public, private.Public(), " private.Public() is not Equal to public")

	otherPub, otherPriv, _ := ed25519.GenerateKey(rand.Reader)
	assert.NotEqual(t, public, otherPub)
	assert.NotEqual(t, private, otherPriv)
}

func TestMalleability(t *testing.T) {
	// https://tools.ietf.org/html/rfc8032#section-5.1.7 adds an additional test
	// that s be in [0, order). This prevents someone from adding a multiple of
	// order to s and obtaining a second valid sig for the same message.
	message := []byte{0x54, 0x65, 0x73, 0x74}
	sig := []byte{
		0x7c, 0x38, 0xe0, 0x26, 0xf2, 0x9e, 0x14, 0xaa, 0xbd, 0x05, 0x9a,
		0x0f, 0x2d, 0xb8, 0xb0, 0xcd, 0x78, 0x30, 0x40, 0x60, 0x9a, 0x8b,
		0xe6, 0x84, 0xdb, 0x12, 0xf8, 0x2a, 0x27, 0x77, 0x4a, 0xb0, 0x67,
		0x65, 0x4b, 0xce, 0x38, 0x32, 0xc2, 0xd7, 0x6f, 0x8f, 0x6f, 0x5d,
		0xaf, 0xc0, 0x8d, 0x93, 0x39, 0xd4, 0xee, 0xf6, 0x76, 0x57, 0x33,
		0x36, 0xa5, 0xc5, 0x1e, 0xb6, 0xf9, 0x46, 0xb3, 0x1d,
	}
	publicKey := []byte{
		0x7d, 0x4d, 0x0e, 0x7f, 0x61, 0x53, 0xa6, 0x9b, 0x62, 0x42, 0xb5,
		0x22, 0xab, 0xbe, 0xe6, 0x85, 0xfd, 0xa4, 0x42, 0x0f, 0x88, 0x34,
		0xb1, 0x08, 0xc3, 0xbd, 0xae, 0x36, 0x9e, 0xf5, 0x49, 0xfa,
	}

	require.False(t, ed25519.Verify(publicKey, message, sig))
}

func TestGolden(t *testing.T) {
	// sign.input.gz is a selection of test tests from https://ed25519.cr.yp.to/python/sign.input
	file, err := os.Open(path.Join("testdata", "sign.input.gz"))
	require.NoError(t, err)
	defer file.Close()

	testData, err := gzip.NewReader(file)
	require.NoError(t, err)
	defer testData.Close()

	// parse the test data as colon separated values
	r := csv.NewReader(testData)
	r.Comma = ':'

	records, err := r.ReadAll()
	require.NoError(t, err)

	// fields on each input line: sk, pk, m, sm
	for _, record := range records {
		sk, _ := hex.DecodeString(record[0])
		pk, _ := hex.DecodeString(record[1])
		m, _ := hex.DecodeString(record[2])
		sm, _ := hex.DecodeString(record[3])
		s := sm[:ed25519.SignatureSize]

		privateKey := ed25519.NewKeyFromSeed(sk[:ed25519.SeedSize])
		assert.EqualValues(t, sk, privateKey, "different private key")
		assert.EqualValues(t, sk[:ed25519.SeedSize], privateKey.Seed(), "different seed")

		publicKey := privateKey.Public().(ed25519.PublicKey)
		assert.EqualValues(t, pk, publicKey, "different public key")

		sig := ed25519.Sign(privateKey, m)
		assert.Equal(t, s, sig, "different signature")
		assert.True(t, ed25519.Verify(publicKey, m, sig), "invalid signature")
	}
}

func BenchmarkSign(b *testing.B) {
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	data := make([][64]byte, b.N)
	for i := range data {
		if _, err := rand.Read(data[i][:]); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := range data {
		_ = ed25519.Sign(privateKey, data[i][:])
	}
}

func BenchmarkVerify(b *testing.B) {
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	data := make([]struct {
		message []byte
		sig     []byte
	}, b.N)
	for i := range data {
		data[i].message = make([]byte, 64)
		if _, err := rand.Read(data[i].message); err != nil {
			b.Fatal(err)
		}
		data[i].sig = ed25519.Sign(privateKey, data[i].message)
	}

	b.ResetTimer()
	for i := range data {
		_ = ed25519.Verify(publicKey, data[i].message, data[i].sig)
	}
}
