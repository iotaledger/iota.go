package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

var sigUnlockSigGuard = serializer.SerializableGuard{
	ReadGuard: func(ty uint32) (serializer.Serializable, error) {
		return SignatureSelector(ty)
	},
	WriteGuard: func(seri serializer.Serializable) error {
		if seri == nil {
			return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedSignature)
		}
		switch seri.(type) {
		case *Ed25519Signature:
		default:
			return ErrTypeIsNotSupportedSignature
		}

		return nil
	},
}

// SignatureUnlock holds a signature which unlocks inputs.
type SignatureUnlock struct {
	// The signature of this unlock.
	Signature Signature `json:"signature"`
}

func (s *SignatureUnlock) Type() UnlockType {
	return UnlockSignature
}

func (s *SignatureUnlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(UnlockSignature), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize signature unlock: %w", err)
		}).
		ReadObject(&s.Signature, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, sigUnlockSigGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize signature within signature unlock: %w", err)
		}).
		Done()
}

func (s *SignatureUnlock) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(UnlockSignature), func(err error) error {
			return fmt.Errorf("unable to serialize signature unlock type ID: %w", err)
		}).
		WriteObject(s.Signature, deSeriMode, deSeriCtx, sigUnlockSigGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize signature unlock signature: %w", err)
		}).
		Serialize()
}

func (s *SignatureUnlock) Size() int {
	return util.NumByteLen(byte(UnlockSignature)) + s.Signature.Size()
}

func (s *SignatureUnlock) MarshalJSON() ([]byte, error) {
	jSignatureUnlock := &jsonSignatureUnlock{}
	jSignature, err := s.Signature.MarshalJSON()
	if err != nil {
		return nil, err
	}
	rawMsgJSONSig := json.RawMessage(jSignature)
	jSignatureUnlock.Signature = &rawMsgJSONSig
	jSignatureUnlock.Type = int(UnlockSignature)

	return json.Marshal(jSignatureUnlock)
}

func (s *SignatureUnlock) UnmarshalJSON(bytes []byte) error {
	jSignatureUnlock := &jsonSignatureUnlock{}
	if err := json.Unmarshal(bytes, jSignatureUnlock); err != nil {
		return err
	}
	seri, err := jSignatureUnlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*SignatureUnlock)

	return nil
}

// jsonSignatureUnlock defines the json representation of a SignatureUnlock.
type jsonSignatureUnlock struct {
	Type      int              `json:"type"`
	Signature *json.RawMessage `json:"signature"`
}

func (j *jsonSignatureUnlock) ToSerializable() (serializer.Serializable, error) {
	sig, err := signatureFromJSONRawMsg(j.Signature)
	if err != nil {
		return nil, err
	}

	return &SignatureUnlock{Signature: sig}, nil
}
