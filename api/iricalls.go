package api

import (
	"github.com/iotaledger/iota.go/checksum"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/guards"
	. "github.com/iotaledger/iota.go/guards/validators"
	"github.com/iotaledger/iota.go/pow"
	. "github.com/iotaledger/iota.go/trinary"
	"strconv"
)

// AddNeighbors adds a list of neighbors to the connected IRI node.
// Assumes addNeighbors command is available on the node.
// AddNeighbors has only a temporary effect until the node relaunches.
func (api *API) AddNeighbors(uris ...string) (int64, error) {
	if err := Validate(ValidateURIs(uris...), ValidateNonEmptyStrings(ErrInvalidURI, uris...)); err != nil {
		return 0, err
	}
	cmd := &AddNeighborsCommand{Command: Command{AddNeighborsCmd}, URIs: uris}
	rsp := &AddNeighborsResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return 0, err
	}
	return rsp.AddedNeighbors, nil
}

// AttachToTangle performs the Proof-of-Work required to attach a transaction to the Tangle by
// calling the attachToTangle IRI API command. Returns a list of transaction trytes and overwrites the following
// fields: Hash, Nonce, AttachmentTimestamp, AttachmentTimestampLowerBound, AttachmentTimestampUpperBound.
//
// If a Proof-of-Work function is supplied when composing the API, then that function is used
// instead of using the connected node.
func (api *API) AttachToTangle(trunkTxHash Hash, branchTxHash Hash, mwm uint64, trytes []Trytes) ([]Trytes, error) {
	if api.localProofOfWorkFunc != nil {
		return pow.DoPoW(trunkTxHash, branchTxHash, trytes, mwm, api.localProofOfWorkFunc)
	}

	if err := Validate(ValidateTransactionTrytes(trytes...)); err != nil {
		return nil, err
	}

	if !IsTransactionHash(trunkTxHash) {
		return nil, ErrInvalidTrunkTransaction
	}

	if !IsTransactionHash(branchTxHash) {
		return nil, ErrInvalidBranchTransaction
	}

	cmd := &AttachToTangleCommand{
		TrunkTransaction: trunkTxHash, BranchTransaction: branchTxHash,
		Command: Command{AttachToTangleCmd}, Trytes: trytes, MinWeightMagnitude: mwm,
	}
	rsp := &AttachToTangleResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return nil, err
	}

	return rsp.Trytes, nil
}

// BroadcastTransactions broadcasts a list of attached transaction trytes to the network.
// Tip-selection and Proof-of-Work must be done first by calling
// GetTransactionsToApprove and AttachToTangle or an equivalent attach method.
//
// You may use this method to increase odds of effective transaction propagation.
//
// Persist the transaction trytes in local storage before calling this command for first time, to ensure
// that reattachment is possible, until your bundle has been included.
func (api *API) BroadcastTransactions(trytes ...Trytes) ([]Trytes, error) {
	if err := Validate(ValidateAttachedTransactionTrytes(trytes...)); err != nil {
		return nil, err
	}
	cmd := &BroadcastTransactionsCommand{Trytes: trytes, Command: Command{BroadcastTransactionsCmd}}
	if err := api.provider.Send(cmd, nil); err != nil {
		return nil, err
	}
	return trytes, nil
}

// CheckConsistency checks if a transaction is consistent or a set of transactions are co-consistent.
//
// Co-consistent transactions and the transactions that they approve (directly or indirectly),
// are not conflicting with each other and the rest of the ledger.
//
// As long as a transaction is consistent, it might be accepted by the network.
// In case a transaction is inconsistent, it will not be accepted and a reattachment
// is required by calling ReplayBundle().
func (api *API) CheckConsistency(hashes ...Hash) (bool, string, error) {
	if err := Validate(
		ValidateTransactionHashes(hashes...),
		ValidateNonEmptyStrings(ErrInvalidTransactionHash, hashes...),
	); err != nil {
		return false, "", err
	}
	cmd := &CheckConsistencyCommand{Tails: hashes, Command: Command{CheckConsistencyCmd}}
	rsp := &CheckConsistencyResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return false, "", err
	}
	return rsp.State, rsp.Info, nil
}

func validateFindTransactions(query *FindTransactionsQuery) error {
	if err := Validate(
		ValidateAddresses(false, query.Addresses...),
		ValidateHashes(query.Bundles...),
		ValidateTransactionHashes(query.Approvees...),
		ValidateTags(query.Tags...),
	); err != nil {
		return err
	}
	if query.Addresses != nil {
		if err := Validate(ValidateNonEmptyStrings(ErrInvalidAddress, query.Addresses...)); err != nil {
			return err
		}
	}
	if query.Bundles != nil {
		if err := Validate(ValidateNonEmptyStrings(ErrInvalidBundleHash, query.Bundles...)); err != nil {
			return err
		}
	}
	if query.Approvees != nil {
		if err := Validate(ValidateNonEmptyStrings(ErrInvalidTransactionHash, query.Approvees...)); err != nil {
			return err
		}
	}
	if query.Tags != nil {
		if err := Validate(ValidateNonEmptyStrings(ErrInvalidTag, query.Tags...)); err != nil {
			return err
		}
	}
	return nil
}

// FindTransactions searches for transaction hashes.
// It allows to search for transactions by passing a query object with addresses, bundle hashes, tags and/or approvees fields.
// Multiple query fields are supported and FindTransactions returns the intersection of the results.
func (api *API) FindTransactions(query FindTransactionsQuery) (Hashes, error) {
	if err := validateFindTransactions(&query); err != nil {
		return nil, err
	}

	if len(query.Addresses) > 0 {
		cleanedAddrs, err := checksum.RemoveChecksums(query.Addresses)
		if err != nil {
			return nil, err
		}
		query.Addresses = cleanedAddrs
	}

	cmd := &FindTransactionsCommand{FindTransactionsQuery: query, Command: Command{FindTransactionsCmd}}
	rsp := &FindTransactionsResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return nil, err
	}
	return rsp.Hashes, nil
}

// GetBalances fetches confirmed balances of the given addresses at the latest solid milestone.
func (api *API) GetBalances(addresses Hashes, threshold uint64, tips ...Hash) (*Balances, error) {
    if err := Validate(ValidateAddresses(false, addresses...)); err != nil {
		return nil, err
	}

	if threshold > 100 {
		return nil, ErrInvalidThreshold
	}

	cleanedAddrs, err := checksum.RemoveChecksums(addresses)
	if err != nil {
		return nil, err
	}

	cmd := &GetBalancesCommand{
		Addresses: cleanedAddrs,
		Threshold: threshold,
		Command:   Command{GetBalancesCmd},
		Tips:      tips,
	}
	rsp := &GetBalancesResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return nil, err
	}
	balances := &Balances{
		Balances:  make([]uint64, len(rsp.Balances)),
		Milestone: rsp.Milestone, MilestoneIndex: rsp.MilestoneIndex,
	}
	for i, s := range rsp.Balances {
		num, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		balances.Balances[i] = num
	}
	return balances, err
}

// GetInclusionStates fetches inclusion states of a given list of transactions.
func (api *API) GetInclusionStates(txHashes Hashes, tips ...Hash) ([]bool, error) {
	if err := Validate(
		ValidateTransactionHashes(txHashes...),
		ValidateNonEmptyStrings(ErrInvalidTransactionHash, txHashes...),
		ValidateTransactionHashes(tips...),
		ValidateNonEmptyStrings(ErrInvalidTransactionHash, tips...),
	); err != nil {
		return nil, err
	}

	cmd := &GetInclusionStatesCommand{Transactions: txHashes, Tips: tips, Command: Command{GetInclusionStatesCmd}}
	rsp := &GetInclusionStatesResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return nil, err
	}
	return rsp.States, nil
}

// GetNeighbors returns the list of connected neighbors of the connected node.
func (api *API) GetNeighbors() (Neighbors, error) {
	cmd := &GetNeighborsCommand{Command: Command{GetNeighborsCmd}}
	rsp := &GetNeighborsResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return nil, err
	}
	return rsp.Neighbors, nil
}

// GetNodeInfo returns information about the connected node.
func (api *API) GetNodeInfo() (*GetNodeInfoResponse, error) {
	cmd := &GetNodeInfoCommand{Command: Command{GetNodeInfoCmd}}
	rsp := &GetNodeInfoResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return nil, err
	}
	return rsp, nil
}

// GetTips returns a list of tips (transactions not referenced by other transactions) as seen by the connected node.
func (api *API) GetTips() (Hashes, error) {
	cmd := &GetTipsCommand{Command: Command{GetTipsCmd}}
	rsp := &GetTipsResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return nil, err
	}
	return rsp.Hashes, nil
}

// GetTransactionsToApprove does the tip selection via the connected node.
//
// Returns a pair of approved transactions which are chosen randomly after validating the transaction trytes,
// the signatures and cross-checking for conflicting transactions.
//
// Tip selection is executed by a Random Walk (RW) starting at random point in the given depth,
// ending up to the pair of selected tips. For more information about tip selection please refer to the
// whitepaper (http://iotatoken.com/IOTA_Whitepaper.pdf).
//
// The reference option allows to select tips in a way that the reference transaction is being approved too.
// This is useful for promoting transactions, for example with PromoteTransaction().
func (api *API) GetTransactionsToApprove(depth uint64, reference ...Hash) (*TransactionsToApprove, error) {
	cmd := &GetTransactionsToApproveCommand{Command: Command{GetTransactionsToApproveCmd}, Depth: depth}
	if len(reference) > 0 {
		if err := Validate(ValidateTransactionHashes(reference...)); err != nil {
			return nil, ErrInvalidReferenceHash
		}
		cmd.Reference = reference[0]
	}
	rsp := &GetTransactionsToApproveResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return nil, err
	}
	return &rsp.TransactionsToApprove, nil
}

// GetTrytes fetches the transaction trytes given a list of transaction hashes.
func (api *API) GetTrytes(hashes ...Hash) ([]Trytes, error) {
	if err := Validate(
		ValidateNonEmptyStrings(ErrInvalidTransactionHash, hashes...),
		ValidateTransactionHashes(hashes...),
	); err != nil {
		return nil, err
	}
	cmd := &GetTrytesCommand{Hashes: hashes, Command: Command{GetTrytesCmd}}
	rsp := &GetTrytesResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return nil, err
	}
	return rsp.Trytes, nil
}

// InterruptAttachToTangle interrupts the currently ongoing Proof-of-Work on the connected node.
func (api *API) InterruptAttachToTangle() error {
	cmd := &InterruptAttachToTangleCommand{Command: Command{InterruptAttachToTangleCmd}}
	return api.provider.Send(cmd, nil)
}

// RemoveNeighbors removes a list of neighbors from the connected IRI node.
// This method has a temporary effect until the IRI node relaunches.
func (api *API) RemoveNeighbors(uris ...string) (int64, error) {
	if err := Validate(
		ValidateNonEmptyStrings(ErrInvalidURI, uris...),
		ValidateURIs(uris...),
	); err != nil {
		return 0, err
	}
	cmd := &RemoveNeighborsCommand{Command: Command{RemoveNeighborsCmd}, URIs: uris}
	rsp := &RemoveNeighborsResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return 0, err
	}
	return rsp.RemovedNeighbors, nil
}

// StoreTransactions persists a list of attached transaction trytes in the store of the connected node.
// Tip-selection and Proof-of-Work must be done first by calling GetTransactionsToApprove and
// AttachToTangle or an equivalent attach method.
//
// Persist the transaction trytes in local storage before calling this command, to ensure
// reattachment is possible, until your bundle has been included.
//
// Any transactions stored with this command will eventually be erased as a result of a snapshot.
func (api *API) StoreTransactions(trytes ...Trytes) ([]Trytes, error) {
	if err := Validate(
		ValidateNonEmptyStrings(ErrInvalidTrytes, trytes...),
		ValidateAttachedTransactionTrytes(trytes...),
	); err != nil {
		return nil, err
	}
	cmd := &StoreTransactionsCommand{Trytes: trytes, Command: Command{StoreTransactionsCmd}}
	if err := api.provider.Send(cmd, nil); err != nil {
		return nil, err
	}
	return trytes, nil
}

// WereAddressesSpentFrom checks whether the given addresses were already spent.
func (api *API) WereAddressesSpentFrom(addresses ...Hash) ([]bool, error) {
	if err := Validate(
		ValidateAddresses(false, addresses...),
	); err != nil {
		return nil, err
	}

	cleanedAddrs, err := checksum.RemoveChecksums(addresses)
	if err != nil {
		return nil, err
	}

	cmd := &WereAddressesSpentFromCommand{Addresses: cleanedAddrs, Command: Command{WereAddressesSpentFromCmd}}
	rsp := &WereAddressesSpentFromResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return nil, err
	}
	return rsp.States, nil
}
