package iotago

import (
	"crypto/ed25519"

	"github.com/iotaledger/hive.go/ierrors"
)

var (
	// ErrAddressKeysNotMapped gets returned if the needed keys to sign a message are absent/not mapped.
	ErrAddressKeysNotMapped = ierrors.New("key(s) for address not mapped")
	// ErrAddressKeysWrongType gets returned if the specified keys to sign a message for a given address are of the wrong type.
	ErrAddressKeysWrongType = ierrors.New("key(s) for address are of wrong type")
)

// AddressSigner produces signatures for messages which get verified against a given address.
type AddressSigner interface {
	// Sign produces the signature for the given message.
	Sign(addr Address, msg []byte) (signature Signature, err error)
}

// AddressSignerFunc implements the AddressSigner interface.
type AddressSignerFunc func(addr Address, msg []byte) (signature Signature, err error)

func (s AddressSignerFunc) Sign(addr Address, msg []byte) (signature Signature, err error) {
	return s(addr, msg)
}

// AddressKeys pairs an address and its source key(s).
type AddressKeys struct {
	// The target address.
	Address Address `json:"address"`
	// The signing keys.
	Keys interface{} `json:"keys"`
}

// NewAddressKeysForEd25519Address returns new AddressKeys for Ed25519Address.
func NewAddressKeysForEd25519Address(addr *Ed25519Address, prvKey ed25519.PrivateKey) AddressKeys {
	return AddressKeys{Address: addr, Keys: prvKey}
}

// NewAddressKeysForImplicitAccountCreationAddress returns new AddressKeys for ImplicitAccountCreationAddress.
func NewAddressKeysForImplicitAccountCreationAddress(addr *ImplicitAccountCreationAddress, prvKey ed25519.PrivateKey) AddressKeys {
	return AddressKeys{Address: addr, Keys: prvKey}
}

// NewAddressKeysForRestrictedEd25519Address returns new AddressKeys for a restricted Ed25519Address.
func NewAddressKeysForRestrictedEd25519Address(addr *RestrictedAddress, prvKey ed25519.PrivateKey) (AddressKeys, error) {
	switch addr.Address.(type) {
	case *Ed25519Address:
		return AddressKeys{Address: addr, Keys: prvKey}, nil
	default:
		return AddressKeys{}, ierrors.Wrapf(ErrUnknownAddrType, "unknown underlying address type %T in restricted address", addr)
	}
}

// NewInMemoryAddressSigner creates a new InMemoryAddressSigner holding the given AddressKeys.
func NewInMemoryAddressSigner(addrKeys ...AddressKeys) AddressSigner {
	ss := &InMemoryAddressSigner{
		addrKeys: map[string]interface{}{},
	}
	for _, c := range addrKeys {
		ss.addrKeys[c.Address.Key()] = c.Keys
	}

	return ss
}

// NewInMemoryAddressSignerFromEd25519PrivateKey creates a new InMemoryAddressSigner
// for the Ed25519Address derived from the public key of the given private key
// as well as the related ImplicitAccountCreationAddress.
func NewInMemoryAddressSignerFromEd25519PrivateKeys(privKeys ...ed25519.PrivateKey) AddressSigner {
	addressKeys := make([]AddressKeys, 0)
	for _, privKey := range privKeys {
		//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
		pubKey := privKey.Public().(ed25519.PublicKey)

		ed25519Address := Ed25519AddressFromPubKey(pubKey)
		addressKeys = append(addressKeys, NewAddressKeysForEd25519Address(ed25519Address, privKey))

		implicitAccountCreationAddress := ImplicitAccountCreationAddressFromPubKey(pubKey)
		addressKeys = append(addressKeys, NewAddressKeysForImplicitAccountCreationAddress(implicitAccountCreationAddress, privKey))
	}

	// add both address types for simplicity
	return NewInMemoryAddressSigner(addressKeys...)
}

// InMemoryAddressSigner implements AddressSigner by holding keys simply in-memory.
type InMemoryAddressSigner struct {
	addrKeys map[string]interface{}
}

func (s *InMemoryAddressSigner) Sign(addr Address, msg []byte) (signature Signature, err error) {

	signatureForEd25519Address := func(edAddr DirectUnlockableAddress, msg []byte) (signature Signature, err error) {
		maybePrvKey, ok := s.addrKeys[edAddr.Key()]
		if !ok {
			return nil, ierrors.Errorf("can't sign message for Ed25519 address: %w", ErrAddressKeysNotMapped)
		}

		prvKey, ok := maybePrvKey.(ed25519.PrivateKey)
		if !ok {
			return nil, ierrors.Wrapf(ErrAddressKeysWrongType, "Ed25519 address needs to have a %T private key mapped but got %T", ed25519.PrivateKey{}, maybePrvKey)
		}

		ed25519Sig := &Ed25519Signature{}
		copy(ed25519Sig.Signature[:], ed25519.Sign(prvKey, msg))
		//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
		copy(ed25519Sig.PublicKey[:], prvKey.Public().(ed25519.PublicKey))

		return ed25519Sig, nil
	}

	switch address := addr.(type) {
	case *Ed25519Address:
		return signatureForEd25519Address(address, msg)

	case *RestrictedAddress:
		switch underlyingAddr := address.Address.(type) {
		case *Ed25519Address:
			return signatureForEd25519Address(underlyingAddr, msg)
		default:
			return nil, ierrors.Wrapf(ErrUnknownAddrType, "unknown underlying address type %T in restricted address", addr)
		}
	case *ImplicitAccountCreationAddress:
		return signatureForEd25519Address(address, msg)

	default:
		return nil, ierrors.Wrapf(ErrUnknownAddrType, "type %T", addr)
	}
}

// EmptyAddressSigner returns an empty signature for the given address.
// This can be used to calculate the WorkScore of transactions without actually signing the transaction.
type EmptyAddressSigner struct{}

func (s *EmptyAddressSigner) Sign(addr Address, _ []byte) (signature Signature, err error) {
	switch address := addr.(type) {
	case *Ed25519Address:
		return &Ed25519Signature{}, nil

	case *RestrictedAddress:
		switch address.Address.(type) {
		case *Ed25519Address:
			return &Ed25519Signature{}, nil
		default:
			return nil, ierrors.Wrapf(ErrUnknownAddrType, "unknown underlying address type %T in restricted address", addr)
		}
	case *ImplicitAccountCreationAddress:
		return &Ed25519Signature{}, nil

	default:
		return nil, ierrors.Wrapf(ErrUnknownAddrType, "type %T", addr)
	}
}
