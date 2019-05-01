// Package mam provides functions for creating Masked Authentication Messaging messages
package mam

import (
	"reflect"

	"github.com/pkg/errors"

	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/curl/hamming"
	. "github.com/iotaledger/iota.go/guards/validators"
	. "github.com/iotaledger/iota.go/merkle"
	signing "github.com/iotaledger/iota.go/signing/legacy"
	sponge "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
)

var (
	ErrPayloadTooShort        = errors.New("payload too short")
	ErrMerkleRootDoesNotMatch = errors.New("Merkle root does not match")
	ErrWrongSecurityLevel     = errors.New("wrong security level")
)

// MAMInitEncryption initializes the encryption/decryption state for a MAM session.
//
//	sideKey is the encryption/decryption key
//	merkleRoot is the merkle root
//	spongeFunc is the spongeFunction instance used for encryption/decryption
func MAMInitEncryption(sideKey Trits, merkleRoot Trits, spongeFunc sponge.SpongeFunction) {
	spongeFunc.Absorb(sideKey)
	spongeFunc.Absorb(merkleRoot[:HashTrinarySize])
}

// PayloadMinLength computes the minimum length of a payload.
//
//	messageLength is the length of the message
//	merkleTreeLength is the length of the merkle tree
//	index is the leaf index
//	security is the security level used to generate addresses
//
// Returns the minimum length of the payload
func PayloadMinLength(messageLength, merkleTreeLength, index uint64, security SecurityLevel) uint64 {
	siblingNumber := MerkleDepth(merkleTreeLength/HashTrinarySize) - 1
	return EncodedLength(int64(index)) + EncodedLength(int64(messageLength)) +
		HashTrinarySize + messageLength + HashTrytesSize +
		uint64(security)*ISSKeyLength + EncodedLength(int64(siblingNumber)) +
		siblingNumber*HashTrinarySize
}

// MAMCreate creates a signed, encrypted payload from a message.
//
//	payloadLength is the length of the payload
//	message is the message to encrypt
//	sideKey is the encryption key
//	merkleTree is the merkle tree
//	merkleTreeLength is the length of the merkle tree
//	leafCount is the number of leaves of the merkle tree
//	index is the leaf index
//	nextRoot is the merkle root of the next MAM payload
//	start is the offset used to generate addresses
//	seed is the seed used to generate addresses - Not sent over the network
//	security is the security level used to generate addresses
//
// Returns:
//  payload is the payload of the encrypted message
//  payloadMinLength is the length of the payload needed to encrypt the message
//	err is the error message
func MAMCreate(payloadLength uint64,
	message Trytes, sideKey Trytes,
	merkleTree Trits, merkleTreeLength uint64,
	leafCount uint64, index uint64,
	nextRoot Trits, start uint64,
	seed Trytes, security SecurityLevel) (Trits, uint64, error) {

	err := Validate(ValidateSecurityLevel(security))
	if err != nil {
		return nil, 0, err
	}

	messageTrits, err := TrytesToTrits(message)
	if err != nil {
		return nil, 0, err
	}
	messageLength := len(messageTrits)

	sideKey = Pad(sideKey, 81)
	sideKeyTrits, err := TrytesToTrits(sideKey)
	if err != nil {
		return nil, 0, err
	}

	siblingsNumber := MerkleDepth(merkleTreeLength/HashTrinarySize) - 1

	indexTrits, encIndexLength, err := EncodeInt64(int64(index))
	if err != nil {
		return nil, 0, err
	}

	messageLengthTrits, encMessageLengthLenght, err := EncodeInt64(int64(messageLength))
	if err != nil {
		return nil, 0, err
	}

	siblingsNumberTrits, encSiblingsNumberLength, err := EncodeInt64(int64(siblingsNumber))
	if err != nil {
		return nil, 0, err
	}

	signatureLength := uint64(security) * ISSKeyLength
	payloadMinLength := encIndexLength + encMessageLengthLenght + HashTrinarySize +
		uint64(messageLength) + HashTrytesSize + signatureLength +
		encSiblingsNumberLength + (siblingsNumber * HashTrinarySize)

	if payloadLength < payloadMinLength {
		return nil, 0, errors.Wrapf(ErrPayloadTooShort, "needed %d, given %d", payloadMinLength, payloadLength)
	}

	var offset uint64

	payload := make(Trits, payloadLength)

	encCurl := curl.NewCurlP27().(*curl.Curl)
	MAMInitEncryption(sideKeyTrits, merkleTree, encCurl)

	// encode index to payload
	copy(payload[offset:offset+encIndexLength], indexTrits)
	offset += encIndexLength

	// encode message length to payload
	copy(payload[offset:offset+encMessageLengthLenght], messageLengthTrits)
	offset += encMessageLengthLenght
	encCurl.Absorb(payload[:offset])

	// encrypt next root to payload
	Mask(payload[offset:], nextRoot, HashTrinarySize, encCurl)
	offset += HashTrinarySize

	// encrypt message to payload
	Mask(payload[offset:], messageTrits, uint64(messageLength), encCurl)
	offset += uint64(messageLength)

	// encrypt nonce to payload
	c := curl.NewCurlP27().(*curl.Curl)
	copy(c.State, encCurl.State)

	Hamming(c, 0, HashTrytesSize, int(security))

	Mask(payload[offset:], c.State, HashTrytesSize, encCurl)
	offset += HashTrytesSize

	// encrypt signature to payload
	c.Reset()
	subSeed, err := signing.Subseed(seed, start+index, c)
	if err != nil {
		return nil, 0, err
	}

	key, err := signing.Key(subSeed, security, c)
	if err != nil {
		return nil, 0, err
	}

	signatureFragment, err := signing.SignatureFragment(encCurl.State, key, 0, c)
	copy(payload[offset:offset+signatureLength], signatureFragment)
	offset += signatureLength

	// encrypt siblings number to payload
	copy(payload[offset:offset+encSiblingsNumberLength], siblingsNumberTrits)
	offset += encSiblingsNumberLength

	// encrypt siblings to payload
	MerkleBranch(merkleTree, payload[offset:], merkleTreeLength, siblingsNumber+1, index, leafCount)
	offset += siblingsNumber * HashTrinarySize

	toMask := signatureLength + encSiblingsNumberLength + siblingsNumber*HashTrinarySize
	Mask(payload[offset-toMask:], payload[offset-toMask:], toMask, encCurl)

	encCurl.Reset()

	return payload, payloadMinLength, nil
}

// MAMParse decrypts, parses and validates an encrypted payload.
//
//	payload is the payload
//	payloadLength is the length of the payload
//	sideKey is the decryption key
//	root is the merkle root
//
// Returns:
//  parsedIndex is the parsed index of the message
//  parsedNextRoot is the merkle root of the next MAM payload
//	parsedMessage is the decrypted message
//	parsedSecurity is the parsed security level of the message
//	err is the error message
func MAMParse(payload Trits, payloadLength uint64, sideKey Trytes, root Trits) (uint64, Trytes, Trytes, SecurityLevel, error) {
	var offset uint64 = 0

	sideKey = Pad(sideKey, 81)
	sideKeyTrits, err := TrytesToTrits(sideKey)
	if err != nil {
		return 0, "", "", SecurityLevel(0), err
	}

	encCurl := curl.NewCurlP27().(*curl.Curl)
	MAMInitEncryption(sideKeyTrits, root, encCurl)

	// decode index from payload
	if offset >= payloadLength {
		return 0, "", "", SecurityLevel(0), ErrPayloadTooShort
	}

	index, encIndexLength, err := DecodeInt64(payload[offset:])
	if err != nil {
		return 0, "", "", SecurityLevel(0), err
	}
	parsedIndex := uint64(index)
	offset += encIndexLength

	// decode message length from payload
	if offset >= payloadLength {
		return 0, "", "", SecurityLevel(0), ErrPayloadTooShort
	}

	messageLength, encMessageLengthLenght, err := DecodeInt64(payload[offset:])
	if err != nil {
		return 0, "", "", SecurityLevel(0), err
	}
	offset += encMessageLengthLenght

	if offset >= payloadLength {
		return 0, "", "", SecurityLevel(0), ErrPayloadTooShort
	}
	encCurl.Absorb(payload[:offset])

	// decrypt next root from payload
	nextRoot := Unmask(payload[offset:], HashTrinarySize, encCurl)
	parsedNextRoot := MustTritsToTrytes(nextRoot)
	offset += HashTrinarySize

	// decrypt message from payload
	if offset >= payloadLength {
		return 0, "", "", SecurityLevel(0), ErrPayloadTooShort
	}
	message := Unmask(payload[offset:], uint64(messageLength), encCurl)
	parsedMessage := MustTritsToTrytes(message)

	offset += uint64(messageLength)

	// decrypt nonce from payload
	if offset >= payloadLength {
		return 0, "", "", SecurityLevel(0), ErrPayloadTooShort
	}
	// Information is returned in encCurl.State
	Unmask(payload[offset:], uint64(HashTrytesSize), encCurl)
	offset += HashTrytesSize

	// get security back from state
	if offset >= payloadLength {
		return 0, "", "", SecurityLevel(0), ErrPayloadTooShort
	}
	hash := make(Trits, HashTrinarySize)
	copy(hash, encCurl.State[:HashTrinarySize])

	parsedSecurity, err := signing.GetSecurityLevel(hash)
	if err != nil {
		return 0, "", "", SecurityLevel(0), err
	}
	copy(payload[offset:], Unmask(payload[offset:], payloadLength-offset, encCurl))

	if parsedSecurity == 0 {
		return 0, "", "", SecurityLevel(0), ErrWrongSecurityLevel
	}

	// decrypt signature from payload
	encCurl.Reset()
	digest, err := signing.Digest(hash, payload[offset:offset+(uint64(parsedSecurity)*ISSKeyLength)], 0, encCurl)

	// complete the address
	address, err := signing.Address(digest, encCurl)
	offset += uint64(parsedSecurity) * ISSKeyLength

	// decrypt siblings number from payload
	if offset >= payloadLength {
		return 0, "", "", SecurityLevel(0), ErrPayloadTooShort
	}

	siblingsNumber, encSiblingsNumberLength, err := DecodeInt64(payload[offset:])
	offset += encSiblingsNumberLength

	// get merkle root from siblings from payload
	if siblingsNumber != 0 {
		if offset >= payloadLength {
			return 0, "", "", SecurityLevel(0), ErrPayloadTooShort
		}
		address, err = MerkleRoot(address, payload[offset:], uint64(siblingsNumber), parsedIndex, encCurl)
	}

	// check merkle root with the given root
	if !reflect.DeepEqual(address, root[:HashTrinarySize]) {
		return 0, "", "", SecurityLevel(0), ErrMerkleRootDoesNotMatch
	}

	return parsedIndex, parsedNextRoot, parsedMessage, parsedSecurity, nil
}
