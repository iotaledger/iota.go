package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestAllotmentDeSerialize(t *testing.T) {
	type allotmentDeSerializeTest struct {
		name      string
		source    iotago.TxEssenceAllotments
		seriErr   error
		deSeriErr error
	}

	accountID1 := iotago.MustAccountIDFromHexString("0x7238fbce2f6ae391bd4eb2ce1c51085e0945943bb1bb8e9133e29672c8ef2c74")
	accountID2 := iotago.MustAccountIDFromHexString("0x98f3e0f153461a73f09b6f9eedf7acbd11447f86d6ad20817973a2e2c9240f32")
	accountID3 := iotago.MustAccountIDFromHexString("0xf23ae970dc1359ff48f4169b7cec237873992dc30d9eeb6ccacdecc7679e4f69")

	tests := []allotmentDeSerializeTest{
		{
			name: "ok - multiple unique allotments in order",
			source: iotago.TxEssenceAllotments{
				&iotago.Allotment{
					AccountID: accountID1,
					Mana:      5,
				},
				&iotago.Allotment{
					AccountID: accountID2,
					Mana:      4,
				},
				&iotago.Allotment{
					AccountID: accountID3,
					Mana:      6,
				},
			},
		},
		{
			name: "err - account id in allotments not lexically ordered",
			source: iotago.TxEssenceAllotments{
				&iotago.Allotment{
					AccountID: accountID2,
					Mana:      500,
				},
				&iotago.Allotment{
					AccountID: accountID1,
					Mana:      800,
				},
				&iotago.Allotment{
					AccountID: accountID3,
					Mana:      800,
				},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deSeriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "err - account id in allotments not unique",
			source: iotago.TxEssenceAllotments{
				&iotago.Allotment{
					AccountID: accountID1,
					Mana:      500,
				},
				&iotago.Allotment{
					AccountID: accountID1,
					Mana:      800,
				},
			},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deSeriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, test := range tests {
		stx := tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI, &iotago.Transaction{
			API: tpkg.ZeroCostTestAPI,
			TransactionEssence: &iotago.TransactionEssence{
				Allotments:    test.source,
				Capabilities:  iotago.TransactionCapabilitiesBitMaskWithCapabilities(),
				ContextInputs: iotago.TxEssenceContextInputs{},
				NetworkID:     tpkg.ZeroCostTestAPI.ProtocolParameters().NetworkID(),
				Inputs: iotago.TxEssenceInputs{
					tpkg.RandUTXOInput(),
				},
			},
			Outputs: iotago.TxEssenceOutputs{
				tpkg.RandBasicOutput(),
			},
		})

		tst := deSerializeTest{
			name:      test.name,
			source:    stx,
			target:    &iotago.SignedTransaction{},
			seriErr:   test.seriErr,
			deSeriErr: test.deSeriErr,
		}

		t.Run(tst.name, tst.deSerialize)
	}
}
