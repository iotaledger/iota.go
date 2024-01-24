package tpkg

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

// RandBlock returns a random block with the given inner payload.
func RandBlock(blockBody iotago.BlockBody, api iotago.API, rmc iotago.Mana) *iotago.Block {
	block := &iotago.Block{
		API: api,
		Header: iotago.BlockHeader{
			ProtocolVersion:  ZeroCostTestAPI.Version(),
			NetworkID:        api.ProtocolParameters().NetworkID(),
			IssuingTime:      RandUTCTime(),
			SlotCommitmentID: iotago.NewEmptyCommitment(api).MustID(),
			IssuerID:         RandAccountID(),
		},
		Body:      blockBody,
		Signature: RandEd25519Signature(),
	}

	if basicBlock, isBasic := blockBody.(*iotago.BasicBlockBody); isBasic {
		burnedMana, err := block.ManaCost(rmc)
		if err != nil {
			panic(err)
		}
		basicBlock.MaxBurnedMana = burnedMana
	}

	return block
}

func RandBasicBlockWithIssuerAndRMC(api iotago.API, issuerID iotago.AccountID, rmc iotago.Mana) *iotago.Block {
	basicBlock := RandBasicBlockBody(api, iotago.PayloadSignedTransaction)

	block := RandBlock(basicBlock, ZeroCostTestAPI, rmc)
	block.Header.IssuerID = issuerID

	return block
}

func RandBasicBlockBodyWithPayload(api iotago.API, payload iotago.ApplicationPayload) *iotago.BasicBlockBody {
	return &iotago.BasicBlockBody{
		API:                api,
		StrongParents:      SortedRandBlockIDs(1 + RandInt(iotago.BasicBlockMaxParents)),
		WeakParents:        iotago.BlockIDs{},
		ShallowLikeParents: iotago.BlockIDs{},
		Payload:            payload,
		MaxBurnedMana:      RandMana(1000),
	}
}

func RandBasicBlockBody(api iotago.API, withPayloadType iotago.PayloadType) *iotago.BasicBlockBody {
	var payload iotago.ApplicationPayload

	//nolint:exhaustive
	switch withPayloadType {
	case iotago.PayloadSignedTransaction:
		payload = RandSignedTransaction(api)
	case iotago.PayloadTaggedData:
		payload = RandTaggedData([]byte("tag"))
	case iotago.PayloadCandidacyAnnouncement:
		payload = &iotago.CandidacyAnnouncement{}
	}

	return RandBasicBlockBodyWithPayload(api, payload)
}

func RandValidationBlockBody(api iotago.API) *iotago.ValidationBlockBody {
	return &iotago.ValidationBlockBody{
		API:                     api,
		StrongParents:           SortedRandBlockIDs(1 + RandInt(iotago.ValidationBlockMaxParents)),
		WeakParents:             iotago.BlockIDs{},
		ShallowLikeParents:      iotago.BlockIDs{},
		HighestSupportedVersion: ZeroCostTestAPI.Version() + 1,
	}
}
