package api

import (
	. "github.com/iotaledger/giota/trinary"
	"strconv"
)

func createAddNeighbors(provider Provider) func(...string) (int64, error) {
	return func(uris ...string) (int64, error) {
		cmd := &AddNeighborsCommand{URIs: uris, Command: AddNeighborsCmd}
		rsp := &AddNeighborsResponse{}
		err := provider.Send(cmd, rsp)
		return rsp.AddedNeighbors, err
	}
}

func createAttachToTangle(provider Provider) func(trunkTxHash Hash, branchTxHash Hash, mwm uint64, trytes []Trytes) ([]Trytes, error) {
	return func(trunkTxHash Hash, branchTxHash Hash, mwm uint64, trytes []Trytes) ([]Trytes, error) {
		// TODO: validate mwm, trytes, trunk/branch hashes
		cmd := &AttachToTangleCommand{
			TrunkTransaction: trunkTxHash, BranchTransaction: branchTxHash,
			Command: AttachToTangleCmd, Trytes: trytes, MinWeightMagnitude: mwm,
		}
		rsp := &AttachToTangleResponse{}
		err := provider.Send(cmd, rsp)
		if err != nil {
			return nil, err
		}
		return rsp.Trytes, nil
	}
}

func createBroadcastTransactions(provider Provider) func(trytes ...Trytes) ([]Trytes, error) {
	return func(trytes ...Trytes) ([]Trytes, error) {
		cmd := &BroadcastTransactionsCommand{Trytes: trytes, Command: BroadcastTransactionsCmd}
		err := provider.Send(cmd, nil)
		if err != nil {
			return nil, err
		}
		return trytes, err
	}
}

func createCheckConsistency(provider Provider) func(hashes ...Hash) (bool, error) {
	return func(hashes ...Hash) (bool, error) {
		cmd := &CheckConsistencyCommand{Tails: hashes, Command: CheckConsistencyCmd}
		rsp := &CheckConsistencyResponse{}
		err := provider.Send(cmd, rsp)
		if err != nil {
			return false, err
		}
		return rsp.State, nil
	}
}

func createFindTransactions(provider Provider) func(query FindTransactionsQuery) (Hashes, error) {
	return func(query FindTransactionsQuery) (Hashes, error) {
		// TODO: strip away checksums in query object
		cmd := &FindTransactionsCommand{FindTransactionsQuery: query, Command: FindTransactionsCmd}
		rsp := &FindTransactionsResponse{}
		err := provider.Send(cmd, rsp)
		if err != nil {
			return nil, err
		}
		return rsp.Hashes, nil
	}
}

func createGetBalances(provider Provider) func(addresses Hashes, threshold uint64) (*Balances, error) {
	return func(addresses Hashes, threshold uint64) (*Balances, error) {
		// TODO: validate hashes and threshold
		cmd := &GetBalancesCommand{Addresses: addresses, Threshold: threshold, Command: GetBalancesCmd}
		rsp := &GetBalancesResponse{}
		err := provider.Send(cmd, rsp)
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
}

func createInclusionState(provider Provider) func(txHash Hashes, tips ...Hash) ([]bool, error) {
	return func(txHash Hashes, tips ...Hash) ([]bool, error) {
		// TODO: validate tx and tip hashes
		cmd := &GetInclusionStateCommand{Transactions: txHash, Tips: tips, Command: GetInclusionStatesCmd}
		rsp := &GetInclusionStatesResponse{}
		err := provider.Send(cmd, rsp)
		if err != nil {
			return nil, err
		}
		return rsp.States, nil
	}
}

func createGetNeighbors(provider Provider) func() (Neighbors, error) {
	return func() (Neighbors, error) {
		cmd := &GetNeighborsCommand{Command: GetNeighborsCmd}
		rsp := &GetNeighborsResponse{}
		err := provider.Send(cmd, rsp)
		if err != nil {
			return nil, err
		}
		return rsp.Neighbors, nil
	}
}

func createGetNodeInfo(provider Provider) func() (*GetNodeInfoResponse, error) {
	return func() (*GetNodeInfoResponse, error) {
		cmd := &GetNodeInfoCommand{Command: GetNodeInfoCmd}
		rsp := &GetNodeInfoResponse{}
		err := provider.Send(cmd, rsp)
		if err != nil {
			return nil, err
		}
		return rsp, nil
	}
}

func createGetTips(provider Provider) func() (Hashes, error) {
	return func() (Hashes, error) {
		cmd := &GetTipsCommand{Command: GetTipsCmd}
		rsp := &GetTipsResponse{}
		err := provider.Send(cmd, rsp)
		if err != nil {
			return nil, err
		}
		return rsp.Hashes, nil
	}
}

func createGetTransactionsToApprove(provider Provider) func(depth uint64, reference ...Hash) (*TransactionsToApprove, error) {
	return func(depth uint64, reference ...Hash) (*TransactionsToApprove, error) {
		// TODO: validate depth and reference
		cmd := &GetTransactionsToApproveCommand{Command: GetTransactionsToApproveCmd, Depth: depth}
		if len(reference) > 0 {
			cmd.Reference = reference[0]
		}
		rsp := &GetTransactionsToApproveResponse{}
		err := provider.Send(cmd, rsp)
		if err != nil {
			return nil, err
		}
		return &rsp.TransactionsToApprove, nil
	}
}

func createGetTrytes(provider Provider) func(...Hash) ([]Trytes, error) {
	return func(hashes ...Hash) ([]Trytes, error) {
		// TODO: validate transaction hashes
		cmd := &GetTrytesCommand{Hashes: hashes, Command: GetTrytesCmd}
		rsp := &GetTrytesResponse{}
		err := provider.Send(cmd, rsp)
		if err != nil {
			return nil, err
		}
		return rsp.Trytes, nil
	}
}

func createInterruptAttachToTangle(provider Provider) func() error {
	return func() error {
		cmd := &InterruptAttachToTangleCommand{Command: InterruptAttachToTangleCmd}
		return provider.Send(cmd, nil)
	}
}

func createRemoveNeighbors(provider Provider) func(uris ...string) (int64, error) {
	return func(uris ...string) (int64, error) {
		// TODO: validate uris
		cmd := &RemoveNeighborsCommand{Command: RemoveNeighborsCmd}
		rsp := &RemoveNeighborsResponse{}
		err := provider.Send(cmd, rsp)
		if err != nil {
			return 0, err
		}
		return rsp.RemovedNeighbors, nil
	}
}

func createStoreTransactions(provider Provider) func(trytes ...Trytes) ([]Trytes, error) {
	return func(trytes ...Trytes) ([]Trytes, error) {
		// TODO: validate _attached_ trytes
		cmd := &StoreTransactionsCommand{Trytes: trytes, Command: StoreTransactionsCmd}
		err := provider.Send(cmd, nil)
		if err != nil {
			return nil, err
		}
		return trytes, nil
	}
}

func createWereAddressesSpentFrom(provider Provider) func(addresses ...Hash) ([]bool, error) {
	return func(addresses ...Hash) ([]bool, error) {
		cmd := &WereAddressesSpentFromCommand{Addresses: addresses, Command: WereAddressesSpentFromCmd}
		rsp := &WereAddressesSpentFromResponse{}
		err := provider.Send(cmd, rsp)
		if err != nil {
			return nil, err
		}
		return rsp.States, nil
	}
}
