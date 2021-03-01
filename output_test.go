package iotago_test

import (
	"errors"
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
		target iotago.Serializable
		err    error
	}
	tests := []test{
		func() test {
			dep, depData := randSigLockedSingleOutput(iotago.AddressEd25519)
			return test{"ok ed25519", depData, dep, nil}
		}(),
		func() test {
			dep, depData := randSigLockedSingleOutput(iotago.AddressEd25519)
			return test{"not enough data ed25519", depData[:5], dep, iotago.ErrDeserializationNotEnoughData}
		}(),
		func() test {
			dep, depData := randSigLockedSingleOutput(iotago.AddressEd25519)
			depData[iotago.SigLockedSingleOutputAddressOffset] = 100
			return test{"unknown addr type", depData, dep, iotago.ErrUnknownAddrType}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := &iotago.SigLockedSingleOutput{}
			bytesRead, err := dep.Deserialize(tt.source, iotago.DeSeriModePerformValidation)
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
			dep, depData := randSigLockedSingleOutput(iotago.AddressEd25519)
			return test{"ok", dep, depData, nil}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.source.Serialize(iotago.DeSeriModePerformValidation)
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
		outputs iotago.Serializables
		funcs   []iotago.OutputsValidatorFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"ok addr",
			args{outputs: []iotago.Serializable{
				&iotago.SigLockedSingleOutput{
					Address: func() iotago.Serializable {
						addr, _ := randEd25519Addr()
						return addr
					}(),
					Amount: 0,
				},
				&iotago.SigLockedSingleOutput{
					Address: func() iotago.Serializable {
						addr, _ := randEd25519Addr()
						return addr
					}(),
					Amount: 0,
				},
			}, funcs: []iotago.OutputsValidatorFunc{iotago.OutputsAddrUniqueValidator()}}, false,
		},
		{
			"addr not unique",
			args{outputs: []iotago.Serializable{
				&iotago.SigLockedSingleOutput{
					Address: func() iotago.Serializable {
						addr, _ := randEd25519Addr()
						for i := 0; i < len(addr); i++ {
							addr[i] = 3
						}
						return addr
					}(),
					Amount: 0,
				},
				&iotago.SigLockedSingleOutput{
					Address: func() iotago.Serializable {
						addr, _ := randEd25519Addr()
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
			args{outputs: []iotago.Serializable{
				&iotago.SigLockedSingleOutput{
					Address: nil,
					Amount:  iotago.TokenSupply,
				},
			}, funcs: []iotago.OutputsValidatorFunc{iotago.OutputsDepositAmountValidator()}}, false,
		},
		{
			"spends more than total supply",
			args{outputs: []iotago.Serializable{
				&iotago.SigLockedSingleOutput{
					Address: nil,
					Amount:  iotago.TokenSupply + 1,
				},
			}, funcs: []iotago.OutputsValidatorFunc{iotago.OutputsDepositAmountValidator()}}, true,
		},
		{
			"sum more than total supply",
			args{outputs: []iotago.Serializable{
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
