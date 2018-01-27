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

import (
	"testing"
	"time"
)

type tx struct {
	addr      Address
	value     int64
	timestamp string
}

func TestBundle(t *testing.T) {
	tests := []struct {
		name         string
		transactions []tx
		hash         Trytes
	}{
		{
			name: "test transaction bundle validates correctly",
			transactions: []tx{
				tx{
					addr:      "PQTDJXXKSNYZGRJDXEHHMNCLUVOIRZC9VXYLSITYMVCQDQERAHAUZJKRNBQEUHOLEAXRUSQBNYVJWESYR",
					value:     50,
					timestamp: "2017-03-11 12:25:05 +0900 JST",
				},
				tx{
					addr:      "KTXFP9XOVMVWIXEWMOISJHMQEXMYMZCUGEQNKGUNVRPUDPRX9IR9LBASIARWNFXXESPITSLYAQMLCLVTL",
					value:     -100,
					timestamp: "2017-03-11 12:25:18 +0900 JST",
				},
				tx{
					addr:      "KTXFP9XOVMVWIXEWMOISJHMQEXMYMZCUGEQNKGUNVRPUDPRX9IR9LBASIARWNFXXESPITSLYAQMLCLVTL",
					value:     0,
					timestamp: "2017-03-11 12:25:18 +0900 JST",
				},
				tx{
					addr:      "GXZWHBLRGGY9BCWCAVTFGHCOEWDBFLBTVTIBOQICKNLCCZIPYGPESAPUPDNBDQYENNMJTWSWDHZTYEHAJ",
					value:     50,
					timestamp: "2017-03-11 12:25:28 +0900 JST",
				},
			},
			hash: "ERWNDFZINIYEJQGLNFEZOU9FBHQLZOINIWJVLQ9UONHGRPSSYX9E9KQZMWCULVDNDUSUDSDMVVOICKTSY",
		},
	}

	for _, tt := range tests {
		var bs Bundle

		for _, tx := range tt.transactions {
			parsedTime, err := time.Parse("2006-01-02 15:04:05 -0700 MST", tx.timestamp)
			if err != nil {
				t.Fatal(err)
			}

			bs.Add(1, tx.addr, tx.value, parsedTime, "")
		}

		if bs.Hash() != tt.hash {
			t.Errorf("%s: hash of bundles is illegal: %s", tt.name, bs.Hash())
		}

		bs.Finalize([]Trytes{})

		send, receive := bs.Categorize(tt.transactions[1].addr)
		if len(send) != 1 || len(receive) != 1 {
			t.Errorf("%s: Categorize is incorrect", tt.name)
		}
	}

}
