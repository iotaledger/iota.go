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

//trytes
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
)

//PowFunc is the tyoe of func for PoW
type PowFunc func(Trytes, int) (Trytes, error)

var pows = make(map[string]PowFunc)

func init() {
	pows["PowGo"] = PowGo
}

//GetBestPoW returns most preferable PoW func.
func GetBestPoW() (string, PowFunc) {
	if p, exist := pows["PowSSE"]; exist {
		return "PowSSE", p
	}
	if p, exist := pows["PowC"]; exist {
		return "PowC", p
	}
	return "PowGo", PowGo
}

func transform64(lmid *[stateSize]uint64, hmid *[stateSize]uint64) {
	var ltmp, htmp [stateSize]uint64
	lfrom := lmid
	hfrom := hmid
	lto := &ltmp
	hto := &htmp

	for r := 0; r < 26; r++ {
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

	for j := 0; j < HashSize; j++ {
		t1 := indices[j]
		t2 := indices[j+1]

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
	i := 0
	//to avoid boundry check, i believe.
	for i = 4; i < HashSize && carry != 0; i++ {
		low := lmid[i]
		high := hmid[i]
		lmid[i] = high ^ low
		hmid[i] = low
		carry = high & (^low)
	}
	return i == HashSize
}

func seri(l *[stateSize]uint64, h *[stateSize]uint64, n uint) Trits {
	r := make(Trits, HashSize)
	for i := 0; i < HashSize; i++ {
		ll := (l[i] >> n) & 1
		hh := (h[i] >> n) & 1
		if hh == 0 && ll == 1 {
			r[i] = -1
		}
		if hh == 1 && ll == 1 {
			r[i] = 0
		}
		if hh == 1 && ll == 0 {
			r[i] = 1
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

func loop(lmid *[stateSize]uint64, hmid *[stateSize]uint64, m int) (Trits, int) {
	var lcpy, hcpy [stateSize]uint64
	i := 0
	for i = 0; !incr(lmid, hmid); i++ {
		copy(lcpy[:], lmid[:])
		copy(hcpy[:], hmid[:])
		transform64(&lcpy, &hcpy)
		if n := check(&lcpy, &hcpy, m); n >= 0 {
			nonce := seri(lmid, hmid, uint(n))
			return nonce, i * 64
		}
	}
	return nil, -i * 64
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

//PowGo is proof of work of iota in pure Go.
func PowGo(trytes Trytes, mwm int) (Trytes, error) {
	c := NewCurl()
	c.Absorb(trytes[:(transactionTrinarySize-HashSize)/3])

	lmid, hmid := para(c.state)
	lmid[0] = low0
	hmid[0] = high0
	lmid[1] = low1
	hmid[1] = high1
	lmid[2] = low2
	hmid[2] = high2
	lmid[3] = low3
	hmid[3] = high3

	nonce, _ := loop(lmid, hmid, mwm)
	return nonce.Trytes(), nil
}
