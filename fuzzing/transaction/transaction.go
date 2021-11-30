package transaction

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3"
)

func Fuzz(data []byte) int {
	m := &iotago.Transaction{}
	if _, err := m.Deserialize(data, serializer.DeSeriModePerformValidation, nil); err != nil {
		return 0
	}
	seriData, err := m.Serialize(serializer.DeSeriModePerformValidation, nil)
	if err != nil {
		panic(fmt.Sprintf("should be able to serialize: %q", err))
	}
	if !bytes.Equal(data[:len(seriData)], seriData) {
		panic("data from serialization should be same as origin")
	}
	return 1
}
