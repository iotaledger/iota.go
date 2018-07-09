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
	"errors"
	"fmt"
	"runtime"
	"sync"
)

// trytes
const (
	hBits uint64 = 0xFFFFFFFFFFFFFFFF
	lBits uint64 = 0x0000000000000000

	low0  uint64 = 0xDB6DB6DB6DB6DB6D
	high0 uint64 = 0xB6DB6DB6DB6DB6DB
	low1  uint64 = 0xF1F8FC7E3F1F8FC7
	high1 uint64 = 0x8FC7E3F1F8FC7E3F
	low2  uint64 = 0x7FFFE00FFFFC01FF
	high2 uint64 = 0xFFC01FFFF803FFFF
	low3  uint64 = 0xFFC0000007FFFFFF
	high3 uint64 = 0x003FFFFFFFFFFFFF

	nonceOffset         = HashSize - NonceTrinarySize
	nonceInitStart      = nonceOffset + 4
	nonceIncrementStart = nonceInitStart + NonceTrinarySize/3
)

// PowFunc is the func type for PoW
type PowFunc func(Trytes, int) (Trytes, error)

var (
	powFuncs = make(map[string]PowFunc)
	// PowProcs is number of concurrent processes (default is NumCPU()-1)
	PowProcs int
)

func init() {
	powFuncs["PowGo"] = PowGo
	PowProcs = runtime.NumCPU()
	if PowProcs != 1 {
		PowProcs--
	}
}

// GetPowFunc returns a specific PoW func
func GetPowFunc(pow string) (PowFunc, error) {
	if p, exist := powFuncs[pow]; exist {
		return p, nil
	}

	return nil, fmt.Errorf("PowFunc %v does not exist", pow)
}

// GetPowFuncNames returns an array with the names of the existing PoW methods
func GetPowFuncNames() (powFuncNames []string) {
	powFuncNames = make([]string, len(powFuncs))

	i := 0
	for k := range powFuncs {
		powFuncNames[i] = k
		i++
	}

	return powFuncNames
}

// GetBestPoW returns most preferable PoW func.
func GetBestPoW() (string, PowFunc) {

	// PowGo is the last and default return value
	powOrderPreference := []string{"PowCL", "PowSSE", "PowCARM64", "PowC128", "PowC"}

	for _, pow := range powOrderPreference {
		if p, exist := powFuncs[pow]; exist {
			return pow, p
		}
	}

	return "PowGo", PowGo // default return PowGo if no others
}

func transform64(lmid *[stateSize]uint64, hmid *[stateSize]uint64) {
	var ltmp, htmp [stateSize]uint64
	lfrom := lmid
	hfrom := hmid
	lto := &ltmp
	hto := &htmp

	for r := 0; r < numberOfRounds-1; r++ {
		for j := 0; j < stateSize; j++ {
			t1 := indices[j]
			t2 := indices[j+1]

			alpha := lfrom[t1]
			beta := hfrom[t1]
			gamma := hfrom[t2]
			delta := (alpha | (^gamma)) & (lfrom[t2] ^ beta)

			lto[j] = ^delta
			hto[j] = (alpha ^ gamma) | delta
		}

		lfrom, lto = lto, lfrom
		hfrom, hto = hto, hfrom
	}

	for j := 0; j < stateSize; j++ {
		t1, t2 := indices[j], indices[j+1]

		alpha := lfrom[t1]
		beta := hfrom[t1]
		gamma := hfrom[t2]
		delta := (alpha | (^gamma)) & (lfrom[t2] ^ beta)

		lto[j] = ^delta
		hto[j] = (alpha ^ gamma) | delta
	}

	copy(lmid[:], ltmp[:])
	copy(hmid[:], htmp[:])
}

func incr(lmid *[stateSize]uint64, hmid *[stateSize]uint64) bool {
	var carry uint64 = 1
	var i int
	//to avoid boundry check, I believe.
	for i = nonceInitStart; i < HashSize && carry != 0; i++ {
		low := lmid[i]
		high := hmid[i]
		lmid[i] = high ^ low
		hmid[i] = low
		carry = high & (^low)
	}
	return i == HashSize
}

func seri(l *[stateSize]uint64, h *[stateSize]uint64, n uint) Trits {
	r := make(Trits, NonceTrinarySize)
	for i := nonceOffset; i < HashSize; i++ {
		ll := (l[i] >> n) & 1
		hh := (h[i] >> n) & 1

		switch {
		case hh == 0 && ll == 1:
			r[i-nonceOffset] = -1
		case hh == 1 && ll == 1:
			r[i-nonceOffset] = 0
		case hh == 1 && ll == 0:
			r[i-nonceOffset] = 1
		}
	}
	return r
}

func check(l *[stateSize]uint64, h *[stateSize]uint64, m int) int {
	nonceProbe := hBits
	for i := HashSize - m; i < HashSize; i++ {
		nonceProbe &= ^(l[i] ^ h[i])
		if nonceProbe == 0 {
			return -1
		}
	}

	var i uint
	for i = 0; i < 64; i++ {
		if (nonceProbe>>i)&1 == 1 {
			return int(i)
		}
	}
	return -1
}

var stopGO = true

func loop(lmid *[stateSize]uint64, hmid *[stateSize]uint64, m int) (Trits, int64) {
	var lcpy, hcpy [stateSize]uint64
	var i int64
	for i = 0; !incr(lmid, hmid) && !stopGO; i++ {
		copy(lcpy[:], lmid[:])
		copy(hcpy[:], hmid[:])
		transform64(&lcpy, &hcpy)

		if n := check(&lcpy, &hcpy, m); n >= 0 {
			nonce := seri(lmid, hmid, uint(n))
			return nonce, i * 64
		}
	}
	return nil, i * 64
}

// 01:-1 11:0 10:1
func para(in Trits) (*[stateSize]uint64, *[stateSize]uint64) {
	var l, h [stateSize]uint64

	for i := 0; i < stateSize; i++ {
		switch in[i] {
		case 0:
			l[i] = hBits
			h[i] = hBits
		case 1:
			l[i] = lBits
			h[i] = hBits
		case -1:
			l[i] = hBits
			h[i] = lBits
		}
	}
	return &l, &h
}

func incrN(n int, lmid *[stateSize]uint64, hmid *[stateSize]uint64) {
	for j := 0; j < n; j++ {
		var carry uint64 = 1

		// to avoid boundry check, i believe.
		for i := nonceInitStart; i < nonceIncrementStart && carry != 0; i++ {
			low := lmid[i]
			high := hmid[i]
			lmid[i] = high ^ low
			hmid[i] = low
			carry = high & (^low)
		}
	}
}

var countGo int64 = 1

// PowGo is proof of work for iota in pure Go
func PowGo(trytes Trytes, mwm int) (Trytes, error) {
	if !stopGO {
		stopGO = true
		return "", errors.New("pow is already running, stopped")
	}

	if trytes == "" {
		return "", errors.New("invalid trytes")
	}

	countGo = 0
	stopGO = false

	c := NewCurl()
	c.Absorb(trytes[:(TransactionTrinarySize-HashSize)/3])
	tr := trytes.Trits()
	copy(c.state, tr[TransactionTrinarySize-HashSize:])

	var (
		result Trytes
		wg     sync.WaitGroup
		mutex  sync.Mutex
	)

	for i := 0; i < PowProcs; i++ {
		wg.Add(1)
		go func(i int) {
			lmid, hmid := para(c.state)
			lmid[nonceOffset] = low0
			hmid[nonceOffset] = high0
			lmid[nonceOffset+1] = low1
			hmid[nonceOffset+1] = high1
			lmid[nonceOffset+2] = low2
			hmid[nonceOffset+2] = high2
			lmid[nonceOffset+3] = low3
			hmid[nonceOffset+3] = high3

			incrN(i, lmid, hmid)
			nonce, cnt := loop(lmid, hmid, mwm)

			mutex.Lock()
			if nonce != nil {
				result = nonce.Trytes()
				stopGO = true
			}

			countGo += cnt
			mutex.Unlock()
			wg.Done()
		}(i)
	}

	wg.Wait()
	stopGO = true
	return result, nil
}
