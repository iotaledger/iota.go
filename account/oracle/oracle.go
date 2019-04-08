package oracle

import "github.com/iotaledger/iota.go/account/deposit"

// SendOracle tells whether it makes sense to send a transaction or not
// by using its sources to make a decision.
type SendOracle interface {
	// OkToSend tells whether it makes sense to send a transaction
	// to the given conditional deposit address.
	OkToSend(conds *deposit.CDA) (bool, string, error)
}

// OracleSource defines an input for the SendOracle.
type OracleSource interface {
	Ok(conds *deposit.CDA) (bool, string, error)
}

// New creates a new SendOracle which uses the provided sources to make its decision.
func New(sources ...OracleSource) SendOracle {
	return &sendoracle{sources: sources}
}

type sendoracle struct {
	sources []OracleSource
}

func (so *sendoracle) OkToSend(conds *deposit.CDA) (bool, string, error) {
	for _, src := range so.sources {
		ok, info, err := src.Ok(conds)
		if err != nil {
			return false, "", err
		}
		if !ok {
			return false, info, nil
		}
	}
	return true, "", nil
}
