package iotago

import (
	"crypto"
	"crypto/ed25519"
	"fmt"

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
	// SignerUIDForAddress returns the signer unique identifier for a given address.
	// This can be used to identify the uniqueness of the signer in the unlocks (e.g. unique public key).
	SignerUIDForAddress(addr Address) (Identifier, error)
	// Sign produces the signature for the given message.
	Sign(addr Address, msg []byte) (signature Signature, err error)
	// EmptySignatureForAddress returns an empty signature for the given address.
	// This can be used to calculate the WorkScore of transactions without actually signing the transaction.
	EmptySignatureForAddress(addr Address) (signature Signature, err error)
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
	case *ImplicitAccountCreationAddress:
		panic("ImplicitAccountCreationAddress is not allowed in restricted addresses")
	default:
		panic(fmt.Sprintf("address type %T is not supported in the address signer since it only handles addresses backed by keypairs", addr))
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

// privateKeyForAddress returns the private key for the given address.
func (s *InMemoryAddressSigner) privateKeyForAddress(addr Address) (crypto.PrivateKey, error) {
	privateKeyForEd25519Address := func(edAddr DirectUnlockableAddress) (ed25519.PrivateKey, error) {
		maybePrvKey, ok := s.addrKeys[edAddr.Key()]
		if !ok {
			return nil, ierrors.Errorf("can't sign message for Ed25519 address: %w", ErrAddressKeysNotMapped)
		}

		prvKey, ok := maybePrvKey.(ed25519.PrivateKey)
		if !ok {
			return nil, ierrors.WithMessagef(ErrAddressKeysWrongType, "Ed25519 address needs to have a %T private key mapped but got %T", ed25519.PrivateKey{}, maybePrvKey)
		}

		return prvKey, nil
	}

	switch address := addr.(type) {
	case *Ed25519Address:
		return privateKeyForEd25519Address(address)

	case *RestrictedAddress:
		switch underlyingAddr := address.Address.(type) {
		case *Ed25519Address:
			return privateKeyForEd25519Address(underlyingAddr)
		default:
			panic(fmt.Sprintf("underlying address type %T in restricted address is not supported in the address signer since it only handles addresses backed by keypairs", addr))
		}

	case *ImplicitAccountCreationAddress:
		return privateKeyForEd25519Address(address)

	default:
		panic(fmt.Sprintf("address type %T is not supported in the address signer since it only handles addresses backed by keypairs", addr))
	}
}

// SignerUIDForAddress returns the signer unique identifier for a given address.
// This can be used to identify the uniqueness of the signer in the unlocks.
func (s *InMemoryAddressSigner) SignerUIDForAddress(addr Address) (Identifier, error) {
	prvKey, err := s.privateKeyForAddress(addr)
	if err != nil {
		return EmptyIdentifier, ierrors.Errorf("can't get private key for address: %w", err)
	}

	ed25519PrvKey, ok := prvKey.(ed25519.PrivateKey)
	if !ok {
		return EmptyIdentifier, ierrors.WithMessagef(ErrAddressKeysWrongType, "Ed25519 address needs to have a %T private key mapped but got %T", ed25519.PrivateKey{}, prvKey)
	}

	// the UID is the blake2b 256 hash of the public key
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	return IdentifierFromData(ed25519PrvKey.Public().(ed25519.PublicKey)), nil
}

func (s *InMemoryAddressSigner) Sign(addr Address, msg []byte) (signature Signature, err error) {
	prvKey, err := s.privateKeyForAddress(addr)
	if err != nil {
		return nil, ierrors.Wrap(err, "can't sign message for address")
	}

	ed25519PrvKey, ok := prvKey.(ed25519.PrivateKey)
	if !ok {
		return nil, ierrors.WithMessagef(ErrAddressKeysWrongType, "Ed25519 address needs to have a %T private key mapped but got %T", ed25519.PrivateKey{}, prvKey)
	}

	ed25519Sig := &Ed25519Signature{}
	copy(ed25519Sig.Signature[:], ed25519.Sign(ed25519PrvKey, msg))
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	copy(ed25519Sig.PublicKey[:], ed25519PrvKey.Public().(ed25519.PublicKey))

	return ed25519Sig, nil
}

// EmptySignatureForAddress returns an empty signature for the given address.
// This can be used to calculate the WorkScore of transactions without actually signing the transaction.
func (s *InMemoryAddressSigner) EmptySignatureForAddress(addr Address) (signature Signature, err error) {
	switch address := addr.(type) {
	case *Ed25519Address:
		return &Ed25519Signature{}, nil

	case *RestrictedAddress:
		switch address.Address.(type) {
		case *Ed25519Address:
			return &Ed25519Signature{}, nil
		default:
			panic(fmt.Sprintf("underlying address type %T in restricted address is not supported in the address signer since it only handles addresses backed by keypairs", addr))
		}
	case *ImplicitAccountCreationAddress:
		return &Ed25519Signature{}, nil

	default:
		panic(fmt.Sprintf("address type %T is not supported in the address signer since it only handles addresses backed by keypairs", addr))
	}
}
