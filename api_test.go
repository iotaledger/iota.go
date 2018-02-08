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

import "testing"

func TestAPIGetNodeInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	var err error
	var resp *GetNodeInfoResponse

	for i := 0; i < 5; i++ {
		var server = RandomNode()
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

	ftr := &FindTransactionsRequest{Bundles: []Trytes{"DEXRPLKGBROUQMKCLMRPG9HFKCACDZ9AB9HOJQWERTYWERJNOYLW9PKLOGDUPC9DLGSUH9UHSKJOASJRU"}}
	for i := 0; i < 5; i++ {
		var server = RandomNode()
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
		var server = RandomNode()
		api := NewAPI(server, nil)

		resp, err = api.GetTrytes([]Trytes{})
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
		var server = RandomNode()
		api := NewAPI(server, nil)
		resp, err = api.GetInclusionStates([]Trytes{}, []Trytes{})
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
		var server = RandomNode()
		api := NewAPI(server, nil)

		resp, err = api.GetBalances([]Address{}, 100)
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
		var server = RandomNode()
		api := NewAPI(server, nil)

		resp, err = api.GetTransactionsToApprove(Depth, NumberOfWalks, "")
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
		var server = RandomNode()
		api := NewAPI(server, nil)

		resp, err = api.GetLatestInclusion([]Trytes{"B9OETFYOEIUYEVB9WWCMGIHIJLFU9IJOBYYGSTZBLFBZLGZRKBIREYTIPPFGC9SPEOJFIYFRRSPX99999"})
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
