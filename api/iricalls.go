package api

import (
	"github.com/iotaledger/iota.go/checksum"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/iotaledger/iota.go/utils"
	. "github.com/iotaledger/iota.go/utils"
	"strconv"
)

func (api *API) AddNeighbors(uris ...string) (int64, error) {
	cmd := &AddNeighborsCommand{URIs: uris, Command: AddNeighborsCmd}
	rsp := &AddNeighborsResponse{}
	err := api.provider.Send(cmd, rsp)
	return rsp.AddedNeighbors, err
}

func (api *API) AttachToTangle(trunkTxHash Hash, branchTxHash Hash, mwm uint64, trytes []Trytes) ([]Trytes, error) {
	if api.attachToTangle != nil {
		return api.attachToTangle(trunkTxHash, branchTxHash, mwm, trytes)
	}

	if err := Validate(ValidateTransactionTrytes(trytes...)); err != nil {
		return nil, err
	}

	if !utils.IsTransactionHash(trunkTxHash) {
		return nil, ErrInvalidTrunkTransaction
	}

	if !utils.IsTransactionHash(branchTxHash) {
		return nil, ErrInvalidBranchTransaction
	}

	cmd := &AttachToTangleCommand{
		TrunkTransaction: trunkTxHash, BranchTransaction: branchTxHash,
		Command: AttachToTangleCmd, Trytes: trytes, MinWeightMagnitude: mwm,
	}
	rsp := &AttachToTangleResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return nil, err
	}
	return rsp.Trytes, nil
}

func (api *API) BroadcastTransactions(trytes ...Trytes) ([]Trytes, error) {
	if err := Validate(ValidateAttachedTransactionTrytes(trytes...)); err != nil {
		return nil, err
	}
	cmd := &BroadcastTransactionsCommand{Trytes: trytes, Command: BroadcastTransactionsCmd}
	err := api.provider.Send(cmd, nil)
	if err != nil {
		return nil, err
	}
	return trytes, err
}

func (api *API) CheckConsistency(hashes ...Hash) (bool, error) {
	if err := Validate(ValidateTransactionHashes(hashes...)); err != nil {
		return false, err
	}
	cmd := &CheckConsistencyCommand{Tails: hashes, Command: CheckConsistencyCmd}
	rsp := &CheckConsistencyResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return false, err
	}
	return rsp.State, nil
}

func validateFindTransactions(query *FindTransactionsQuery) error {
	return Validate(
		ValidateHashes(query.Addresses...),
		ValidateHashes(query.Bundles...),
		ValidateTransactionHashes(query.Approvees...),
		ValidateTags(query.Tags...),
	)
}

func (api *API) FindTransactions(query FindTransactionsQuery) (Hashes, error) {
	if err := validateFindTransactions(&query); err != nil {
		return nil, err
	}

	cleanedAddrs, err := checksum.RemoveChecksums(query.Addresses)
	if err != nil {
		return nil, err
	}
	query.Addresses = cleanedAddrs

	cmd := &FindTransactionsCommand{FindTransactionsQuery: query, Command: FindTransactionsCmd}
	rsp := &FindTransactionsResponse{}
	if err := api.provider.Send(cmd, rsp); err != nil {
		return nil, err
	}
	return rsp.Hashes, nil
}

func (api *API) GetBalances(addresses Hashes, threshold uint64) (*Balances, error) {
	if err := Validate(ValidateHashes(addresses...)); err != nil {
		return nil, err
	}

	if threshold > 100 {
		return nil, ErrInvalidThreshold
	}

	cmd := &GetBalancesCommand{Addresses: addresses, Threshold: threshold, Command: GetBalancesCmd}
	rsp := &GetBalancesResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
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

func (api *API) GetInclusionStates(txHash Hashes, tips ...Hash) ([]bool, error) {
	if err := Validate(
		ValidateTransactionHashes(txHash...),
		ValidateTransactionHashes(tips...)); err != nil {
		return nil, err
	}

	cmd := &GetInclusionStateCommand{Transactions: txHash, Tips: tips, Command: GetInclusionStatesCmd}
	rsp := &GetInclusionStatesResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return nil, err
	}
	return rsp.States, nil
}

func (api *API) GetNeighbors() (Neighbors, error) {
	cmd := &GetNeighborsCommand{Command: GetNeighborsCmd}
	rsp := &GetNeighborsResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return nil, err
	}
	return rsp.Neighbors, nil
}

func (api *API) GetNodeInfo() (*GetNodeInfoResponse, error) {
	cmd := &GetNodeInfoCommand{Command: GetNodeInfoCmd}
	rsp := &GetNodeInfoResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (api *API) GetTips() (Hashes, error) {
	cmd := &GetTipsCommand{Command: GetTipsCmd}
	rsp := &GetTipsResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return nil, err
	}
	return rsp.Hashes, nil
}

func (api *API) GetTransactionsToApprove(depth uint64, reference ...Hash) (*TransactionsToApprove, error) {
	cmd := &GetTransactionsToApproveCommand{Command: GetTransactionsToApproveCmd, Depth: depth}
	if len(reference) > 0 {
		if !IsTransactionHash(reference[0]) {
			return nil, ErrInvalidReferenceHash
		}
		cmd.Reference = reference[0]
	}
	rsp := &GetTransactionsToApproveResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return nil, err
	}
	return &rsp.TransactionsToApprove, nil
}

func (api *API) GetTrytes(hashes ...Hash) ([]Trytes, error) {
	if err := Validate(ValidateTransactionHashes(hashes...)); err != nil {
		return nil, err
	}
	cmd := &GetTrytesCommand{Hashes: hashes, Command: GetTrytesCmd}
	rsp := &GetTrytesResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return nil, err
	}
	return rsp.Trytes, nil
}

func (api *API) InterruptAttachToTangle() error {
	cmd := &InterruptAttachToTangleCommand{Command: InterruptAttachToTangleCmd}
	return api.provider.Send(cmd, nil)
}

func (api *API) RemoveNeighbors(uris ...string) (int64, error) {
	if err := Validate(ValidateTags(uris...)); err != nil {
		return 0, err
	}
	cmd := &RemoveNeighborsCommand{Command: RemoveNeighborsCmd}
	rsp := &RemoveNeighborsResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return 0, err
	}
	return rsp.RemovedNeighbors, nil
}

func (api *API) StoreTransactions(trytes ...Trytes) ([]Trytes, error) {
	if err := Validate(ValidateAttachedTransactionTrytes(trytes...)); err != nil {
		return nil, err
	}
	cmd := &StoreTransactionsCommand{Trytes: trytes, Command: StoreTransactionsCmd}
	err := api.provider.Send(cmd, nil)
	if err != nil {
		return nil, err
	}
	return trytes, nil
}

func (api *API) WereAddressesSpentFrom(addresses ...Hash) ([]bool, error) {
	if err := Validate(ValidateHashes(addresses...)); err != nil {
		return nil, err
	}
	cmd := &WereAddressesSpentFromCommand{Addresses: addresses, Command: WereAddressesSpentFromCmd}
	rsp := &WereAddressesSpentFromResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return nil, err
	}
	return rsp.States, nil
}
