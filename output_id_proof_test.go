package iotago_test

import (
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

type outputIDProofTest struct {
	name string
	tx   *iotago.Transaction
}

func TestOutputIDProof(t *testing.T) {
	addr1 := tpkg.RandEd25519Address()

	inputIDs := tpkg.RandOutputIDs(1)

	tests := []outputIDProofTest{
		{
			name: "single output",
			tx: &iotago.Transaction{
				API: tpkg.ZeroCostTestAPI,
				TransactionEssence: &iotago.TransactionEssence{
					CreationSlot: tpkg.RandSlot(),
					NetworkID:    tpkg.TestNetworkID,
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMask{},
				},
				Outputs: lo.RepeatBy(1, func(_ int) iotago.TxEssenceOutput {
					return &iotago.BasicOutput{
						Amount: OneIOTA,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: addr1},
						},
					}
				}),
			},
		},
		{
			name: "two outputs",
			tx: &iotago.Transaction{
				API: tpkg.ZeroCostTestAPI,
				TransactionEssence: &iotago.TransactionEssence{
					CreationSlot: tpkg.RandSlot(),
					NetworkID:    tpkg.TestNetworkID,
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMask{},
				},
				Outputs: lo.RepeatBy(2, func(_ int) iotago.TxEssenceOutput {
					return &iotago.BasicOutput{
						Amount: OneIOTA,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: addr1},
						},
					}
				}),
			},
		},
		{
			name: "three outputs",
			tx: &iotago.Transaction{
				API: tpkg.ZeroCostTestAPI,
				TransactionEssence: &iotago.TransactionEssence{
					CreationSlot: tpkg.RandSlot(),
					NetworkID:    tpkg.TestNetworkID,
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMask{},
				},
				Outputs: lo.RepeatBy(3, func(_ int) iotago.TxEssenceOutput {
					return &iotago.BasicOutput{
						Amount: OneIOTA,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: addr1},
						},
					}
				}),
			},
		},
		{
			name: "max outputs",
			tx: &iotago.Transaction{
				API: tpkg.ZeroCostTestAPI,
				TransactionEssence: &iotago.TransactionEssence{
					CreationSlot: tpkg.RandSlot(),
					NetworkID:    tpkg.TestNetworkID,
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMask{},
				},
				Outputs: lo.RepeatBy(iotago.MaxOutputsCount, func(_ int) iotago.TxEssenceOutput {
					return &iotago.BasicOutput{
						Amount: OneIOTA,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: addr1},
						},
					}
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testOutputs)
	}
}

func (p *outputIDProofTest) testOutputs(t *testing.T) {
	outputSet, err := p.tx.OutputsSet()
	require.NoError(t, err)

	for outputID, output := range outputSet {
		proof, err := iotago.OutputIDProofFromTransaction(p.tx, outputID.Index())
		require.NoError(t, err)

		serializedProof, err := proof.Bytes()
		require.NoError(t, err)

		jsonEncoded, err := tpkg.ZeroCostTestAPI.JSONEncode(proof)
		require.NoError(t, err)
		fmt.Println(string(jsonEncoded))

		deserializedProof, consumedBytes, err := iotago.OutputIDProofFromBytes(tpkg.ZeroCostTestAPI)(serializedProof)
		require.NoError(t, err)
		require.Equal(t, len(serializedProof), consumedBytes)

		require.Equal(t, proof, deserializedProof)

		computedOutputID, err := deserializedProof.OutputID(output)
		require.NoError(t, err)

		require.Equal(t, outputID, computedOutputID)
	}
}
