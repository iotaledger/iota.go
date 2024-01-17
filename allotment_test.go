package iotago_test

import (
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/stretchr/testify/require"
)

func TestAllotmentDeSerialize(t *testing.T) {
	type allotmentDeSerializeTest struct {
		name      string
		source    iotago.TxEssenceAllotments
		target    *iotago.TxEssenceAllotments
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
			target: &iotago.TxEssenceAllotments{},
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
			target:    &iotago.TxEssenceAllotments{},
			seriErr:   serix.ErrArrayValidationOrderViolatesLexicalOrder,
			deSeriErr: serix.ErrArrayValidationOrderViolatesLexicalOrder,
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
			target:    &iotago.TxEssenceAllotments{},
			seriErr:   serix.ErrArrayValidationViolatesUniqueness,
			deSeriErr: serix.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serixData, err := tpkg.ZeroCostTestAPI.Encode(test.source, serix.WithValidation())
			if test.seriErr != nil {
				require.ErrorIs(t, err, test.seriErr)

				// Encode again without validation so we can test decoding.
				serixData, err = tpkg.ZeroCostTestAPI.Encode(test.source)
				require.NoError(t, err)
			} else {
				require.NoError(t, err)
			}

			bytesRead, err := tpkg.ZeroCostTestAPI.Decode(serixData, test.target, serix.WithValidation())
			if test.deSeriErr != nil {
				require.ErrorIs(t, err, test.deSeriErr)

				return
			}
			require.NoError(t, err)
			require.Len(t, serixData, bytesRead)
			require.EqualValues(t, test.source, *test.target)
		})
	}
}
