package iota_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
)

func TestOutputSelector(t *testing.T) {
	_, err := iota.OutputSelector(100)
	assert.True(t, errors.Is(err, iota.ErrUnknownOutputType))
}

func TestSigLockedSingleOutput_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iota.Serializable
		err    error
	}
	tests := []test{
		func() test {
			dep, depData := randSigLockedSingleOutput(iota.AddressWOTS)
			return test{"ok wots", depData, dep, nil}
		}(),
		func() test {
			dep, depData := randSigLockedSingleOutput(iota.AddressWOTS)
			return test{"not enough data wots", depData[:5], dep, iota.ErrDeserializationNotEnoughData}
		}(),
		func() test {
			dep, depData := randSigLockedSingleOutput(iota.AddressEd25519)
			return test{"ok ed25519", depData, dep, nil}
		}(),
		func() test {
			dep, depData := randSigLockedSingleOutput(iota.AddressEd25519)
			return test{"not enough data ed25519", depData[:5], dep, iota.ErrDeserializationNotEnoughData}
		}(),
		func() test {
			dep, depData := randSigLockedSingleOutput(iota.AddressEd25519)
			depData[iota.SigLockedSingleOutputAddressOffset] = 100
			return test{"unknown addr type", depData, dep, iota.ErrUnknownAddrType}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := &iota.SigLockedSingleOutput{}
			bytesRead, err := dep.Deserialize(tt.source, iota.DeSeriModePerformValidation)
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
		source *iota.SigLockedSingleOutput
		target []byte
		err    error
	}
	tests := []test{
		func() test {
			dep, depData := randSigLockedSingleOutput(iota.AddressEd25519)
			return test{"ok", dep, depData, nil}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
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
		outputs iota.Serializables
		funcs   []iota.OutputsValidatorFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"ok addr",
			args{outputs: []iota.Serializable{
				&iota.SigLockedSingleOutput{
					Address: func() iota.Serializable {
						addr, _ := randEd25519Addr()
						return addr
					}(),
					Amount: 0,
				},
				&iota.SigLockedSingleOutput{
					Address: func() iota.Serializable {
						addr, _ := randEd25519Addr()
						return addr
					}(),
					Amount: 0,
				},
			}, funcs: []iota.OutputsValidatorFunc{iota.OutputsAddrUniqueValidator()}}, false,
		},
		{
			"addr not unique",
			args{outputs: []iota.Serializable{
				&iota.SigLockedSingleOutput{
					Address: func() iota.Serializable {
						addr, _ := randEd25519Addr()
						for i := 0; i < len(addr); i++ {
							addr[i] = 3
						}
						return addr
					}(),
					Amount: 0,
				},
				&iota.SigLockedSingleOutput{
					Address: func() iota.Serializable {
						addr, _ := randEd25519Addr()
						for i := 0; i < len(addr); i++ {
							addr[i] = 3
						}
						return addr
					}(),
					Amount: 0,
				},
			}, funcs: []iota.OutputsValidatorFunc{iota.OutputsAddrUniqueValidator()}}, true,
		},
		{
			"ok amount",
			args{outputs: []iota.Serializable{
				&iota.SigLockedSingleOutput{
					Address: nil,
					Amount:  iota.TokenSupply,
				},
			}, funcs: []iota.OutputsValidatorFunc{iota.OutputsDepositAmountValidator()}}, false,
		},
		{
			"spends more than total supply",
			args{outputs: []iota.Serializable{
				&iota.SigLockedSingleOutput{
					Address: nil,
					Amount:  iota.TokenSupply + 1,
				},
			}, funcs: []iota.OutputsValidatorFunc{iota.OutputsDepositAmountValidator()}}, true,
		},
		{
			"sum more than total supply",
			args{outputs: []iota.Serializable{
				&iota.SigLockedSingleOutput{
					Address: nil,
					Amount:  iota.TokenSupply - 1,
				},
				&iota.SigLockedSingleOutput{
					Address: nil,
					Amount:  iota.TokenSupply - 1,
				},
			}, funcs: []iota.OutputsValidatorFunc{iota.OutputsDepositAmountValidator()}}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := iota.ValidateOutputs(tt.args.outputs, tt.args.funcs...); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
