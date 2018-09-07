/*
MIT License

Copyright (c) 2016 Sascha Hanse
Copyright (c) 2017 Shinya Yagyu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package giota

import (
	"github.com/iotaledger/giota/bundle"
	"github.com/iotaledger/giota/pow"
	"github.com/iotaledger/giota/signing"
	"github.com/iotaledger/giota/trinary"
	"math/rand"
	"testing"
)

// publicNodes is a list of known public nodes from http://iotasupport.com/lightwallet.shtml.
var (
	publicNodes = []string{
		"http://service.iotasupport.com:14265",
		"http://eugene.iota.community:14265",
		"http://eugene.iotasupport.com:14999",
		"http://eugeneoldisoft.iotasupport.com:14265",
		"http://mainnet.necropaz.com:14500",
		"http://iotatoken.nl:14265",
		"http://iota.digits.blue:14265",
		"http://wallets.iotamexico.com:80",
		"http://5.9.137.199:14265",
		"http://5.9.118.112:14265",
		"http://5.9.149.169:14265",
		"http://88.198.230.98:14265",
		"http://176.9.3.149:14265",
		"http://iota.bitfinex.com:80",
	}
)

// randomNode returns a random node from publicNodes. If local IRI exists, return
// localhost address.
func randomNode() string {
	// local
	api := NewAPI("", nil)
	_, err := api.GetNodeInfo()
	if err == nil {
		return api.endpoint
	}

	// random node
	b := make([]byte, 1)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	// return a node from the public node slice
	return publicNodes[int(b[0])%len(publicNodes)]
}

func TestAPIGetNodeInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var err error
	var resp *GetNodeInfoResponse

	for i := 0; i < 5; i++ {
		var server = randomNode()
		api := NewAPI(server, nil)
		resp, err = api.GetNodeInfo()
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("GetNodeInfo() expected err to be nil but got %v", err)
	}

	if resp.AppName == "" {
		t.Errorf("GetNodeInfo() returned invalid response: %#v", resp)
	}

}

/*
func TestAPIGetNeighbors(t *testing.T) {
	api := NewAPI(server, nil)

	_, err := api.GetNeighbors()
	if err != nil {
		t.Errorf("GetNeighbors() expected err to be nil but got %v", err)
	}
}

func TestAPIAddNeighbors(t *testing.T) {
	api := NewAPI(server, nil)

	resp, err := api.AddNeighbors([]string{"udp://127.0.0.1:14265/"})
	if err != nil {
		t.Errorf("AddNeighbors([]) expected err to be nil but got %v", err)
	} else if resp.AddedNeighbors != 1 {
		t.Errorf("AddNeighbors([]) expected to add %d got %d", 0, resp.AddedNeighbors)
	}
}

func TestAPIRemoveNeighbors(t *testing.T) {
	api := NewAPI(server, nil)

	resp, err := api.RemoveNeighbors([]string{"udp://127.0.0.1:14265/"})
	if err != nil {
		t.Errorf("RemoveNeighbors([]) expected err to be nil but got %v", err)
	} else if resp.RemovedNeighbors != 1 {
		t.Errorf("RemoveNeighbors([]) expected to remove %d got %d", 0, resp.RemovedNeighbors)
	}
}
func TestAPIGetTips(t *testing.T) {
	api := NewAPI(server, nil)

	resp, err := api.GetTips()
	if err != nil {
		t.Fatalf("GetTips() expected err to be nil but got %v", err)
	}

	if len(resp.Hashes) < 1 {
		t.Errorf("GetTips() returned less than one tip")
	}
	t.Log(len(resp.Hashes))
}
*/
func TestAPIFindTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var err error
	var resp *FindTransactionsResponse

	ftr := &FindTransactionsRequest{Bundles: []trinary.Trytes{"DEXRPLKGBROUQMKCLMRPG9HFKCACDZ9AB9HOJQWERTYWERJNOYLW9PKLOGDUPC9DLGSUH9UHSKJOASJRU"}}
	for i := 0; i < 5; i++ {
		var server = randomNode()
		api := NewAPI(server, nil)

		resp, err = api.FindTransactions(ftr)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Errorf("FindTransactions([]) expected err to be nil but got %v", err)
	}

	t.Logf("FindTransactions() = %#v", resp)
}

func TestAPIGetTrytes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var err error
	var resp *GetTrytesResponse

	for i := 0; i < 5; i++ {
		var server = randomNode()
		api := NewAPI(server, nil)

		resp, err = api.GetTrytes([]trinary.Trytes{}...)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Errorf("GetTrytes([]) expected err to be nil but got %v", err)
	}

	t.Logf("GetTrytes() = %#v", resp)
}

func TestAPIGetInclusionStates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var err error
	var resp *GetInclusionStatesResponse

	for i := 0; i < 5; i++ {
		var server = randomNode()
		api := NewAPI(server, nil)
		resp, err = api.GetInclusionStates([]trinary.Trytes{}, []trinary.Trytes{})
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Errorf("GetInclusionStates([]) expected err to be nil but got %v", err)
	}

	t.Logf("GetInclusionStates() = %#v", resp)
}

func TestAPIGetBalances(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var err error
	var resp *GetBalancesResponse

	for i := 0; i < 5; i++ {
		var server = randomNode()
		api := NewAPI(server, nil)

		resp, err = api.GetBalances([]signing.Address{}, 100)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Errorf("GetBalances([]) expected err to be nil but got %v", err)
	}

	t.Logf("GetBalances() = %#v", resp)
}

func TestAPIGetTransactionsToApprove(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var err error
	var resp *GetTransactionsToApproveResponse

	for i := 0; i < 5; i++ {
		var server = randomNode()
		api := NewAPI(server, nil)

		resp, err = api.GetTransactionsToApprove(3, "")
		if err == nil {
			break
		}
	}

	switch {
	case err != nil:
		t.Errorf("GetTransactionsToApprove() expected err to be nil but got %v", err)
	case resp.BranchTransaction == "" || resp.TrunkTransaction == "":
		t.Errorf("GetTransactionsToApprove() return empty branch and/or trunk transactions\n%#v", resp)
	}
}

func TestAPIGetLatestInclusion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var err error
	var resp []bool

	for i := 0; i < 5; i++ {
		var server = randomNode()
		api := NewAPI(server, nil)

		resp, err = api.GetLatestInclusion([]trinary.Trytes{"B9OETFYOEIUYEVB9WWCMGIHIJLFU9IJOBYYGSTZBLFBZLGZRKBIREYTIPPFGC9SPEOJFIYFRRSPX99999"})
		if err == nil && len(resp) > 0 {
			break
		}
	}

	switch {
	case err != nil:
		t.Errorf("GetLatestInclustion() expected err to be nil but got %v", err)
	case len(resp) == 0:
		t.Error("GetLatestInclustion() is invalid resp:", resp)
	}
}

func TestAPICheckConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	var server = randomNode()
	api := NewAPI(server, nil)

	resp, err := api.CheckConsistency([]trinary.Trytes{"NLNRYUTSLRQONSQEXBAJI9AIOJOEEJDOFJTETPFMB9AEEPUDIXXOTKXG9BYALEXOMSUYJEJSCZTY99999"})

	switch {
	case err != nil:
		t.Errorf("CheckConsistency() expected err to be nil but got '%v'", err)
	case resp.State != true:
		t.Error("CheckConsistency() expected true, got false")
	}
}

var (
	seed             trinary.Trytes
	skipTransferTest = false
)

func TestTransfer1(t *testing.T) {
	if skipTransferTest {
		t.Skip("transfer test skipped because a valid $TRANSFER_TEST_SEED was not specified")
	}

	var (
		err  error
		adr  signing.Address
		adrs []signing.Address
	)

	for i := 0; i < 5; i++ {
		api := NewAPI(randomNode(), nil)
		adr, adrs, err = api.GetUntilFirstUnusedAddress(seed, 2)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Error(err)
	}

	t.Log(adr, adrs)
	if len(adrs) < 1 {
		t.Error("GetUntilFirstUnusedAddress is incorrect")
	}

	var bal Balances
	for i := 0; i < 5; i++ {
		api := NewAPI(randomNode(), nil)
		bal, err = api.GetInputs(seed, 0, 10, 1000, 2)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Error(err)
	}

	t.Log(bal)
	if len(bal) < 1 {
		t.Error("GetInputs is incorrect")
	}
}

// nolint: gocyclo
func TestTransfer2(t *testing.T) {
	if skipTransferTest {
		t.Skip("transfer test skipped because a valid $TRANSFER_TEST_SEED was not specified")
	}

	var err error
	trs := []bundle.Transfer{
		{
			Address: "KTXFP9XOVMVWIXEWMOISJHMQEXMYMZCUGEQNKGUNVRPUDPRX9IR9LBASIARWNFXXESPITSLYAQMLCLVTL9QTIWOWTY",
			Value:   20,
			Tag:     "MOUDAMEPO",
		},
	}

	var bdl bundle.Bundle
	for i := 0; i < 5; i++ {
		api := NewAPI(randomNode(), nil)
		bdl, err = api.PrepareTransfers(seed, trs, nil, "", 2)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Error(err)
	}

	if len(bdl) < 3 {
		for _, tx := range bdl {
			t.Log(tx.Trytes())
		}
		t.Fatal("PrepareTransfers is incorrect len(bdl)=", len(bdl))
	}

	if err = bdl.IsValid(); err != nil {
		t.Error(err)
	}

	name, pow := pow.GetBestPoW()
	t.Log("using PoW: ", name)

	for i := 0; i < 5; i++ {
		api := NewAPI(randomNode(), nil)
		bdl, err = api.Send(seed, 2, 3, trs, 18, pow)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Error(err)
	}

	for _, tx := range bdl {
		t.Log(tx.Trytes())
	}
}

/*
func TestAPIInterruptAttachingToTangle(t *testing.T) {
	api := NewAPI(server, nil)

	err := api.InterruptAttachingToTangle()
	if err != nil {
		t.Errorf("InterruptAttachingToTangle() expected err to be nil but got %v", err)
	}
}

// XXX: The following tests are failing because I'd rather not just
//      constantly attach/broadcast/store the same transaction
func TestAPIAttachToTangle(t *testing.T) {
	api := NewAPI(server, nil)

	anr := &AttachToTangleRequest{}
	resp, err := api.AttachToTangle(anr)
	if err != nil {
		t.Errorf("AttachToTangle([]) expected err to be nil but got %v", err)
	}
	t.Logf("AttachToTangle() = %#v", resp)
}

func TestAPIBroadcastTransactions(t *testing.T) {
	api := NewAPI(server, nil)

	err := api.BroadcastTransactions([]Transaction{})
	if err != nil {
		t.Errorf("BroadcastTransactions() expected err to be nil but got %v", err)
	}
}

func TestAPIStoreTransactions(t *testing.T) {
	api := NewAPI(server, nil)

	err := api.StoreTransactions([]Trytes{})
	if err != nil {
		t.Errorf("StoreTransactions() expected err to be nil but got %v", err)
	}
}
*/
