package giota

import (
	"time"
)

var (
	emptySig  Trytes
	emptyHash Trytes
)

func init() {
	for i := 0; i < signatureMessageFragmentTrinarySize/3; i++ {
		emptySig += "9"
	}
	for i := 0; i < HashSize; i++ {
		emptyHash += "9"
	}
}

//Bundle is one menber of  a group in a transaction.
type Bundle struct {
	Address   Trytes
	Value     int64
	Tag       Trytes
	Timestamp time.Time
}

//Bundles represents one group of bundles.
type Bundles []Bundle

//Add adds one bundle to bundle slice.
func (bs Bundles) Add(address Trytes, value int64, timestamp int64, tag Trytes, index int) {
	if len(bs) != 0 {
		value = 0
	}
	b := Bundle{
		Address:   address,
		Value:     value,
		Tag:       tag,
		Timestamp: time.Unix(timestamp, 0),
	}
	bs = append(bs, b)
}

//Txs converts bundles to transactions.
func (bs Bundles) Txs(sig []Trytes) []Transaction {
	tx := make([]Transaction, len(bs))
	for i, b := range bs {
		tx[i].Address = b.Address
		tx[i].Value = b.Value
		tx[i].Timestamp = b.Timestamp
		tx[i].Tag = b.Tag
		if sig[i] != "" {
			tx[i].SignatureMessageFragment = sig[i]
		} else {
			tx[i].SignatureMessageFragment = emptySig
		}
		tx[i].TrunkTransaction = emptyHash
		tx[i].BranchTransaction = emptyHash
		tx[i].Nonce = emptyHash
		tx[i].CurrentIndex = int64(i)
		tx[i].LastIndex = int64(len(bs) - 1)
		//Bundle is not assigned.
	}
	return tx
}

//Hash calculates hash of bundles.
func (bs Bundles) Hash() Trytes {
	c := NewCurl()
	for i, b := range bs {
		copy(c.State, Int2Trits(b.Value, 81))
		copy(c.State[81:], Int2Trits(b.Timestamp.Unix(), 27))
		copy(c.State[81+27:], Int2Trits(int64(i), 27))            //CurrentIndex
		copy(c.State[81+27+27:], Int2Trits(int64(len(bs)-1), 27)) //LastIndex
		c.Transform()
	}
	return c.Squeeze().Trytes()
}
