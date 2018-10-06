package api

import (
	. "github.com/iotaledger/iota.go/trinary"
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
	// TODO: validate mwm, trytes, trunk/branch hashes
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
	cmd := &BroadcastTransactionsCommand{Trytes: trytes, Command: BroadcastTransactionsCmd}
	err := api.provider.Send(cmd, nil)
	if err != nil {
		return nil, err
	}
	return trytes, err
}

func (api *API) CheckConsistency(hashes ...Hash) (bool, error) {
	cmd := &CheckConsistencyCommand{Tails: hashes, Command: CheckConsistencyCmd}
	rsp := &CheckConsistencyResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return false, err
	}
	return rsp.State, nil

}

func (api *API) FindTransactions(query FindTransactionsQuery) (Hashes, error) {
	// TODO: strip away checksums in query object
	cmd := &FindTransactionsCommand{FindTransactionsQuery: query, Command: FindTransactionsCmd}
	rsp := &FindTransactionsResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return nil, err
	}
	return rsp.Hashes, nil
}

func (api *API) GetBalances(addresses Hashes, threshold uint64) (*Balances, error) {
	// TODO: validate hashes and threshold
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
	// TODO: validate tx and tip hashes
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
	// TODO: validate depth and reference
	cmd := &GetTransactionsToApproveCommand{Command: GetTransactionsToApproveCmd, Depth: depth}
	if len(reference) > 0 {
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
	// TODO: validate transaction hashes
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
	// TODO: validate uris
	cmd := &RemoveNeighborsCommand{Command: RemoveNeighborsCmd}
	rsp := &RemoveNeighborsResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return 0, err
	}
	return rsp.RemovedNeighbors, nil
}

func (api *API) StoreTransactions(trytes ...Trytes) ([]Trytes, error) {
	// TODO: validate _attached_ trytes
	cmd := &StoreTransactionsCommand{Trytes: trytes, Command: StoreTransactionsCmd}
	err := api.provider.Send(cmd, nil)
	if err != nil {
		return nil, err
	}
	return trytes, nil
}

func (api *API) WereAddressesSpentFrom(addresses ...Hash) ([]bool, error) {
	cmd := &WereAddressesSpentFromCommand{Addresses: addresses, Command: WereAddressesSpentFromCmd}
	rsp := &WereAddressesSpentFromResponse{}
	err := api.provider.Send(cmd, rsp)
	if err != nil {
		return nil, err
	}
	return rsp.States, nil
}
