package iotago_test

import (
	"errors"
	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

func TestOutputSelector(t *testing.T) {
	_, err := iotago.OutputSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownOutputType))
}

func TestSigLockedSingleOutput_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target serializer.Serializable
		err    error
	}
	tests := []test{
		func() test {
			dep, depData := tpkg.RandSigLockedSingleOutput(iotago.AddressEd25519)
			return test{"ok ed25519", depData, dep, nil}
		}(),
		func() test {
			dep, depData := tpkg.RandSigLockedSingleOutput(iotago.AddressEd25519)
			return test{"not enough data ed25519", depData[:5], dep, serializer.ErrDeserializationNotEnoughData}
		}(),
		func() test {
			dep, depData := tpkg.RandSigLockedSingleOutput(iotago.AddressEd25519)
			depData[iotago.SigLockedSingleOutputAddressOffset] = 100
			return test{"unknown addr type", depData, dep, iotago.ErrUnknownAddrType}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := &iotago.SigLockedSingleOutput{}
			bytesRead, err := dep.Deserialize(tt.source, serializer.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, dep)
		})
	}
}

func TestSigLockedSingleOutput_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.SigLockedSingleOutput
		target []byte
		err    error
	}
	tests := []test{
		func() test {
			dep, depData := tpkg.RandSigLockedSingleOutput(iotago.AddressEd25519)
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

func TestOutputsValidatorFunc(t *testing.T) {
	type args struct {
		outputs serializer.Serializables
		funcs   []iotago.OutputsValidatorFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"ok addr",
			args{outputs: []serializer.Serializable{
				&iotago.SigLockedSingleOutput{
					Address: func() serializer.Serializable {
						addr, _ := tpkg.RandEd25519Address()
						return addr
					}(),
					Amount: 0,
				},
				&iotago.SigLockedSingleOutput{
					Address: func() serializer.Serializable {
						addr, _ := tpkg.RandEd25519Address()
						return addr
					}(),
					Amount: 0,
				},
			}, funcs: []iotago.OutputsValidatorFunc{iotago.OutputsAddrUniqueValidator()}}, false,
		},
		{
			"addr not unique",
			args{outputs: []serializer.Serializable{
				&iotago.SigLockedSingleOutput{
					Address: func() serializer.Serializable {
						addr, _ := tpkg.RandEd25519Address()
						for i := 0; i < len(addr); i++ {
							addr[i] = 3
						}
						return addr
					}(),
					Amount: 0,
				},
				&iotago.SigLockedSingleOutput{
					Address: func() serializer.Serializable {
						addr, _ := tpkg.RandEd25519Address()
						for i := 0; i < len(addr); i++ {
							addr[i] = 3
						}
						return addr
					}(),
					Amount: 0,
				},
			}, funcs: []iotago.OutputsValidatorFunc{iotago.OutputsAddrUniqueValidator()}}, true,
		},
		{
			"ok amount",
			args{outputs: []serializer.Serializable{
				&iotago.SigLockedSingleOutput{
					Address: nil,
					Amount:  iotago.TokenSupply,
				},
			}, funcs: []iotago.OutputsValidatorFunc{iotago.OutputsDepositAmountValidator()}}, false,
		},
		{
			"spends more than total supply",
			args{outputs: []serializer.Serializable{
				&iotago.SigLockedSingleOutput{
					Address: nil,
					Amount:  iotago.TokenSupply + 1,
				},
			}, funcs: []iotago.OutputsValidatorFunc{iotago.OutputsDepositAmountValidator()}}, true,
		},
		{
			"sum more than total supply",
			args{outputs: []serializer.Serializable{
				&iotago.SigLockedSingleOutput{
					Address: nil,
					Amount:  iotago.TokenSupply - 1,
				},
				&iotago.SigLockedSingleOutput{
					Address: nil,
					Amount:  iotago.TokenSupply - 1,
				},
			}, funcs: []iotago.OutputsValidatorFunc{iotago.OutputsDepositAmountValidator()}}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := iotago.ValidateOutputs(tt.args.outputs, tt.args.funcs...); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
