// Package multisig provides functionality for creating multisig bundles.
package multisig

import (
	. "github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/checksum"
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/guards"
	. "github.com/iotaledger/iota.go/guards/validators"
	"github.com/iotaledger/iota.go/kerl"
	"github.com/iotaledger/iota.go/signing"
	. "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
	"math"
	"strings"
	"time"
)

// MultisigInput represents a multisig input.
type MultisigInput struct {
	Address     Hash
	Balance     uint64
	SecuritySum int64
}

// NewMultisig creates a new Multisig object which uses the given API object.
func NewMultisig(api *API) *Multisig {
	m := &Multisig{API: api}
	return m
}

// Multisig encapsulates the processes of generating and validating multisig components.
type Multisig struct {
	Address MultisigAddress
	API     *API
}

// Key gets the key value of a seed.
func (m *Multisig) Key(seed Trytes, index uint64, security SecurityLevel, spongeFunc ...SpongeFunction) (Trytes, error) {
	h := GetSpongeFunc(spongeFunc, kerl.NewKerl)

	subseed, err := signing.Subseed(seed, index, h)
	if err != nil {
		return "", err
	}

	keyTrits, err := signing.Key(subseed, security, h)
	if err != nil {
		return "", err
	}

	return MustTritsToTrytes(keyTrits), nil
}

// Digest gets the digest of a seed under the given index and security.
func (m *Multisig) Digest(seed Trytes, index uint64, security SecurityLevel, spongeFunc ...SpongeFunction) (Trytes, error) {
	h := GetSpongeFunc(spongeFunc, kerl.NewKerl)

	subseed, err := signing.Subseed(seed, index, h)
	if err != nil {
		return "", err
	}

	keyTrits, err := signing.Key(subseed, security, h)
	if err != nil {
		return "", err
	}

	digestTrits, err := signing.Digests(keyTrits, h)
	if err != nil {
		return "", err
	}

	return MustTritsToTrytes(digestTrits), nil
}

// ValidateAddress validates the given multisig address.
func (m *Multisig) ValidateAddress(addr Trytes, digests []Trytes, spongeFunc ...SpongeFunction) (bool, error) {
	h := GetSpongeFunc(spongeFunc, kerl.NewKerl)

	for i := range digests {
		digestTrits, err := TrytesToTrits(digests[i])
		if err != nil {
			return false, err
		}
		if err := h.Absorb(digestTrits); err != nil {
			return false, err
		}
	}

	addressTrits, err := h.Squeeze(HashTrinarySize)
	if err != nil {
		return false, err
	}
	return MustTritsToTrytes(addressTrits) == addr, nil
}

// InitiateTransfer prepares a transfer by generating the bundle with the corresponding cosigner transactions.
// Does not contain signatures.
func (m *Multisig) InitiateTransfer(input MultisigInput, transfers bundle.Transfers, remainderAddress *Trytes) (bundle.Bundle, error) {
	if err := validateMultisigInput(input); err != nil {
		return nil, err
	}
	if err := Validate(ValidateTransfers(transfers...)); err != nil {
		return nil, err
	}
	if remainderAddress != nil {
		if err := Validate(ValidateHashes(*remainderAddress)); err != nil {
			return nil, err
		}
	}

	for i := range transfers {
		transfer := &transfers[i]
		addr, err := checksum.RemoveChecksum(transfer.Address)
		if err != nil {
			return nil, err
		}
		transfer.Address = addr
		transfer.Tag = bundle.PadTag(transfer.Tag)
	}

	if input.Balance > 0 {
		return createBundle(input, transfers, remainderAddress)
	}
	balances, err := m.API.GetBalances(Hashes{input.Address}, 100)
	if err != nil {
		return nil, err
	}

	inputWithBalance := MultisigInput{
		Balance:     balances.Balances[0],
		Address:     input.Address,
		SecuritySum: input.SecuritySum,
	}

	return createBundle(inputWithBalance, transfers, remainderAddress)

}

func validateMultisigInput(input MultisigInput) error {
	if err := Validate(
		ValidateSecurityLevel(SecurityLevel(input.SecuritySum)),
		ValidateHashes(input.Address),
	); err != nil {
		return err
	}
	if input.Balance <= 0 {
		return ErrInvalidInput
	}
	return nil
}

// AddSignature returns cosigner signatures for the transactions.
func (m *Multisig) AddSignature(bndl bundle.Bundle, inputAddr Trytes, key Trytes) ([]Trytes, error) {
	security := len(key) / SignatureMessageFragmentSizeInTrytes
	keyTrits, err := TrytesToTrits(key)
	if err != nil {
		return nil, err
	}

	signedFrags := []Trytes{}

	numSignedTxs := 0

	for i := 0; i < len(bndl); i++ {
		tx := &bndl[i]
		if tx.Address != inputAddr {
			continue
		}

		if !guards.IsEmptyTrytes(tx.SignatureMessageFragment) {
			numSignedTxs++
			continue
		}

		bundleHash := tx.Bundle
		firstFrag := keyTrits[0:6561]
		normalizedBundleHash := signing.NormalizedBundleHash(bundleHash)
		normalizedBundleFrags := [][]int8{}

		for k := 0; k < 3; k++ {
			normalizedBundleFrags[k] = normalizedBundleHash[k*27 : (k+1)*27]
		}

		firstBundleFrag := normalizedBundleFrags[numSignedTxs%3]
		firstSignedFrag, err := signing.SignatureFragment(firstBundleFrag, firstFrag)
		if err != nil {
			return nil, err
		}

		signedFrags = append(signedFrags, MustTritsToTrytes(firstSignedFrag))

		for j := 1; j < security; j++ {
			nextFrag := keyTrits[6561*j : (j+1)*6561]
			nextBundleFrag := normalizedBundleFrags[(numSignedTxs+j)%3]
			nextSignedFrag, err := signing.SignatureFragment(nextBundleFrag, nextFrag)
			if err != nil {
				return nil, err
			}
			signedFrags = append(signedFrags, MustTritsToTrytes(nextSignedFrag))
		}
	}

	return signedFrags, nil
}

func createBundle(input MultisigInput, transfers bundle.Transfers, remainderAddress *Trytes) (bundle.Bundle, error) {
	if remainderAddress != nil {
		if err := Validate(ValidateHashes(*remainderAddress)); err != nil {
			return nil, err
		}
	}
	bndl := bundle.Bundle{}
	sigFrags := []Trytes{}
	totalBalance := input.Balance
	var totalValue uint64
	tag := strings.Repeat("9", 27)

	now := time.Now().UnixNano() / int64(time.Second)
	for i := 0; i < len(transfers); i++ {
		transfer := &transfers[i]
		sigFragLength := 1

		if len(transfers[i].Message) > SignatureMessageFragmentSizeInTrytes {
			sigFragLength += int(math.Floor(float64(len(transfer.Message)) / float64(SignatureMessageFragmentSizeInTrytes)))

			msg := transfer.Message

			for len(msg) > 0 {
				frag := msg[:SignatureMessageFragmentSizeInTrytes]
				msg = msg[SignatureMessageFragmentSizeInTrytes:len(msg)]

				for j := 0; len(frag) < 2187; j++ {
					frag += "9"
				}

				sigFrags = append(sigFrags, frag)
			}
		} else {
			frag := transfer.Message

			for j := 0; len(frag) < SignatureMessageFragmentSizeInTrytes; j++ {
				frag += "9"
			}

			sigFrags = append(sigFrags, frag)
		}

		tag = Pad(transfer.Tag, TagTrinarySize/3)

		bndl = bundle.AddEntry(bndl, bundle.BundleEntry{
			Length:  uint64(sigFragLength),
			Address: transfer.Address[:HashTrytesSize],
			Value:   int64(transfer.Value),
			Tag:     transfer.Tag, Timestamp: uint64(now),
		})

		totalValue += transfer.Value
	}

	if totalBalance > 0 {
		sub := 0 - int64(totalBalance)

		bndl = bundle.AddEntry(bndl, bundle.BundleEntry{
			Length:  uint64(input.SecuritySum),
			Address: input.Address,
			Value:   sub, Tag: tag,
			Timestamp: uint64(now),
		})
	}

	if totalValue > totalBalance {
		return nil, ErrInsufficientBalance
	}

	if totalBalance > totalValue {
		if remainderAddress == nil {
			return nil, ErrNoRemainderSpecified
		}

		remainder := int64(totalBalance) - int64(totalValue)

		bndl = bundle.AddEntry(bndl, bundle.BundleEntry{
			Length: 1, Address: *remainderAddress,
			Value: remainder, Tag: tag, Timestamp: uint64(now),
		})
	}

	bndl, err := bundle.Finalize(bndl)
	if err != nil {
		return nil, err
	}

	index := 0
	for i := range bndl {
		if bndl[i].Value < 0 {
			index = i
			break
		}
	}

	return bundle.AddTrytes(bndl, sigFrags, index), nil
}
