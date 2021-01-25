// Package ed25519 implements the Ed25519 signature algorithm.
//
// These functions are compatible with the particular validation rules around
// edge cases described in IOTA protocol RFC-0028.
// This package is a drop-in replacement for "crypto/ed25519".
package ed25519

import (
	"bytes"
	"crypto"
	cryptorand "crypto/rand"
	"crypto/sha512"
	"errors"
	"io"
	"strconv"

	"filippo.io/edwards25519"
)

const (
	// PublicKeySize is the size, in bytes, of public keys as used in this package.
	PublicKeySize = 32
	// PrivateKeySize is the size, in bytes, of private keys as used in this package.
	PrivateKeySize = 64
	// SignatureSize is the size, in bytes, of signatures generated and verified by this package.
	SignatureSize = 64
	// SeedSize is the size, in bytes, of private key seeds. These are the private key representations used by RFC 8032.
	SeedSize = 32
)

// PublicKey is the type of Ed25519 public keys.
type PublicKey []byte

// Any methods implemented on PublicKey might need to also be implemented on
// PrivateKey, as the latter embeds the former and will expose its methods.

// Equal reports whether pub and x have the same value.
func (pub PublicKey) Equal(x crypto.PublicKey) bool {
	xx, ok := x.(PublicKey)
	if !ok {
		return false
	}
	return bytes.Equal(pub, xx)
}

// PrivateKey is the type of Ed25519 private keys. It implements crypto.Signer.
type PrivateKey []byte

// Public returns the PublicKey corresponding to priv.
func (priv PrivateKey) Public() crypto.PublicKey {
	publicKey := make([]byte, PublicKeySize)
	copy(publicKey, priv[32:])
	return PublicKey(publicKey)
}

// Equal reports whether priv and x have the same value.
func (priv PrivateKey) Equal(x crypto.PrivateKey) bool {
	xx, ok := x.(PrivateKey)
	if !ok {
		return false
	}
	return bytes.Equal(priv, xx)
}

// Seed returns the private key seed corresponding to priv. It is provided for
// interoperability with RFC 8032. RFC 8032's private keys correspond to seeds
// in this package.
func (priv PrivateKey) Seed() []byte {
	seed := make([]byte, SeedSize)
	copy(seed, priv[:32])
	return seed
}

// Sign signs the given message with priv.
// Ed25519 performs two passes over messages to be signed and therefore cannot
// handle pre-hashed messages. Thus opts.HashFunc() must return zero to
// indicate the message hasn't been hashed. This can be achieved by passing
// crypto.Hash(0) as the value for opts.
func (priv PrivateKey) Sign(_ io.Reader, message []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	if opts.HashFunc() != crypto.Hash(0) {
		return nil, errors.New("ed25519: cannot sign hashed message")
	}

	return Sign(priv, message), nil
}

// GenerateKey generates a public/private key pair using entropy from rand.
// If rand is nil, crypto/rand.Reader will be used.
func GenerateKey(rand io.Reader) (PublicKey, PrivateKey, error) {
	if rand == nil {
		rand = cryptorand.Reader
	}

	seed := make([]byte, SeedSize)
	if _, err := io.ReadFull(rand, seed); err != nil {
		return nil, nil, err
	}

	privateKey := NewKeyFromSeed(seed)
	publicKey := make([]byte, PublicKeySize)
	copy(publicKey, privateKey[32:])

	return publicKey, privateKey, nil
}

// NewKeyFromSeed calculates a private key from a seed. It will panic if
// len(seed) is not SeedSize. This function is provided for interoperability
// with RFC 8032. RFC 8032's private keys correspond to seeds in this
// package.
func NewKeyFromSeed(seed []byte) PrivateKey {
	// when NewKeyFromSeed is inlined, the returned signature can be stack-allocated
	privateKey := make([]byte, PrivateKeySize)
	newKeyFromSeed(privateKey, seed)
	return privateKey
}

func newKeyFromSeed(privateKey, seed []byte) {
	if l := len(seed); l != SeedSize {
		panic("ed25519: bad seed length: " + strconv.Itoa(l))
	}

	digest := sha512.Sum512(seed)
	s := new(edwards25519.Scalar).SetBytesWithClamping(digest[:32])
	A := new(edwards25519.Point).ScalarBaseMult(s)

	copy(privateKey, seed)
	copy(privateKey[32:], A.Bytes())
}

// Sign signs the message with privateKey and returns a signature. It will
// panic if len(privateKey) is not PrivateKeySize.
func Sign(privateKey PrivateKey, message []byte) []byte {
	// when Sign is inlined, the returned signature can be stack-allocated
	signature := make([]byte, SignatureSize)
	sign(signature, privateKey, message)
	return signature
}

func sign(signature, privateKey, message []byte) {
	if l := len(privateKey); l != PrivateKeySize {
		panic("ed25519: bad private key length: " + strconv.Itoa(l))
	}

	h := sha512.New()
	h.Write(privateKey[:32])
	digest1 := h.Sum(nil)

	s := new(edwards25519.Scalar).SetBytesWithClamping(digest1[:32])

	h.Reset()
	h.Write(digest1[32:])
	h.Write(message)
	messageDigest := h.Sum(nil)

	rReduced := new(edwards25519.Scalar).SetUniformBytes(messageDigest[:])
	R := new(edwards25519.Point).ScalarBaseMult(rReduced)

	encodedR := R.Bytes()

	h.Reset()
	h.Write(encodedR[:])
	h.Write(privateKey[32:])
	h.Write(message)
	hramDigest := h.Sum(nil)

	kReduced := new(edwards25519.Scalar).SetUniformBytes(hramDigest[:])
	S := new(edwards25519.Scalar).MultiplyAdd(kReduced, s, rReduced)

	copy(signature[:], encodedR[:])
	copy(signature[32:], S.Bytes())
}

// Verify reports whether sig is a valid signature of message by publicKey.
// It uses precisely-specified validation criteria (ZIP 215) suitable for use in consensus-critical contexts.
func Verify(publicKey PublicKey, message, sig []byte) bool {
	if len(publicKey) != PublicKeySize {
		return false
	}
	if len(sig) != SignatureSize || sig[63]&224 != 0 {
		return false
	}

	// ZIP215: this works because SetBytes does not check that encodings are canonical
	A, err := new(edwards25519.Point).SetBytes(publicKey)
	if err != nil {
		return false
	}
	A.Negate(A)

	h := sha512.New()
	h.Write(sig[:32])
	h.Write(publicKey[:])
	h.Write(message)
	var digest [64]byte
	h.Sum(digest[:0])

	hReduced := new(edwards25519.Scalar).SetUniformBytes(digest[:])

	// ZIP215: this works because SetBytes does not check that encodings are canonical
	checkR, err := new(edwards25519.Point).SetBytes(sig[:32])
	if err != nil {
		return false
	}

	// https://tools.ietf.org/html/rfc8032#section-5.1.7 requires that s be in
	// the range [0, order) in order to prevent signature malleability
	s, err := new(edwards25519.Scalar).SetCanonicalBytes(sig[32:])
	if err != nil {
		return false
	}

	R := new(edwards25519.Point).VarTimeDoubleScalarBaseMult(hReduced, A, s)

	// ZIP215: We want to check [8](R - checkR) == 0
	p := new(edwards25519.Point).Subtract(R, checkR)     // p = R - checkR
	p.Add(p, p)                                          // p = [2]p
	p.Add(p, p)                                          // p = [4]p
	p.Add(p, p)                                          // p = [8]p
	return p.Equal(edwards25519.NewIdentityPoint()) == 1 // p == 0
}
