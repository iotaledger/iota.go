package message

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/iota.go/v2"
)

func Fuzz(data []byte) int {
	m := &iotago.Message{}
	if _, err := m.Deserialize(data, iotago.DeSeriModePerformValidation); err != nil {
		return 0
	}
	seriData, err := m.Serialize(iotago.DeSeriModePerformValidation)
	if err != nil {
		panic(fmt.Sprintf("should be able to serialize: %q", err))
	}
	if !bytes.Equal(data[:len(seriData)], seriData) {
		panic("data from serialization should be same as origin")
	}
	return 1
}
