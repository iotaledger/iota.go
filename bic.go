package iotago

type BlockIssuanceCredit struct {
	AccountID
	CommitmentID
	Value int64
}

type BICInputSet map[AccountID]BlockIssuanceCredit

func (b BlockIssuanceCredit) Negative() bool {
	return b.Value < 0
}
