// Copyright © 2014,2015 Lawrence E. Bakst. All rights reserved.

package hashtable

import (
	"fmt"
	"github.com/tildeleb/cuckoo/primes"
	"leb.io/hashland/hashf"
	"leb.io/hrff"
	"time"
)

type Bucket struct {
	Key []byte
}

type Stats struct {
	Inserts      int // number of elements inserted
	Cols         int // number of collisions
	Probes       int // number of probes
	Heads        int // number of chains > 1
	Dups         int // number of dup hashes on the same chain
	Dups2        int // number of dup hashes
	Nbuckets     int // number of new buckets added
	Entries      int
	LongestChain int // longest chain of entries
	Q            float64
	Dur          time.Duration
	//
	Lines    int
	Size     uint64
	SizeLog2 uint64
	SizeMask uint64
}

type TBS struct {
	t int
	b int
	s int
}

type HashTable struct {
	Buckets [][]Bucket
	Stats
	tbs   []TBS
	Seed  uint64
	Tcnt  int // trace counter
	extra int
	pd    bool
	oa    bool
	prime bool
}

var Trace bool

// Henry Warren, "Hacker's Delight", ch. 5.3
func NextLog2(x uint32) uint32 {
	if x <= 1 {
		return x
	}
	x--
	n := uint32(0)
	y := uint32(0)
	y = x >> 16
	if y != 0 {
		n += 16
		x = y
	}
	y = x >> 8
	if y != 0 {
		n += 8
		x = y
	}
	y = x >> 4
	if y != 0 {
		n += 4
		x = y
	}
	y = x >> 2
	if y != 0 {
		n += 2
		x = y
	}
	y = x >> 1
	if y != 0 {
		return n + 2
	}
	return n + x
}

func NewHashTable(size int, seed int64, extra int, pd, oa, prime bool) *HashTable {
	ht := new(HashTable)
	ht.Lines = size
	ht.extra = extra
	if size < 0 {
		ht.Size = uint64(-size)
	} else {
		ht.SizeLog2 = uint64(NextLog2(uint32(ht.Lines)) + uint32(extra))
		ht.Size = 1 << ht.SizeLog2
	}
	ht.prime = prime
	if prime {
		ht.Size = uint64(primes.NextPrime(int(ht.Size)))
	}
	ht.Seed = uint64(seed)
	ht.pd = pd
	ht.oa = oa
	ht.SizeMask = ht.Size - 1
	ht.Buckets = make([][]Bucket, ht.Size, ht.Size)

	//ht.Nbuckets = int(ht.Size)
	return ht
}

func btoi(b []byte) int {
	return int(b[3])<<24 | int(b[2])<<16 | int(b[1])<<8 | int(b[0])
}

var ph uint64

var Ones = [16]int{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4}
var Zeros = [16]int{4, 3, 3, 2, 3, 2, 2, 1, 3, 2, 2, 1, 2, 1, 1, 0}

func diff(d uint64) (zeros int, ones int) {
	for i := 0; i < 16; i++ {
		four := d & 0xF
		zeros += Zeros[four]
		ones += Ones[four]
		d = d >> 4
	}
	if zeros+ones != 64 {
		panic("diff")
	}
	return
}

func (ht *HashTable) Insert(ka []byte) {
	k := make([]byte, len(ka), len(ka))
	k = k[:]
	amt := copy(k, ka)
	if amt != len(ka) {
		panic("Add")
	}
	ht.Inserts++
	idx := uint64(0)
	h := hashf.Hashf(k, ht.Seed) // jenkins.Hash232(k, 0)
	if ht.prime {
		idx = h % ht.Size
	} else {
		idx = h & ht.SizeMask
	}
	//fmt.Printf("index=%d\n", idx)
	cnt := 0
	pass := 0

	//fmt.Printf("Add: %x\n", k)
	//ht.Buckets[idx].Key = k
	//len(ht.Buckets[idx].Key) == 0
	for {
		if ht.Buckets[idx] == nil {
			// no entry or chain at this location, make it
			ht.Buckets[idx] = append(ht.Buckets[idx], Bucket{Key: k})
			//fmt.Printf("Insert: ins idx=%d, len=%d, hash=%#016x, key=%q\n", idx, len(ht.Buckets[idx]), h, ht.Buckets[idx][0].Key)
			//z, o := diff(h)
			//fmt.Printf("%02d %02d %#064b z=%02d o=%02d", btoi(k), idx, h, z, o)
			if Trace {
				fmt.Printf("{%q: %d, %q: %d, %q: %q, %q: %d, %q: %d, %q: %d, %q: %v, %q: %v},\n",
					"i", ht.Tcnt, "l", cnt, "op", "I", "t", 0, "b", idx, "s", 0, "k", btoi(k), "v", btoi(k))
				ht.Tcnt++
				//fmt.Printf("len(ht.tbs)=%d\n", ht.tbs)
				/*
					for _, _ := range ht.tbs {
								fmt.Printf("{%q: %d, %q: %d, %q: %q, %q: %d, %q: %d, %q: %d, %q: %v, %q: %v},\n",
									"i", ht.Tcnt, "l", cnt, "op", "U", "t", tbs.t, "b", tbs.b, "s", tbs.s, "k", btoi(k), "v", btoi(k))
							ht.Tcnt++
					}
				*/
				ht.tbs = nil
			}
			ht.Probes++
			ht.Heads++
			break
		}
		if ht.oa {
			//fmt.Printf("Insert: col idx=%d, len=%d, hash=0x%08x, key=%q\n", idx, len(ht.Buckets[idx]), h, ht.Buckets[idx][0].Key)
			if cnt == 0 {
				ht.Probes++
			} else {
				ht.Cols++
			}
			if Trace {
				/*
					ht.tbs = append(ht.tbs, TBS{0, int(idx), 0})
				*/
				fmt.Printf("{%q: %d, %q: %d, %q: %q, %q: %d, %q: %d, %q: %d, %q: %v, %q: %v},\n",
					"i", ht.Tcnt, "l", cnt, "op", "P", "t", 0, "b", idx, "s", 0, "k", btoi(k), "v", btoi(k))
				ht.Tcnt++
			}
			// check for a duplicate key
			bh := hashf.Hashf(ht.Buckets[idx][0].Key, ht.Seed)
			if bh == h {
				if ht.pd {
					fmt.Printf("hash=0x%08x, idx=%d, key=%q\n", h, idx, k)
					fmt.Printf("hash=0x%08x, idx=%d, key=%q\n", bh, idx, ht.Buckets[idx][0].Key)
				}
				ht.Dups++
			}
			idx++
			cnt++
			if idx > ht.Size-1 {
				pass++
				if pass > 1 {
					panic("Add: pass")
				}
				idx = 0
			}
		} else {
			// first scan slice for dups
			for j := range ht.Buckets[idx] {
				bh := hashf.Hashf(ht.Buckets[idx][j].Key, ht.Seed)
				//fmt.Printf("idx=%d, j=%d/%d, bh=%#016x, h=%#016x, key=%q\n", idx, j, len(ht.Buckets[idx]), bh, h, ht.Buckets[idx][j].Key)
				if bh == h {
					if ht.pd {
						//fmt.Printf("idx=%d, j=%d/%d, bh=0x%08x, h=0x%08x, key=%q, bkey=%q\n", idx, j, len(ht.Buckets[idx]), bh, h, k, ht.Buckets[idx][j].Key)
						//fmt.Printf("hash=0x%08x, idx=%d, key=%q\n", h, idx, k)
						//fmt.Printf("hash=0x%08x, idx=%d, key=%q\n", bh, idx, ht.Buckets[idx][0].Key)
					}
					ht.Dups++
				}
			}
			// add element
			ht.Buckets[idx] = append(ht.Buckets[idx], Bucket{Key: k})
			if len(ht.Buckets[idx]) > ht.LongestChain {
				//fmt.Printf("len(ht.Buckets[idx])=%d, ht.LongestChain=%d\n", len(ht.Buckets[idx]), ht.LongestChain)
				ht.LongestChain = len(ht.Buckets[idx])
			}
			//z, o := diff(h)
			//fmt.Printf("%02d %02d %#064b z=%02d o=%02d", btoi(k), idx, h, z, o)
			if Trace {
				fmt.Printf("{%q: %d, %q: %d, %q: %q, %q: %d, %q: %d, %q: %d, %q: %v, %q: %v},\n",
					"i", ht.Tcnt, "l", cnt, "op", "I", "t", len(ht.Buckets[idx])-1, "b", idx, "s", 0, "k", btoi(k), "v", btoi(k))
				ht.Tcnt++
			}
			ht.Nbuckets++
			ht.Probes++
			break
		}
	}
	if ph != 0 {
		//xor := h ^ ph
		//z, o := diff(xor)
		//fmt.Printf(" xor=%#064b z=%02d o=%02d\n", xor, z, o)
	} else {
		//fmt.Printf("\n")
	}
	ph = h
}

// The theoretical metric from "Red Dragon Book"
// appears to be useless
func (ht *HashTable) HashQuality() float64 {
	if ht.Inserts == 0 {
		return 0.0
	}
	n := float64(0.0)
	buckets := 0
	entries := 0
	for _, v := range ht.Buckets {
		if v != nil {
			buckets++
			count := float64(len(v))
			entries += len(v)
			n += count * (count + 1.0)
		}
	}
	n *= float64(ht.Size)
	d := float64(ht.Inserts) * (float64(ht.Inserts) + 2.0*float64(ht.Size) - 1.0) // (n / 2m) * (n + 2m - 1)
	//fmt.Printf("buckets=%d, entries=%d, inserts=%d, size=%d, n=%f, d=%f, n/d=%f\n", buckets, entries, ht.Inserts, ht.Size, n, d, n/d)
	//ht.Nbuckets = buckets
	//ht.Entries = entries
	ht.Q = n / d
	return n / d
}

func (s *HashTable) Print() {
	var cvt = func(t float64) (ret float64, unit string) {
		unit = "s"
		if t < 1.0 {
			unit = "ms"
			t *= 1000.0
			if t < 1.0 {
				unit = "us"
				t *= 1000.0
			}
		}
		ret = t
		return
	}

	q := s.HashQuality()
	t, units := cvt(s.Dur.Seconds())
	if s.oa {
		/*
			if test.name != "TestI" && test.name != "TestJ" && (s.Lines != s.Inserts || s.Lines != s.Heads || s.Lines != s.Nbuckets || s.Lines != s.Entries) {
				panic("runTestsWithFileAndHashes")
			}
		*/
		fmt.Printf("size=%h, inserts=%04.2h, probes=%04.2h, cols=%04.2h, cpi=%0.2f%%, ppi=%04.2f, dups=%d, time=%0.2f%s\n",
			hrff.Int64{int64(s.Size), ""}, hrff.Float64{float64(s.Inserts), ""}, hrff.Float64{float64(s.Probes), ""}, hrff.Float64{float64(s.Cols), ""},
			float64(s.Cols)/float64(s.Inserts)*100.0, float64(s.Probes)/float64(s.Inserts), s.Dups, t, units)
	} else {
		/*
			if test.name != "TestI" && test.name != "TestJ" && (s.Lines != s.Inserts || s.Lines != s.Probes || s.Lines != s.Entries) {
				fmt.Printf("lines=%d, inserts=%d, probes=%d, entries=%d\n", s.Lines, s.Inserts, s.Probes, s.Entries)
				panic("runTestsWithFileAndHashes")
			}
		*/
		//fmt.Printf("%#v\n", s)
		fmt.Printf("size=%h, inserts=%h, heads=%h, newBuckets=%h, LongestChain=%h, dups=%d, dups2=%d, q=%0.2f, time=%0.2f%s\n",
			hrff.Int64{int64(s.Size), ""}, hrff.Int64{int64(s.Inserts), ""}, hrff.Int64{int64(s.Heads), ""},
			hrff.Int64{int64(s.Nbuckets), ""}, hrff.Int64{int64(s.LongestChain), ""}, s.Dups, s.Dups2, q, t, units)
	}
}
