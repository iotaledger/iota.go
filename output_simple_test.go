package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleOutput_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target serializer.Serializable
		err    error
	}
	tests := []test{
		func() test {
			dep, depData := tpkg.RandSimpleOutput(iotago.AddressEd25519)
			return test{"ok ed25519", depData, dep, nil}
		}(),
		func() test {
			dep, depData := tpkg.RandSimpleOutput(iotago.AddressEd25519)
			return test{"not enough data ed25519", depData[:5], dep, serializer.ErrDeserializationNotEnoughData}
		}(),
		func() test {
			dep, depData := tpkg.RandSimpleOutput(iotago.AddressEd25519)
			depData[iotago.SimpleOutputAddressOffset] = 100
			return test{"unknown addr type", depData, dep, iotago.ErrTypeIsNotSupportedAddress}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := &iotago.SimpleOutput{}
			bytesRead, err := dep.Deserialize(tt.source, serializer.DeSeriModePerformValidation)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, dep)
		})
	}
}

func TestSimpleOutput_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.SimpleOutput
		target []byte
		err    error
	}
	tests := []test{
		func() test {
			dep, depData := tpkg.RandSimpleOutput(iotago.AddressEd25519)
			return test{"ok", dep, depData, nil}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.source.Serialize(serializer.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.EqualValues(t, tt.target, data)
		})
	}
}
