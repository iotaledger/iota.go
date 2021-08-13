package iotago_test

import (
	"errors"
	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/ed25519"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target serializer.Serializable
		err    error
	}
	tests := []test{
		func() test {
			txPayload, txPayloadData := tpkg.RandTransaction()
			return test{"ok", txPayloadData, txPayload, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &iotago.Transaction{}
			bytesRead, err := tx.Deserialize(tt.source, serializer.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, tx)
		})
	}
}

func TestTransaction_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.Transaction
		target []byte
	}
	tests := []test{
		func() test {
			txPayload, txPayloadData := tpkg.RandTransaction()
			return test{"ok", txPayload, txPayloadData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(serializer.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

func TestTransaction_SemanticallyValidate(t *testing.T) {
	identityOne := tpkg.RandEd25519PrivateKey()
	inputAddr := iotago.AddressFromEd25519PubKey(identityOne.Public().(ed25519.PublicKey))
	addrKeys := iotago.AddressKeys{Address: &inputAddr, Keys: identityOne}

	type test struct {
		name       string
		addrSigner iotago.AddressSigner
		builder    *iotago.TransactionBuilder
		inputUTXOs iotago.InputToOutputMapping
		buildErr   error
		validErr   error
	}

	tests := []test{
		func() test {

			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "ok - 1 input/output",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputToOutputMapping{
					inputUTXO1.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
				},
			}
		}(),
		func() test {

			outputAddr1, _ := tpkg.RandEd25519Address()
			outputAddr2, _ := tpkg.RandEd25519Address()
			outputAddr3, _ := tpkg.RandEd25519Address()
			outputAddr4, _ := tpkg.RandEd25519Address()

			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}
			inputUTXO2 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO2}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 20}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr2, Amount: 10}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr3, Amount: 20}).
				AddOutput(&iotago.SigLockedDustAllowanceOutput{Address: outputAddr4, Amount: 1_000_000})

			return test{
				name:       "ok - 2 inputs, 4 outputs",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputToOutputMapping{
					inputUTXO1.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
					inputUTXO2.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: 1_000_000},
				},
			}
		}(),
		func() test {
			builder := iotago.NewTransactionBuilder()
			return test{
				name:       "err - no inputs",
				addrSigner: nil,
				builder:    builder,
				buildErr:   iotago.ErrMinInputsNotReached,
			}
		}(),
		func() test {
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}
			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1})
			return test{
				name:       "err - no outputs",
				addrSigner: nil,
				builder:    builder,
				buildErr:   iotago.ErrMinOutputsNotReached,
			}
		}(),
		func() test {

			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 100})

			return test{
				name:       "err - input output sum mismatch",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				validErr:   iotago.ErrInputOutputSumMismatch,
				inputUTXOs: iotago.InputToOutputMapping{
					inputUTXO1.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
				},
			}
		}(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			payload, err := test.builder.Build(test.addrSigner)
			if test.buildErr != nil {
				assert.True(t, errors.Is(err, test.buildErr))
				return
			}
			assert.NoError(t, err)

			semanticErr := payload.SemanticallyValidate(test.inputUTXOs)
			if test.validErr != nil {
				assert.True(t, errors.Is(semanticErr, test.validErr))
				return
			}
			assert.NoError(t, semanticErr)

			_, err = payload.Serialize(serializer.DeSeriModePerformValidation)
			assert.NoError(t, err)
		})
	}

}

func TestDustAllowance(t *testing.T) {
	identityOne := tpkg.RandEd25519PrivateKey()
	inputAddr := iotago.AddressFromEd25519PubKey(identityOne.Public().(ed25519.PublicKey))
	addrKeys := iotago.AddressKeys{Address: &inputAddr, Keys: identityOne}

	type test struct {
		name              string
		addrSigner        iotago.AddressSigner
		builder           *iotago.TransactionBuilder
		inputUTXOs        iotago.InputToOutputMapping
		dustAllowanceFunc iotago.DustAllowanceFunc
		buildErr          error
		validErr          error
	}

	tests := []test{
		func() test {

			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "ok - create dust output on address with enough allowance by consuming the only dust output",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputToOutputMapping{
					inputUTXO1.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
				},
				dustAllowanceFunc: func() iotago.DustAllowanceFunc {
					dustOutputsAmount := map[string]int64{
						(&inputAddr).String(): 1,
						// zero on output address
					}
					dustAllowanceSum := map[string]uint64{
						// we have one dust allowance output on the target address
						outputAddr1.String(): iotago.OutputSigLockedDustAllowanceOutputMinDeposit,
					}
					return func(addr iotago.Address) (uint64, int64, error) {
						return dustAllowanceSum[addr.String()], dustOutputsAmount[addr.String()], nil
					}
				}(),
			}
		}(),
		func() test {

			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "ok - create dust output on address with enough allowance while still having dust on both source and target",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputToOutputMapping{
					inputUTXO1.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
				},
				dustAllowanceFunc: func() iotago.DustAllowanceFunc {
					dustOutputsAmount := map[string]int64{
						(&inputAddr).String(): 5,
						outputAddr1.String():  5,
					}
					dustAllowanceSum := map[string]uint64{
						(&inputAddr).String(): iotago.OutputSigLockedDustAllowanceOutputMinDeposit,
						outputAddr1.String():  iotago.OutputSigLockedDustAllowanceOutputMinDeposit,
					}
					return func(addr iotago.Address) (uint64, int64, error) {
						return dustAllowanceSum[addr.String()], dustOutputsAmount[addr.String()], nil
					}
				}(),
			}
		}(),
		func() test {

			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 50}).
				AddOutput(&iotago.SigLockedDustAllowanceOutput{Address: outputAddr1, Amount: iotago.OutputSigLockedDustAllowanceOutputMinDeposit})

			return test{
				name:       "ok - create dust output on address with enough allowance through companion dust allowance output",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputToOutputMapping{
					inputUTXO1.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: iotago.OutputSigLockedDustAllowanceOutputMinDeposit + 50},
				},
				dustAllowanceFunc: func() iotago.DustAllowanceFunc {
					dustOutputsAmount := map[string]int64{
						(&inputAddr).String(): 0,
						outputAddr1.String():  0,
					}
					dustAllowanceSum := map[string]uint64{
						(&inputAddr).String(): 0,
						// explicit, we have no dust allowance output on the target
						// but we will have it via the transaction
						outputAddr1.String(): 0,
					}
					return func(addr iotago.Address) (uint64, int64, error) {
						return dustAllowanceSum[addr.String()], dustOutputsAmount[addr.String()], nil
					}
				}(),
			}
		}(),
		func() test {

			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}
			inputUTXO2 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO2}).
				AddOutput(&iotago.SigLockedDustAllowanceOutput{Address: outputAddr1, Amount: iotago.OutputSigLockedDustAllowanceOutputMinDeposit})

			return test{
				name:       "ok - create dust allowance output by combining multiple dust outputs as inputs",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputToOutputMapping{
					inputUTXO1.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: iotago.OutputSigLockedDustAllowanceOutputMinDeposit / 2},
					inputUTXO2.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: iotago.OutputSigLockedDustAllowanceOutputMinDeposit / 2},
				},
				dustAllowanceFunc: func() iotago.DustAllowanceFunc {
					dustOutputsAmount := map[string]int64{
						(&inputAddr).String(): 2,
						outputAddr1.String():  0,
					}
					dustAllowanceSum := map[string]uint64{
						(&inputAddr).String(): iotago.OutputSigLockedDustAllowanceOutputMinDeposit,
						outputAddr1.String():  0,
					}
					return func(addr iotago.Address) (uint64, int64, error) {
						return dustAllowanceSum[addr.String()], dustOutputsAmount[addr.String()], nil
					}
				}(),
			}
		}(),
		func() test {

			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "err - create dust output on address without enough allowance",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputToOutputMapping{
					inputUTXO1.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
				},
				dustAllowanceFunc: func() iotago.DustAllowanceFunc {
					dustOutputsAmount := map[string]int64{
						(&inputAddr).String(): 1,
						outputAddr1.String():  0,
					}
					dustAllowanceSum := map[string]uint64{
						// we are allowed to have outputs on our input address
						(&inputAddr).String(): iotago.OutputSigLockedDustAllowanceOutputMinDeposit,
						// this should result in an error, since no dust allowance is present
						outputAddr1.String(): 0,
					}
					return func(addr iotago.Address) (uint64, int64, error) {
						return dustAllowanceSum[addr.String()], dustOutputsAmount[addr.String()], nil
					}
				}(),
				validErr: iotago.ErrInvalidDustAllowance,
			}
		}(),
		func() test {

			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "err - create dust output on address exceeding allowance by 1",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputToOutputMapping{
					inputUTXO1.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
				},
				dustAllowanceFunc: func() iotago.DustAllowanceFunc {
					dustOutputsAmount := map[string]int64{
						(&inputAddr).String(): 1,
						outputAddr1.String():  100,
					}
					dustAllowanceSum := map[string]uint64{
						(&inputAddr).String(): iotago.OutputSigLockedDustAllowanceOutputMinDeposit,
						// this should result in an error, since we're creating one
						// dust output over the limit
						outputAddr1.String(): iotago.OutputSigLockedDustAllowanceOutputMinDeposit,
					}
					return func(addr iotago.Address) (uint64, int64, error) {
						return dustAllowanceSum[addr.String()], dustOutputsAmount[addr.String()], nil
					}
				}(),
				validErr: iotago.ErrInvalidDustAllowance,
			}
		}(),
		func() test {

			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedDustAllowanceOutput{Address: outputAddr1, Amount: iotago.OutputSigLockedDustAllowanceOutputMinDeposit - 1})

			return test{
				name:       "err - create dust allowance output which does not reach the minimum dust allowance amount",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputToOutputMapping{
					inputUTXO1.ID(): &iotago.SigLockedSingleOutput{Address: &inputAddr, Amount: iotago.OutputSigLockedDustAllowanceOutputMinDeposit},
				},
				dustAllowanceFunc: func() iotago.DustAllowanceFunc {
					dustOutputsAmount := map[string]int64{
						(&inputAddr).String(): 0,
						outputAddr1.String():  0,
					}
					dustAllowanceSum := map[string]uint64{
						(&inputAddr).String(): 0,
						// no dust allowance yet on target
						outputAddr1.String(): 0,
					}
					return func(addr iotago.Address) (uint64, int64, error) {
						return dustAllowanceSum[addr.String()], dustOutputsAmount[addr.String()], nil
					}
				}(),
				buildErr: iotago.ErrOutputDustAllowanceLessThanMinDeposit,
			}
		}(),
		func() test {

			outputAddr1, _ := tpkg.RandEd25519Address()
			inputUTXO1 := &iotago.UTXOInput{TransactionID: tpkg.Rand32ByteArray(), TransactionOutputIndex: 0}

			builder := iotago.NewTransactionBuilder().
				AddInput(&iotago.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iotago.SigLockedSingleOutput{Address: outputAddr1, Amount: iotago.OutputSigLockedDustAllowanceOutputMinDeposit})

			return test{
				name:       "err - consume dust allowance output without enough remaining allowance",
				addrSigner: iotago.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iotago.InputToOutputMapping{
					inputUTXO1.ID(): &iotago.SigLockedDustAllowanceOutput{Address: &inputAddr, Amount: iotago.OutputSigLockedDustAllowanceOutputMinDeposit},
				},
				dustAllowanceFunc: func() iotago.DustAllowanceFunc {
					dustOutputsAmount := map[string]int64{
						(&inputAddr).String(): 50,
						outputAddr1.String():  0,
					}
					dustAllowanceSum := map[string]uint64{
						// we spend the only dust allowance output on the address while still
						// having 50 dust outputs on it
						(&inputAddr).String(): iotago.OutputSigLockedDustAllowanceOutputMinDeposit,
						outputAddr1.String():  0,
					}
					return func(addr iotago.Address) (uint64, int64, error) {
						return dustAllowanceSum[addr.String()], dustOutputsAmount[addr.String()], nil
					}
				}(),
				validErr: iotago.ErrInvalidDustAllowance,
			}
		}(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			payload, err := test.builder.Build(test.addrSigner)
			if test.buildErr != nil {
				assert.True(t, errors.Is(err, test.buildErr))
				return
			}
			assert.NoError(t, err)

			semanticErr := payload.SemanticallyValidate(
				test.inputUTXOs,
				iotago.NewDustSemanticValidation(iotago.DustAllowanceDivisor, iotago.MaxDustOutputsOnAddress, test.dustAllowanceFunc),
			)

			if test.validErr != nil {
				assert.True(t, errors.Is(semanticErr, test.validErr))
				return
			}
			assert.NoError(t, semanticErr)

			_, err = payload.Serialize(serializer.DeSeriModePerformValidation)
			assert.NoError(t, err)
		})
	}
}
