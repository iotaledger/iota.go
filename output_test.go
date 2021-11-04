package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/assert"
)

func TestOutputSelector(t *testing.T) {
	_, err := iotago.OutputSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownOutputType))
}

func TestOutputsValidatorFunc(t *testing.T) {
	type args struct {
		outputs serializer.Serializables
		funcs   []iotago.OutputsPredicateFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"ok addr",
			args{outputs: []serializer.Serializable{
				&iotago.SimpleOutput{
					Address: func() serializer.Serializable {
						addr, _ := tpkg.RandEd25519Address()
						return addr
					}(),
					Amount: 0,
				},
				&iotago.SimpleOutput{
					Address: func() serializer.Serializable {
						addr, _ := tpkg.RandEd25519Address()
						return addr
					}(),
					Amount: 0,
				},
			}, funcs: []iotago.OutputsPredicateFunc{iotago.OutputsPredicateAddrUnique()}}, false,
		},
		{
			"addr not unique",
			args{outputs: []serializer.Serializable{
				&iotago.SimpleOutput{
					Address: func() serializer.Serializable {
						addr, _ := tpkg.RandEd25519Address()
						for i := 0; i < len(addr); i++ {
							addr[i] = 3
						}
						return addr
					}(),
					Amount: 0,
				},
				&iotago.SimpleOutput{
					Address: func() serializer.Serializable {
						addr, _ := tpkg.RandEd25519Address()
						for i := 0; i < len(addr); i++ {
							addr[i] = 3
						}
						return addr
					}(),
					Amount: 0,
				},
			}, funcs: []iotago.OutputsPredicateFunc{iotago.OutputsPredicateAddrUnique()}}, true,
		},
		{
			"ok amount",
			args{outputs: []serializer.Serializable{
				&iotago.SimpleOutput{
					Address: nil,
					Amount:  iotago.TokenSupply,
				},
			}, funcs: []iotago.OutputsPredicateFunc{iotago.OutputsPredicateDepositAmount()}}, false,
		},
		{
			"spends more than total supply",
			args{outputs: []serializer.Serializable{
				&iotago.SimpleOutput{
					Address: nil,
					Amount:  iotago.TokenSupply + 1,
				},
			}, funcs: []iotago.OutputsPredicateFunc{iotago.OutputsPredicateDepositAmount()}}, true,
		},
		{
			"sum more than total supply",
			args{outputs: []serializer.Serializable{
				&iotago.SimpleOutput{
					Address: nil,
					Amount:  iotago.TokenSupply - 1,
				},
				&iotago.SimpleOutput{
					Address: nil,
					Amount:  iotago.TokenSupply - 1,
				},
			}, funcs: []iotago.OutputsPredicateFunc{iotago.OutputsPredicateDepositAmount()}}, true,
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
