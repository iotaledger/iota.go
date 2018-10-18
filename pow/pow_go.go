package pow

import "C"
import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/transaction"
	. "github.com/iotaledger/iota.go/trinary"

	"github.com/pkg/errors"
	"math"
	"runtime"
	"time"
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

	nonceOffset         = HashTrinarySize - NonceTrinarySize
	nonceInitStart      = nonceOffset + 4
	nonceIncrementStart = nonceInitStart + NonceTrinarySize/3
)

var (
	ErrPoWAlreadyRunning   = errors.New("proof of work is already running (at most one can run at the same time)")
	ErrInvalidTrytesForPoW = errors.New("invalid trytes supplied to pow func")
	ErrUnknownPoWFunc      = errors.New("unknown pow func")
)

// PowFunc is the func type for PoW
type PowFunc = func(Trytes, int) (Trytes, error)

var (
	powFuncs = make(map[string]PowFunc)
	// PowProcs is number of concurrent processes (default is NumCPU()-1)
	PowProcs int
)

func init() {
	powFuncs["PoWGo"] = PoWGo
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

	return nil, errors.Wrapf(ErrUnknownPoWFunc, "%s", pow)
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

	// PoWGo is the last and default return value
	powOrderPreference := []string{"PoWCL", "PoWAVX", "PoWSSE", "PoWCARM64", "PoWC128", "PoWC"}

	for _, pow := range powOrderPreference {
		if p, exist := powFuncs[pow]; exist {
			return pow, p
		}
	}

	return "PoWGo", PoWGo // default return PoWGo if no others
}

func transform64(lmid *[curl.StateSize]uint64, hmid *[curl.StateSize]uint64) {
	var ltmp, htmp [curl.StateSize]uint64
	lfrom := lmid
	hfrom := hmid
	lto := &ltmp
	hto := &htmp

	for r := 0; r < curl.NumberOfRounds-1; r++ {
		for j := 0; j < curl.StateSize; j++ {
			t1 := curl.Indices[j]
			t2 := curl.Indices[j+1]

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

	for j := 0; j < curl.StateSize; j++ {
		t1, t2 := curl.Indices[j], curl.Indices[j+1]

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

func incr(lmid *[curl.StateSize]uint64, hmid *[curl.StateSize]uint64) bool {
	var carry uint64 = 1
	var i int

	// to avoid boundary check, I believe.
	for i = nonceInitStart; i < HashTrinarySize && carry != 0; i++ {
		low := lmid[i]
		high := hmid[i]
		lmid[i] = high ^ low
		hmid[i] = low
		carry = high & (^low)
	}
	return i == HashTrinarySize
}

func seri(l *[curl.StateSize]uint64, h *[curl.StateSize]uint64, n uint) Trits {
	r := make(Trits, NonceTrinarySize)
	for i := nonceOffset; i < HashTrinarySize; i++ {
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

func check(l *[curl.StateSize]uint64, h *[curl.StateSize]uint64, m int) int {
	nonceProbe := hBits
	for i := HashTrinarySize - m; i < HashTrinarySize; i++ {
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

func loop(lmid *[curl.StateSize]uint64, hmid *[curl.StateSize]uint64, m int) (Trits, int64) {
	var lcpy, hcpy [curl.StateSize]uint64
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
func para(in Trits) (*[curl.StateSize]uint64, *[curl.StateSize]uint64) {
	var l, h [curl.StateSize]uint64

	for i := 0; i < curl.StateSize; i++ {
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

func incrN(n int, lmid *[curl.StateSize]uint64, hmid *[curl.StateSize]uint64) {
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

// PoWGo does proof of work on the given trytes using only Go code.
func PoWGo(trytes Trytes, mwm int) (Trytes, error) {
	return powGo(trytes, mwm, nil)
}

func powGo(trytes Trytes, mwm int, optRate chan int64) (Trytes, error) {
	if !stopGO {
		return "", ErrPoWAlreadyRunning
	}

	if trytes == "" {
		return "", ErrInvalidTrytesForPoW
	}

	stopGO = false

	c := curl.NewCurl()
	c.Absorb(trytes[:(TransactionTrinarySize-HashTrinarySize)/3])
	tr := MustTrytesToTrits(trytes)
	copy(c.State, tr[TransactionTrinarySize-HashTrinarySize:])

	var result Trytes
	var rate chan int64
	if optRate != nil {
		rate = make(chan int64, PowProcs)
	}
	exit := make(chan struct{})
	nonceChan := make(chan Trytes)

	for i := 0; i < PowProcs; i++ {
		go func(i int) {
			lmid, hmid := para(c.State)
			lmid[nonceOffset] = low0
			hmid[nonceOffset] = high0
			lmid[nonceOffset+1] = low1
			hmid[nonceOffset+1] = high1
			lmid[nonceOffset+2] = low2
			hmid[nonceOffset+2] = high2
			lmid[nonceOffset+3] = low3
			hmid[nonceOffset+3] = high3

			incrN(i, lmid, hmid)
			nonce, r := loop(lmid, hmid, mwm)

			if rate != nil {
				rate <- int64(math.Abs(float64(r)))
			}
			if r >= 0 && len(nonce) > 0 {
				select {
				case <-exit:
				case nonceChan <- MustTritsToTrytes(nonce):
					stopGO = true
				}
			}
		}(i)
	}

	if rate != nil {
		var rateSum int64
		for i := 0; i < PowProcs; i++ {
			rateSum += <-rate
		}
		optRate <- rateSum
	}

	result = <-nonceChan
	close(exit)
	stopGO = true
	return result, nil
}

// DoPow computes the nonce field for each transaction so that the last MWM-length trits of the
// transaction hash are all zeroes. Starting from the 0 index transaction, the transactions get chained to
// each other through the trunk transaction hash field. The last transaction in the bundle approves
// the given branch and trunk transactions. This function also initializes the attachment timestamp fields.
func DoPoW(trunkTx, branchTx Trytes, trytes []Trytes, mwm uint64, pow PowFunc) ([]Trytes, error) {
	txs, err := AsTransactionObjects(trytes, nil)
	if err != nil {
		return nil, err
	}
	var prev Trytes
	for i := len(txs) - 1; i >= 0; i-- {
		switch {
		case i == len(txs)-1:
			txs[i].TrunkTransaction = trunkTx
			txs[i].BranchTransaction = branchTx
		default:
			txs[i].TrunkTransaction = prev
			txs[i].BranchTransaction = trunkTx
		}

		txs[i].AttachmentTimestamp = time.Now().UnixNano() / 1000000
		txs[i].AttachmentTimestampLowerBound = LowerBoundAttachmentTimestamp
		txs[i].AttachmentTimestampUpperBound = UpperBoundAttachmentTimestamp

		var err error
		txs[i].Nonce, err = pow(MustTransactionToTrytes(&txs[i]), int(mwm))
		if err != nil {
			return nil, err
		}

		prev = txs[i].Hash
	}
	powedTxTrytes := MustTransactionsToTrytes(txs)
	return powedTxTrytes, nil
}
