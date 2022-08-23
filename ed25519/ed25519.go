// Package ed25519 implements an Ed25519 Verify function for use in consensus-critical contexts.
// The Verify function is compatible with the particular validation rules around
// edge cases described in IOTA protocol RFC-0028.
package ed25519

import (
	"crypto/ed25519"
	"crypto/sha512"

	// We need to use this package to have access to low-level edwards25519 operations.
	//
	// Excerpt from the docs:
	// https://pkg.go.dev/crypto/ed25519/internal/edwards25519?utm_source=godoc
	//
	// However, developers who do need to interact with low-level edwards25519
	// operations can use filippo.io/edwards25519,
	// an extended version of this package repackaged as an importable module.
	"filippo.io/edwards25519"
)

// Verify reports whether sig is a valid signature of message by publicKey.
// It uses precisely-specified validation criteria (ZIP 215) suitable for use in consensus-critical contexts.
func Verify(publicKey ed25519.PublicKey, message, sig []byte) bool {
	if len(publicKey) != ed25519.PublicKeySize {
		return false
	}
	if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
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

	hReduced, err := new(edwards25519.Scalar).SetUniformBytes(digest[:])
	if err != nil {
		panic(err)
	}

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
	p := new(edwards25519.Point).Subtract(R, checkR) // p = R - checkR
	p.Add(p, p)                                      // p = [2]p
	p.Add(p, p)                                      // p = [4]p
	p.Add(p, p)                                      // p = [8]p

	return p.Equal(edwards25519.NewIdentityPoint()) == 1 // p == 0
}
