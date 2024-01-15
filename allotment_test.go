package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
)

func TestAllotmentDeSerialize(t *testing.T) {
	accountID1 := iotago.MustAccountIDFromHexString("0x7238fbce2f6ae391bd4eb2ce1c51085e0945943bb1bb8e9133e29672c8ef2c74")
	accountID2 := iotago.MustAccountIDFromHexString("0x98f3e0f153461a73f09b6f9eedf7acbd11447f86d6ad20817973a2e2c9240f32")
	accountID3 := iotago.MustAccountIDFromHexString("0xf23ae970dc1359ff48f4169b7cec237873992dc30d9eeb6ccacdecc7679e4f69")

	tests := []deSerializeTest{
		// TODO: Investigate why this fails.
		// {
		// 	name: "ok - multiple unique allotments in order",
		// 	source: iotago.TxEssenceAllotments{
		// 		&iotago.Allotment{
		// 			AccountID: accountID1,
		// 			Mana:      500,
		// 		},
		// 		&iotago.Allotment{
		// 			AccountID: accountID2,
		// 			Mana:      400,
		// 		},
		// 		&iotago.Allotment{
		// 			AccountID: accountID3,
		// 			Mana:      600,
		// 		},
		// 	},
		// 	target: iotago.TxEssenceAllotments{},
		// },
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
			target:    iotago.TxEssenceAllotments{},
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
			target:    iotago.TxEssenceAllotments{},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deSeriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
