/*
MIT License

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

var (
	seed Trytes = "WQNZOHUT99PWKEBFSKQSYNC9XHT9GEBMOSJAQDQAXPEZPJNDIUB9TSNWVMHKWICW9WVZXSMDFGISOD9FZ"
	api         = NewAPI(RandomNode(), nil)
)

func TestTransfer1(t *testing.T) {
	adr, adrs, err := GetUsedAddress(api, seed, 2)
	if err != nil {
		t.Error(err)
	}
	t.Log(adr, adrs)
	if len(adrs) < 1 {
		t.Error("GetUsedAddress is incorrect")
	}

	bal, err := GetInputs(api, seed, 0, 10, 1000, 2)
	if err != nil {
		t.Error(err)
	}
	t.Log(bal)
	if len(bal) < 1 {
		t.Error("GetInputs is incorrect")
	}

}
func TestTransfer2(t *testing.T) {
	trs := []Transfer{
		Transfer{
			Balance: Balance{
				Address: "KTXFP9XOVMVWIXEWMOISJHMQEXMYMZCUGEQNKGUNVRPUDPRX9IR9LBASIARWNFXXESPITSLYAQMLCLVTL9QTIWOWTY",
				Value:   20,
			},
			Tag: "MOUDAMEPO",
		},
	}
	bdl, err := PrepareTransfers(api, seed, trs, nil, "", 2)
	if err != nil {
		t.Error(err)
	}
	if len(bdl) < 4 {
		t.Error("PrepareTransfers is incorrect")
	}
	if err := bdl.IsValid(); err != nil {
		t.Error(err)
	}

	//	bdl, err = Send(api, seed, 2, trs, PowGo)
	bdl, err = Send(api, seed, 2, trs, Pow64)
	if err != nil {
		t.Error(err)
	}
	for _, tx := range bdl {
		t.Log(tx.Trits().Trytes())
	}
}
