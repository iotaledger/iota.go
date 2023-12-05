package tpkg

import (
	cryptorand "crypto/rand"
	"encoding/binary"

	"github.com/iotaledger/hive.go/runtime/options"
	iotago "github.com/iotaledger/iota.go/v4"
)

func RandSignedTransactionIDWithCreationSlot(slot iotago.SlotIndex) iotago.SignedTransactionID {
	var signedTransactionID iotago.SignedTransactionID
	_, err := cryptorand.Read(signedTransactionID[:iotago.IdentifierLength])
	if err != nil {
		panic(err)
	}
	binary.LittleEndian.PutUint32(signedTransactionID[iotago.IdentifierLength:iotago.TransactionIDLength], uint32(slot))

	return signedTransactionID
}

func RandTransactionIDWithCreationSlot(slot iotago.SlotIndex) iotago.TransactionID {
	var transactionID iotago.TransactionID
	_, err := cryptorand.Read(transactionID[:iotago.IdentifierLength])
	if err != nil {
		panic(err)
	}
	binary.LittleEndian.PutUint32(transactionID[iotago.IdentifierLength:iotago.TransactionIDLength], uint32(slot))

	return transactionID
}

func RandSignedTransactionID() iotago.SignedTransactionID {
	return RandSignedTransactionIDWithCreationSlot(RandSlot())
}

func RandTransactionID() iotago.TransactionID {
	return RandTransactionIDWithCreationSlot(RandSlot())
}

// RandTransaction returns a random transaction essence.
func RandTransaction(api iotago.API, opts ...options.Option[iotago.Transaction]) *iotago.Transaction {
	return RandTransactionWithOptions(
		api,
		append([]options.Option[iotago.Transaction]{
			WithUTXOInputCount(RandInt(iotago.MaxInputsCount) + 1),
			WithOutputCount(RandInt(iotago.MaxOutputsCount) + 1),
			WithAllotmentCount(RandInt(iotago.MaxAllotmentCount) + 1),
		}, opts...)...,
	)
}

// RandTransactionWithInputCount returns a random transaction essence with a specific amount of inputs..
func RandTransactionWithInputCount(api iotago.API, inputCount int) *iotago.Transaction {
	return RandTransactionWithOptions(
		api,
		WithUTXOInputCount(inputCount),
		WithOutputCount(RandInt(iotago.MaxOutputsCount)+1),
		WithAllotmentCount(RandInt(iotago.MaxAllotmentCount)+1),
	)
}

// RandTransactionWithOutputCount returns a random transaction essence with a specific amount of outputs.
func RandTransactionWithOutputCount(api iotago.API, outputCount int) *iotago.Transaction {
	return RandTransactionWithOptions(
		api,
		WithUTXOInputCount(RandInt(iotago.MaxInputsCount)+1),
		WithOutputCount(outputCount),
		WithAllotmentCount(RandInt(iotago.MaxAllotmentCount)+1),
	)
}

// RandTransactionWithAllotmentCount returns a random transaction essence with a specific amount of outputs.
func RandTransactionWithAllotmentCount(api iotago.API, allotmentCount int) *iotago.Transaction {
	return RandTransactionWithOptions(
		api,
		WithUTXOInputCount(RandInt(iotago.MaxInputsCount)+1),
		WithOutputCount(RandInt(iotago.MaxOutputsCount)+1),
		WithAllotmentCount(allotmentCount),
	)
}

// RandTransactionWithOptions returns a random transaction essence with options applied.
func RandTransactionWithOptions(api iotago.API, opts ...options.Option[iotago.Transaction]) *iotago.Transaction {
	tx := &iotago.Transaction{
		API: api,
		TransactionEssence: &iotago.TransactionEssence{
			NetworkID:     TestNetworkID,
			ContextInputs: iotago.TxEssenceContextInputs{},
			Inputs:        iotago.TxEssenceInputs{},
			Allotments:    iotago.Allotments{},
			Capabilities:  iotago.TransactionCapabilitiesBitMask{},
		},
		Outputs: iotago.TxEssenceOutputs{},
	}

	inputCount := 1
	for i := inputCount; i > 0; i-- {
		tx.TransactionEssence.Inputs = append(tx.TransactionEssence.Inputs, RandUTXOInput())
	}

	outputCount := 1
	for i := outputCount; i > 0; i-- {
		tx.Outputs = append(tx.Outputs, RandBasicOutput(iotago.AddressEd25519))
	}

	return options.Apply(tx, opts)
}

func WithBlockIssuanceCreditInputCount(inputCount int) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		for i := inputCount; i > 0; i-- {
			tx.TransactionEssence.ContextInputs = append(tx.TransactionEssence.ContextInputs, RandBlockIssuanceCreditInput())
		}
	}
}

func WithRewardInputCount(inputCount uint16) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		for i := inputCount; i > 0; i-- {
			rewardInput := &iotago.RewardInput{
				Index: i,
			}
			tx.TransactionEssence.ContextInputs = append(tx.TransactionEssence.ContextInputs, rewardInput)
		}
	}
}

func WithCommitmentInput() options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.TransactionEssence.ContextInputs = append(tx.TransactionEssence.ContextInputs, RandCommitmentInput())
	}
}

func WithUTXOInputCount(inputCount int) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.TransactionEssence.Inputs = make(iotago.TxEssenceInputs, 0, inputCount)

		for i := inputCount; i > 0; i-- {
			tx.TransactionEssence.Inputs = append(tx.TransactionEssence.Inputs, RandUTXOInput())
		}
	}
}

func WithOutputCount(outputCount int) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.Outputs = make(iotago.TxEssenceOutputs, 0, outputCount)

		for i := outputCount; i > 0; i-- {
			tx.Outputs = append(tx.Outputs, RandBasicOutput(iotago.AddressEd25519))
		}
	}
}

func WithOutputs(outputs iotago.TxEssenceOutputs) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.Outputs = outputs
	}
}

func WithAllotmentCount(allotmentCount int) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.Allotments = RandSortAllotment(allotmentCount)
	}
}

func WithInputs(inputs iotago.TxEssenceInputs) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.TransactionEssence.Inputs = inputs
	}
}

func WithContextInputs(inputs iotago.TxEssenceContextInputs) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.TransactionEssence.ContextInputs = inputs
	}
}

func WithAllotments(allotments iotago.TxEssenceAllotments) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.Allotments = allotments
	}
}

func WithTxEssencePayload(payload iotago.TxEssencePayload) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.Payload = payload
	}
}

// RandSignedTransactionWithTransaction returns a random transaction with a specific essence.
func RandSignedTransactionWithTransaction(api iotago.API, transaction *iotago.Transaction) *iotago.SignedTransaction {
	sigTxPayload := &iotago.SignedTransaction{API: api}
	sigTxPayload.Transaction = transaction

	unlocksCount := len(transaction.TransactionEssence.Inputs)
	for i := unlocksCount; i > 0; i-- {
		sigTxPayload.Unlocks = append(sigTxPayload.Unlocks, RandEd25519SignatureUnlock())
	}

	return sigTxPayload
}

// RandSignedTransaction returns a random transaction.
func RandSignedTransaction(api iotago.API, opts ...options.Option[iotago.Transaction]) *iotago.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransaction(api, opts...))
}

// RandSignedTransactionWithUTXOInputCount returns a random transaction with a specific amount of inputs.
func RandSignedTransactionWithUTXOInputCount(api iotago.API, inputCount int) *iotago.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithInputCount(api, inputCount))
}

// RandSignedTransactionWithOutputCount returns a random transaction with a specific amount of outputs.
func RandSignedTransactionWithOutputCount(api iotago.API, outputCount int) *iotago.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithOutputCount(api, outputCount))
}

// RandSignedTransactionWithAllotmentCount returns a random transaction with a specific amount of allotments.
func RandSignedTransactionWithAllotmentCount(api iotago.API, allotmentCount int) *iotago.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithAllotmentCount(api, allotmentCount))
}

// RandSignedTransactionWithInputOutputCount returns a random transaction with a specific amount of inputs and outputs.
func RandSignedTransactionWithInputOutputCount(api iotago.API, inputCount int, outputCount int) *iotago.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithOptions(api, WithUTXOInputCount(inputCount), WithOutputCount(outputCount)))
}

// OneInputOutputTransaction generates a random transaction with one input and output.
func OneInputOutputTransaction() *iotago.SignedTransaction {
	return &iotago.SignedTransaction{
		API: TestAPI,
		Transaction: &iotago.Transaction{
			API: TestAPI,
			TransactionEssence: &iotago.TransactionEssence{
				NetworkID:     14147312347886322761,
				ContextInputs: iotago.TxEssenceContextInputs{},
				Inputs: iotago.TxEssenceInputs{
					&iotago.UTXOInput{
						TransactionID: func() iotago.TransactionID {
							var b iotago.TransactionID
							copy(b[:], RandBytes(iotago.TransactionIDLength))

							return b
						}(),
						TransactionOutputIndex: 0,
					},
				},
				Allotments:   iotago.Allotments{},
				Capabilities: iotago.TransactionCapabilitiesBitMask{},
				Payload:      nil,
			},
			Outputs: iotago.TxEssenceOutputs{
				&iotago.BasicOutput{
					Amount: 1337,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: RandEd25519Address()},
					},
				},
			},
		},
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{
				Signature: RandEd25519Signature(),
			},
		},
	}
}
