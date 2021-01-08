package iota_test

import (
	"crypto/ed25519"
	"errors"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iota.Serializable
		err    error
	}
	tests := []test{
		func() test {
			txPayload, txPayloadData := randTransaction()
			return test{"ok", txPayloadData, txPayload, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &iota.Transaction{}
			bytesRead, err := tx.Deserialize(tt.source, iota.DeSeriModePerformValidation)
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
		source *iota.Transaction
		target []byte
	}
	tests := []test{
		func() test {
			txPayload, txPayloadData := randTransaction()
			return test{"ok", txPayload, txPayloadData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

func TestTransaction_SemanticallyValidate(t *testing.T) {
	identityOne := randEd25519PrivateKey()
	inputAddr := iota.AddressFromEd25519PubKey(identityOne.Public().(ed25519.PublicKey))
	addrKeys := iota.AddressKeys{Address: &inputAddr, Keys: identityOne}

	type test struct {
		name       string
		addrSigner iota.AddressSigner
		builder    *iota.TransactionBuilder
		inputUTXOs iota.InputToOutputMapping
		buildErr   error
		validErr   error
	}

	tests := []test{
		func() test {

			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "ok - 1 input/output",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
				},
			}
		}(),
		func() test {

			outputAddr1, _ := randEd25519Addr()
			outputAddr2, _ := randEd25519Addr()
			outputAddr3, _ := randEd25519Addr()
			outputAddr4, _ := randEd25519Addr()

			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}
			inputUTXO2 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO2}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 20}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr2, Amount: 10}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr3, Amount: 20}).
				AddOutput(&iota.SigLockedDustAllowanceOutput{Address: outputAddr4, Amount: 1_000_000})

			return test{
				name:       "ok - 2 inputs, 4 outputs",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
					inputUTXO2.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: 1_000_000},
				},
			}
		}(),
		func() test {
			builder := iota.NewTransactionBuilder()
			return test{
				name:       "err - no inputs",
				addrSigner: nil,
				builder:    builder,
				buildErr:   iota.ErrMinInputsNotReached,
			}
		}(),
		func() test {
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}
			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1})
			return test{
				name:       "err - no outputs",
				addrSigner: nil,
				builder:    builder,
				buildErr:   iota.ErrMinOutputsNotReached,
			}
		}(),
		func() test {

			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 100})

			return test{
				name:       "err - input output sum mismatch",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				validErr:   iota.ErrInputOutputSumMismatch,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
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

			_, err = payload.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
		})
	}

}

func TestDustAllowance(t *testing.T) {
	identityOne := randEd25519PrivateKey()
	inputAddr := iota.AddressFromEd25519PubKey(identityOne.Public().(ed25519.PublicKey))
	addrKeys := iota.AddressKeys{Address: &inputAddr, Keys: identityOne}

	type test struct {
		name                        string
		addrSigner                  iota.AddressSigner
		builder                     *iota.TransactionBuilder
		inputUTXOs                  iota.InputToOutputMapping
		numDustOutputsFunc          iota.NumDustOutputsFunc
		dustAllowanceDepositSumFunc iota.DustAllowanceDepositSumFunc
		buildErr                    error
		validErr                    error
	}

	tests := []test{
		func() test {

			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "ok - create dust output on address with enough allowance by consuming the only dust output",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
				},
				numDustOutputsFunc: func() iota.NumDustOutputsFunc {
					m := map[iota.Serializable]int64{
						&inputAddr: 1,
						// zero on output address
					}
					return func(addr iota.Serializable) (int64, error) {
						return m[addr], nil
					}
				}(),
				dustAllowanceDepositSumFunc: func() iota.DustAllowanceDepositSumFunc {
					m := map[iota.Serializable]uint64{
						// we have one dust allowance output on the target address
						outputAddr1: iota.OutputSigLockedDustAllowanceOutputMinDeposit,
					}
					return func(addr iota.Serializable) (uint64, error) {
						return m[addr], nil
					}
				}(),
			}
		}(),
		func() test {

			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "ok - create dust output on address with enough allowance while still having dust on both source and target",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
				},
				numDustOutputsFunc: func() iota.NumDustOutputsFunc {
					// both addresses have dust already on them
					m := map[iota.Serializable]int64{
						&inputAddr:  5,
						outputAddr1: 10,
					}
					return func(addr iota.Serializable) (int64, error) {
						return m[addr], nil
					}
				}(),
				dustAllowanceDepositSumFunc: func() iota.DustAllowanceDepositSumFunc {
					// and both allow for 100 dust outputs to exist
					m := map[iota.Serializable]uint64{
						&inputAddr:  iota.OutputSigLockedDustAllowanceOutputMinDeposit,
						outputAddr1: iota.OutputSigLockedDustAllowanceOutputMinDeposit,
					}
					return func(addr iota.Serializable) (uint64, error) {
						return m[addr], nil
					}
				}(),
			}
		}(),
		func() test {

			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 50}).
				AddOutput(&iota.SigLockedDustAllowanceOutput{Address: outputAddr1, Amount: iota.OutputSigLockedDustAllowanceOutputMinDeposit})

			return test{
				name:       "ok - create dust output on address with enough allowance through companion dust allowance output",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: iota.OutputSigLockedDustAllowanceOutputMinDeposit + 50},
				},
				numDustOutputsFunc: func() iota.NumDustOutputsFunc {
					m := map[iota.Serializable]int64{
						&inputAddr:  0,
						outputAddr1: 0,
					}
					return func(addr iota.Serializable) (int64, error) {
						return m[addr], nil
					}
				}(),
				dustAllowanceDepositSumFunc: func() iota.DustAllowanceDepositSumFunc {
					m := map[iota.Serializable]uint64{
						&inputAddr: 0,
						// explicit, we have no dust allowance output on the target
						// but we will have it via the transaction
						outputAddr1: 0,
					}
					return func(addr iota.Serializable) (uint64, error) {
						return m[addr], nil
					}
				}(),
			}
		}(),
		func() test {

			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}
			inputUTXO2 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO2}).
				AddOutput(&iota.SigLockedDustAllowanceOutput{Address: outputAddr1, Amount: iota.OutputSigLockedDustAllowanceOutputMinDeposit})

			return test{
				name:       "ok - create dust allowance output by combining multiple dust outputs as inputs",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: iota.OutputSigLockedDustAllowanceOutputMinDeposit / 2},
					inputUTXO2.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: iota.OutputSigLockedDustAllowanceOutputMinDeposit / 2},
				},
				numDustOutputsFunc: func() iota.NumDustOutputsFunc {
					m := map[iota.Serializable]int64{
						&inputAddr:  2,
						outputAddr1: 0,
					}
					return func(addr iota.Serializable) (int64, error) {
						return m[addr], nil
					}
				}(),
				dustAllowanceDepositSumFunc: func() iota.DustAllowanceDepositSumFunc {
					m := map[iota.Serializable]uint64{
						&inputAddr:  iota.OutputSigLockedDustAllowanceOutputMinDeposit,
						outputAddr1: 0,
					}
					return func(addr iota.Serializable) (uint64, error) {
						return m[addr], nil
					}
				}(),
			}
		}(),
		func() test {

			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "err - create dust output on address without enough allowance",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
				},
				numDustOutputsFunc: func() iota.NumDustOutputsFunc {
					m := map[iota.Serializable]int64{
						&inputAddr:  1,
						outputAddr1: 0,
					}
					return func(addr iota.Serializable) (int64, error) {
						return m[addr], nil
					}
				}(),
				dustAllowanceDepositSumFunc: func() iota.DustAllowanceDepositSumFunc {
					m := map[iota.Serializable]uint64{
						// we are allowed to have outputs on our input address
						&inputAddr: iota.OutputSigLockedDustAllowanceOutputMinDeposit,
						// this should result in an error, since no dust allowance is present
						outputAddr1: 0,
					}
					return func(addr iota.Serializable) (uint64, error) {
						return m[addr], nil
					}
				}(),
				validErr: iota.ErrInvalidDustAllowance,
			}
		}(),
		func() test {

			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: 50})

			return test{
				name:       "err - create dust output on address exceeding allowance by 1",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: 50},
				},
				numDustOutputsFunc: func() iota.NumDustOutputsFunc {
					m := map[iota.Serializable]int64{
						&inputAddr:  1,
						outputAddr1: 100,
					}
					return func(addr iota.Serializable) (int64, error) {
						return m[addr], nil
					}
				}(),
				dustAllowanceDepositSumFunc: func() iota.DustAllowanceDepositSumFunc {
					m := map[iota.Serializable]uint64{
						&inputAddr: iota.OutputSigLockedDustAllowanceOutputMinDeposit,
						// this should result in an error, since we're creating one
						// dust output over the limit
						outputAddr1: iota.OutputSigLockedDustAllowanceOutputMinDeposit,
					}
					return func(addr iota.Serializable) (uint64, error) {
						return m[addr], nil
					}
				}(),
				validErr: iota.ErrInvalidDustAllowance,
			}
		}(),
		func() test {

			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedDustAllowanceOutput{Address: outputAddr1, Amount: iota.OutputSigLockedDustAllowanceOutputMinDeposit - 1})

			return test{
				name:       "err - create dust allowance output which does not reach the minimum dust allowance amount",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): &iota.SigLockedSingleOutput{Address: &inputAddr, Amount: iota.OutputSigLockedDustAllowanceOutputMinDeposit},
				},
				numDustOutputsFunc: func() iota.NumDustOutputsFunc {
					m := map[iota.Serializable]int64{
						&inputAddr:  0,
						outputAddr1: 0,
					}
					return func(addr iota.Serializable) (int64, error) {
						return m[addr], nil
					}
				}(),
				dustAllowanceDepositSumFunc: func() iota.DustAllowanceDepositSumFunc {
					m := map[iota.Serializable]uint64{
						&inputAddr: 0,
						// no dust allowance yet on target
						outputAddr1: 0,
					}
					return func(addr iota.Serializable) (uint64, error) {
						return m[addr], nil
					}
				}(),
				buildErr: iota.ErrOutputDustAllowanceLessThanMinDeposit,
			}
		}(),
		func() test {

			outputAddr1, _ := randEd25519Addr()
			inputUTXO1 := &iota.UTXOInput{TransactionID: rand32ByteHash(), TransactionOutputIndex: 0}

			builder := iota.NewTransactionBuilder().
				AddInput(&iota.ToBeSignedUTXOInput{Address: &inputAddr, Input: inputUTXO1}).
				AddOutput(&iota.SigLockedSingleOutput{Address: outputAddr1, Amount: iota.OutputSigLockedDustAllowanceOutputMinDeposit})

			return test{
				name:       "err - consume dust allowance output without enough remaining allowance",
				addrSigner: iota.NewInMemoryAddressSigner(addrKeys),
				builder:    builder,
				inputUTXOs: iota.InputToOutputMapping{
					inputUTXO1.ID(): &iota.SigLockedDustAllowanceOutput{Address: &inputAddr, Amount: iota.OutputSigLockedDustAllowanceOutputMinDeposit},
				},
				numDustOutputsFunc: func() iota.NumDustOutputsFunc {
					m := map[iota.Serializable]int64{
						&inputAddr:  50,
						outputAddr1: 0,
					}
					return func(addr iota.Serializable) (int64, error) {
						return m[addr], nil
					}
				}(),
				dustAllowanceDepositSumFunc: func() iota.DustAllowanceDepositSumFunc {
					m := map[iota.Serializable]uint64{
						// we spend the only dust allowance output on the address while still
						// having 50 dust outputs on it
						&inputAddr:  iota.OutputSigLockedDustAllowanceOutputMinDeposit,
						outputAddr1: 0,
					}
					return func(addr iota.Serializable) (uint64, error) {
						return m[addr], nil
					}
				}(),
				validErr: iota.ErrInvalidDustAllowance,
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
				iota.NewDustSemanticValidation(iota.DustAllowanceDivisor, test.numDustOutputsFunc, test.dustAllowanceDepositSumFunc),
			)

			if test.validErr != nil {
				assert.True(t, errors.Is(semanticErr, test.validErr))
				return
			}
			assert.NoError(t, semanticErr)

			_, err = payload.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
		})
	}
}
