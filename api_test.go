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

const server = "http://service.iotasupport.com:14265"

func TestAPIGetNodeInfo(t *testing.T) {
	api := NewAPI(server, nil)

	resp, err := api.GetNodeInfo()
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
	if err != nil {
		t.Fatalf("GetNodeInfo() expected err to be nil but got %v", err)
	}

	_, err = api.GetNeighbors()
	if err != nil {
		t.Errorf("GetNeighbors() expected err to be nil but got %v", err)
	}
}

func TestAPIAddNeighbors(t *testing.T) {
	api := NewAPI(server, nil)
	if err != nil {
		t.Fatalf("GetNodeInfo() expected err to be nil but got %v", err)
	}

	anr := &AddNeighborsRequest{URIS: []string{"udp://127.0.0.1:14265/"}}
	resp, err := api.AddNeighbors(anr)
	if err != nil {
		t.Errorf("AddNeighbors([]) expected err to be nil but got %v", err)
	} else if resp.AddedNeighbors != 1 {
		t.Errorf("AddNeighbors([]) expected to add %d got %d", 0, resp.AddedNeighbors)
	}
}

func TestAPIRemoveNeighbors(t *testing.T) {
	api := NewAPI(server, nil)
	if err != nil {
		t.Fatalf("GetNodeInfo() expected err to be nil but got %v", err)
	}

	anr := &RemoveNeighborsRequest{URIS: []string{"udp://127.0.0.1:14265/"}}
	resp, err := api.RemoveNeighbors(anr)
	if err != nil {
		t.Errorf("RemoveNeighbors([]) expected err to be nil but got %v", err)
	} else if resp.RemovedNeighbors != 1 {
		t.Errorf("RemoveNeighbors([]) expected to remove %d got %d", 0, resp.RemovedNeighbors)
	}
}
func TestAPIGetTips(t *testing.T) {
	api := NewAPI(server, nil)
	if err != nil {
		t.Fatalf("NewAPI(empty, nil) expected err to be nil but got %v", err)
	}

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
	api := NewAPI(server, nil)

	ftr := &FindTransactionsRequest{Bundles: []Trytes{"DEXRPLKGBROUQMKCLMRPG9HFKCACDZ9AB9HOJQWERTYWERJNOYLW9PKLOGDUPC9DLGSUH9UHSKJOASJRU"}}
	resp, err := api.FindTransactions(ftr)
	if err != nil {
		t.Errorf("FindTransactions([]) expected err to be nil but got %v", err)
	}
	t.Logf("FindTransactions() = %#v", resp)
}

func TestAPIGetTrytes(t *testing.T) {
	api := NewAPI(server, nil)

	anr := &GetTrytesRequest{Hashes: []Trytes{}}
	resp, err := api.GetTrytes(anr)
	if err != nil {
		t.Errorf("GetTrytes([]) expected err to be nil but got %v", err)
	}
	t.Logf("GetTrytes() = %#v", resp)
}

func TestAPIGetInclusionStates(t *testing.T) {
	api := NewAPI(server, nil)

	anr := &GetInclusionStatesRequest{Transactions: []Trytes{}, Tips: []string{}}
	resp, err := api.GetInclusionStates(anr)
	if err != nil {
		t.Errorf("GetInclusionStates([]) expected err to be nil but got %v", err)
	}
	t.Logf("GetInclusionStates() = %#v", resp)
}

func TestAPIGetBalances(t *testing.T) {
	api := NewAPI(server, nil)

	gbr := &GetBalancesRequest{Addresses: []Address{}, Threshold: 100}
	resp, err := api.GetBalances(gbr)
	if err != nil {
		t.Errorf("GetBalances([]) expected err to be nil but got %v", err)
	}
	t.Logf("GetBalances() = %#v", resp)
}

func TestAPIGetTransactionsToApprove(t *testing.T) {
	api := NewAPI(server, nil)

	anr := &GetTransactionsToApproveRequest{}
	resp, err := api.GetTransactionsToApprove(anr)
	if err != nil {
		t.Errorf("GetTransactionsToApprove() expected err to be nil but got %v", err)
	} else if resp.BranchTransaction == "" || resp.TrunkTransaction == "" {
		t.Errorf("GetTransactionsToApprove() return empty branch and/or trunk transactions\n%#v", resp)
	}
}

/*
func TestAPIInterruptAttachingToTangle(t *testing.T) {
	api := NewAPI(server, nil)

	resp, err := api.InterruptAttachingToTangle()
	if err != nil {
		t.Errorf("InterruptAttachingToTangle() expected err to be nil but got %v", err)
	}
	t.Logf("InterruptAttachingToTangle() = %#v", resp)
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


	anr := &BroadcastTransactionsRequest{}
	resp, err := api.BroadcastTransactions(anr)
	if err != nil {
		t.Errorf("BroadcastTransactions() expected err to be nil but got %v", err)
	}
	t.Logf("BroadcastTransactions() = %#v", resp)
}

func TestAPIStoreTransactions(t *testing.T) {
	api := NewAPI(server, nil)


	anr := &StoreTransactionsRequest{}
	resp, err := api.StoreTransactions(anr)
	if err != nil {
		t.Errorf("StoreTransactions() expected err to be nil but got %v", err)
	}
	t.Logf("StoreTransactions() = %#v", resp)
}
*/
/*
func TestAAA(t *testing.T) {
	api := NewAPI(server, nil)


	anr := &GetTrytesRequest{Hashes: []Trytes{"9MMRRSLICRITOROFC9FBVWLFEDNN9KJKYHUMRCJEUDGCYCWTBP9HHBEEJRFAU9FALRJWTU99NZK999999",
		"NCTWMMQWMKOGYDROQNNJKO9ALHELEHVGKCNPNWYMKXFBPPRYOAM9CHBNAHMYREVUFIPNPWWCWYP999999",
		"MRYSIXABICSX9XQSLPAPQHGAPCMBDQZXH9EOHPLL9LFQNUDTETNQFUJO9DPHTNPJI9BTQH9RM9I999999",
		"WGQXKLKYVYYELRSGPKDRXAXEKYOTXHWZGMSWJEKGTZRO9OMRDERN9BLC9ADGOPGCBPPKJKEPPYR999999",
	}}
	resp, err := api.GetTrytes(anr)
	if err != nil {
		t.Errorf("GetTrytes([]) expected err to be nil but got %v", err)
	}
	for i := range resp.Trytes {
		tx, errr := NewTransaction(resp.Trytes[i].Trits())
		if errr != nil {
			t.Error(err)
		}
		fmt.Print(tx, tx.Timestamp.Unix(), "\n\n")
	}
	ftr := &FindTransactionsRequest{Bundles: &[]Trytes{"DEXRPLKGBROUQMKCLMRPG9HFKCACDZ9AB9HOJQWERTYWERJNOYLW9PKLOGDUPC9DLGSUH9UHSKJOASJRU"}}
	respp, err := api.FindTransactions(ftr)
	if err != nil {
		t.Errorf("FindTransactions([]) expected err to be nil but got %v", err)
	}
	fmt.Println(respp.Hashes)
}
*/
/*
func TestBBB(t *testing.T) {
	api := NewAPI(server, nil)


	anr := &GetTransactionsToApproveRequest{Depth: 1000}
	resp, err := api.GetTransactionsToApprove(anr)
	if err != nil {
		t.Fatalf("GetTips() expected err to be nil but got %v", err)
	}

	anrr := &GetTrytesRequest{Hashes: []Trytes{resp.TrunkTransaction}}
	respp, err := api.GetTrytes(anrr)
	t.Log(respp.Trytes[0])
}
*/
