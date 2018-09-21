package bundle

import (
	"github.com/iotaledger/giota/signing"
	"github.com/iotaledger/giota/trinary"
	"testing"
	"time"
)

type tx struct {
	addr      signing.Address
	value     int64
	timestamp string
}

func TestBundle(t *testing.T) {
	tests := []struct {
		name         string
		transactions []tx
		hash         trinary.Trytes
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

			bs.AddEntry(1, tx.addr, tx.value, parsedTime, "")
		}

		bundleHash, err := bs.Hash()
		if err != nil {
			t.Fatal(err)
		}
		if bundleHash != tt.hash {
			t.Errorf("%s: hash of bundles is illegal: %s", tt.name, bundleHash)
		}

		bs.Finalize([]trinary.Trytes{})

		send, receive := bs.Categorize(tt.transactions[1].addr)
		if len(send) != 1 || len(receive) != 1 {
			t.Errorf("%s: Categorize is incorrect", tt.name)
		}
	}

}
