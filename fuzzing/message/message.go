package message

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/iota.go"
)

func Fuzz(data []byte) int {
	m := &iota.Message{}
	_, err := m.Deserialize(data, iota.DeSeriModeNoValidation)
	if err != nil {
		return 0
	}
	seriData, err := m.Serialize(iota.DeSeriModeNoValidation)
	if err != nil {
		panic(fmt.Sprintf("should be able to serialize: %q", err))
	}
	if !bytes.Equal(data[:len(seriData)], seriData) {
		panic("data from serialization should be same as origin")
	}
	return 1
}
