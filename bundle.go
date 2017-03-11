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
	"time"
)

//Bundle is one menber of  a group in a transaction.
type Bundle struct {
	Address   Address
	Value     int64
	Tag       Trytes
	Timestamp time.Time
}

//Bundles represents one group of bundles.
type Bundles []Bundle

//Add adds one bundle to bundle slice.
func (bs *Bundles) Add(num int, address Address, value int64, timestamp *time.Time, tag Trytes) {
	if tag == "" {
		tag = EmptyHash
	}
	if timestamp == nil {
		t := time.Now()
		timestamp = &t
	}
	for i := 0; i < num; i++ {
		var v int64
		if i == 0 {
			v = value
		}
		b := Bundle{
			Address:   address,
			Value:     v,
			Tag:       tag,
			Timestamp: *timestamp,
		}
		*bs = append(*bs, b)
	}
}

//Txs converts bundles to transactions.
func (bs Bundles) Txs(sig []Trytes) []Transaction {
	tx := make([]Transaction, len(bs))
	h := bs.Hash()
	for i, b := range bs {
		tx[i].Address = b.Address
		tx[i].Value = b.Value
		tx[i].Timestamp = b.Timestamp
		tx[i].Tag = b.Tag
		if len(sig) > i && sig[i] != "" {
			tx[i].SignatureMessageFragment = sig[i]
		} else {
			tx[i].SignatureMessageFragment = EmptySig
		}
		tx[i].TrunkTransaction = EmptyHash
		tx[i].BranchTransaction = EmptyHash
		tx[i].Nonce = EmptyHash
		tx[i].CurrentIndex = int64(i)
		tx[i].LastIndex = int64(len(bs) - 1)
		tx[i].Bundle = h
	}
	return tx
}

//Hash calculates hash of bundles.
func (bs Bundles) Hash() Trytes {
	c := NewCurl()
	buf := make(Trits, 243+81*3)
	for i, b := range bs {
		copy(buf, b.Address.WithoutChecksum().Trits())
		copy(buf[243:], Int2Trits(b.Value, 81))
		copy(buf[243+81:], b.Tag.Trits())
		copy(buf[243+81+81:], Int2Trits(b.Timestamp.Unix(), 27))
		copy(buf[243+81+81+27:], Int2Trits(int64(i), 27))            //CurrentIndex
		copy(buf[243+81+81+27+27:], Int2Trits(int64(len(bs)-1), 27)) //LastIndex
		c.Absorb(buf)
	}
	return c.Squeeze().Trytes()
}
