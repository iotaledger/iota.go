package mam_test

import (
	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/trinary"
)

type fakeAPI struct {
	calls chan *call
}

func newFakeAPI() *fakeAPI {
	return &fakeAPI{calls: make(chan *call)}
}

func (f *fakeAPI) PrepareTransfers(seed trinary.Trytes, transfers bundle.Transfers, opts api.PrepareTransfersOptions) ([]trinary.Trytes, error) {
	call := newCall("PrepareTransfers", seed, transfers, opts)
	f.calls <- call
	returns := <-call.returnChan
	return returns[0].([]trinary.Trytes), err(returns)
}

func (f *fakeAPI) SendTrytes(trytes []trinary.Trytes, depth uint64, mwm uint64, reference ...trinary.Hash) (bundle.Bundle, error) {
	call := newCall("SendTrytes", trytes, depth, mwm, reference)
	f.calls <- call
	returns := <-call.returnChan
	return returns[0].(bundle.Bundle), err(returns)
}

type call struct {
	method     string
	arguments  []interface{}
	returnChan chan []interface{}
}

func newCall(method string, arguments ...interface{}) *call {
	return &call{
		method:     method,
		arguments:  arguments,
		returnChan: make(chan []interface{}),
	}
}

func (c *call) returns(values ...interface{}) {
	c.returnChan <- values
}

func err(returns []interface{}) error {
	if len(returns) == 0 {
		return nil
	}
	last := returns[len(returns)-1]
	if last == nil {
		return nil
	}
	if e, ok := last.(error); ok {
		return e
	}
	return nil
}
