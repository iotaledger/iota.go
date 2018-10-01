package mocks

import (
	"bytes"
	"encoding/json"
	"github.com/iotaledger/giota"
	. "github.com/iotaledger/giota/samples"
	"github.com/iotaledger/giota/transaction"
	. "github.com/iotaledger/giota/trinary"
	"net/http"
	"strings"
)

type CloseBuffer struct {
	*bytes.Buffer
}

func (cb *CloseBuffer) Close() error {
	return nil
}

var AttachToTangleCommand = giota.AttachToTangleRequest{
	TrunkTransaction:   Hash(Bundle[len(Bundle)-1].TrunkTransaction),
	BranchTransaction:  Hash(Bundle[len(Bundle)-1].BranchTransaction),
	MinWeightMagnitude: 14,
	Trytes: func() []Trytes {
		reversed := []Trytes{}
		for i := len(Bundle) - 1; i >= 0; i-- {
			reversed = append(reversed, transaction.TransactionToTrytes(&Bundle[i]))
		}
		return reversed
	}(),
}

var AttachToTangleResponseMock = giota.AttachToTangleResponse{
	Trytes: BundleTrytes,
}

type AttachToTangleMock struct {
}

func (c AttachToTangleMock) Do(req *http.Request) (*http.Response, error) {
	resBytes, _ := json.Marshal(AttachToTangleResponseMock)
	return &http.Response{
		Body:       &CloseBuffer{bytes.NewBuffer(resBytes)},
		StatusCode: http.StatusOK,
	}, nil
}

var CheckConsistencyCommand = giota.CheckConsistencyRequest{
	Tails: Hashes{
		strings.Repeat("A", 81),
		strings.Repeat("B", 81),
	},
}
