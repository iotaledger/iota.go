package wallet

import (
	"crypto/ed25519"
	"fmt"

	"github.com/iotaledger/hive.go/crypto"
	iotago "github.com/iotaledger/iota.go/v4"
)

// Account represents an account.
type Account interface {
	// ID returns the accountID.
	ID() iotago.AccountID

	// Address returns the account address.
	Address() iotago.Address

	// PrivateKey returns the account private key for signing.
	PrivateKey() ed25519.PrivateKey
}

var _ Account = &Ed25519Account{}

// Ed25519Account is an account that uses an Ed25519 key pair.
type Ed25519Account struct {
	accountID  iotago.AccountID
	privateKey ed25519.PrivateKey
}

// NewEd25519Account creates a new Ed25519Account.
func NewEd25519Account(accountID iotago.AccountID, privateKey ed25519.PrivateKey) *Ed25519Account {
	return &Ed25519Account{
		accountID:  accountID,
		privateKey: privateKey,
	}
}

// ID returns the accountID.
func (e *Ed25519Account) ID() iotago.AccountID {
	return e.accountID
}

// Address returns the account address.
func (e *Ed25519Account) Address() iotago.Address {
	ed25519PubKey, ok := e.privateKey.Public().(ed25519.PublicKey)
	if !ok {
		panic("invalid public key type")
	}

	return iotago.Ed25519AddressFromPubKey(ed25519PubKey)
}

// PrivateKey returns the account private key for signing.
func (e *Ed25519Account) PrivateKey() ed25519.PrivateKey {
	return e.privateKey
}

func AccountFromParams(accountHex string, privateKey string) Account {
	accountID, err := iotago.AccountIDFromHexString(accountHex)
	if err != nil {
		panic(fmt.Sprintln("invalid accountID hex string", err))
	}

	privKey, err := crypto.ParseEd25519PrivateKeyFromString(privateKey)
	if err != nil {
		panic(fmt.Sprintln("invalid ed25519 private key string", err))
	}

	return NewEd25519Account(accountID, privKey)
}
