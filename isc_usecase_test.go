package iotago_test

import (
	"crypto/ed25519"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
)

func TestConsumeRequest(t *testing.T) {
	stateController := tpkg.RandEd25519PrivateKey()
	stateControllerAddr := iotago.Ed25519AddressFromPubKey(stateController.Public().(ed25519.PublicKey))
	addrKeys := iotago.AddressKeys{Address: &stateControllerAddr, Keys: stateController}

	aliasOut1 := &iotago.AliasOutput{
		Amount:               1337,
		AliasID:              tpkg.RandAliasAddress().AliasID(),
		StateController:      &stateControllerAddr,
		GovernanceController: &stateControllerAddr,
		StateIndex:           1,
	}
	aliasOut1Inp := tpkg.RandUTXOInput()

	req := &iotago.ExtendedOutput{
		Amount:  1337,
		Address: aliasOut1.AliasID.ToAddress(),
	}
	reqInp := tpkg.RandUTXOInput()

	aliasOut2 := &iotago.AliasOutput{
		Amount:               1337 * 2,
		AliasID:              aliasOut1.AliasID,
		StateController:      &stateControllerAddr,
		GovernanceController: &stateControllerAddr,
		StateIndex:           2,
	}
	txb := iotago.NewTransactionBuilder()
	txb.AddInput(&iotago.ToBeSignedUTXOInput{
		Address: &stateControllerAddr,
		Input:   aliasOut1Inp,
	})
	txb.AddInput(&iotago.ToBeSignedUTXOInput{
		Address: &stateControllerAddr,
		Input:   reqInp,
	})
	txb.AddOutput(aliasOut2)
	deSeriParams := &iotago.DeSerializationParameters{
		RentStructure: &iotago.RentStructure{1, 1, 1},
	}
	signer := iotago.NewInMemoryAddressSigner(addrKeys)
	tx, err := txb.Build(deSeriParams, signer)
	require.NoError(t, err)

	semValCtx := &iotago.SemanticValidationContext{
		ExtParas: &iotago.ExternalUnlockParameters{
			ConfMsIndex: 1,
			ConfUnix:    uint64(time.Now().Unix()),
		},
	}
	outset := iotago.OutputSet{
		aliasOut1Inp.ID(): aliasOut1,
		reqInp.ID():       req,
	}
	err = tx.SemanticallyValidate(semValCtx, outset)
	require.NoError(t, err)
}
