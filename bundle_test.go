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

func TestBundle(t *testing.T) {
	var bs Bundle
	adr := []Address{
		"PQTDJXXKSNYZGRJDXEHHMNCLUVOIRZC9VXYLSITYMVCQDQERAHAUZJKRNBQEUHOLEAXRUSQBNYVJWESYR",
		"KTXFP9XOVMVWIXEWMOISJHMQEXMYMZCUGEQNKGUNVRPUDPRX9IR9LBASIARWNFXXESPITSLYAQMLCLVTL",
		"KTXFP9XOVMVWIXEWMOISJHMQEXMYMZCUGEQNKGUNVRPUDPRX9IR9LBASIARWNFXXESPITSLYAQMLCLVTL",
		"GXZWHBLRGGY9BCWCAVTFGHCOEWDBFLBTVTIBOQICKNLCCZIPYGPESAPUPDNBDQYENNMJTWSWDHZTYEHAJ",
	}

	value := []int64{
		50, -100, 0, 50,
	}

	ts := []string{
		"2017-03-11 12:25:05 +0900 JST",
		"2017-03-11 12:25:18 +0900 JST",
		"2017-03-11 12:25:18 +0900 JST",
		"2017-03-11 12:25:28 +0900 JST",
	}

	var hash Trytes = "ERWNDFZINIYEJQGLNFEZOU9FBHQLZOINIWJVLQ9UONHGRPSSYX9E9KQZMWCULVDNDUSUDSDMVVOICKTSY"
	for i := 0; i < 4; i++ {
		tss, err := time.Parse("2006-01-02 15:04:05 -0700 MST", ts[i])

		if err != nil {
			t.Fatal(err)
		}

		bs.Add(1, adr[i], value[i], tss, "")
	}

	if bs.Hash() != hash {
		t.Error("hash of bundles is illegal.", bs.Hash())
	}

	bs.Finalize([]Trytes{})

	send, receive := bs.Categorize(adr[1])
	if len(send) != 1 || len(receive) != 1 {
		t.Error("Categorize is incorrect")
	}
}
