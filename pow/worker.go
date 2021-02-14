package pow

import (
	"context"
	"errors"
	"math"
	"math/bits"
	"sync"
	"sync/atomic"

	legacy "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl/bct"
	"github.com/iotaledger/iota.go/encoding/b1t6"
	"github.com/iotaledger/iota.go/trinary"
)

// errors returned by the PoW
var (
	ErrCancelled = errors.New("canceled")
	ErrDone      = errors.New("done")
)

// The Worker performs the PoW.
type Worker struct {
	numWorkers int
}

// New creates a new PoW Worker.
// The optional numWorkers specifies how many go routines should be used to perform the PoW.
func New(numWorkers ...int) *Worker {
	w := &Worker{
		numWorkers: 1,
	}
	if len(numWorkers) > 0 && numWorkers[0] > 0 {
		w.numWorkers = numWorkers[0]
	}
	return w
}

const ln3 = 1.098612288668109691395245236922525704647490557822749451734694333 // https://oeis.org/A002391

// Mine performs the PoW for data.
// It returns a nonce that appended to data results in a PoW score of at least targetScore.
// The computation can be canceled anytime using ctx.
func (w *Worker) Mine(ctx context.Context, data []byte, targetScore float64) (uint64, error) {
	var (
		done    uint32
		counter uint64
		wg      sync.WaitGroup
		results = make(chan uint64, w.numWorkers)
		closing = make(chan struct{})
	)

	// compute the digest
	h := Hash.New()
	h.Write(data)
	powDigest := h.Sum(nil)

	// stop when the context has been canceled
	go func() {
		select {
		case <-ctx.Done():
			atomic.StoreUint32(&done, 1)
		case <-closing:
			return
		}
	}()

	// compute the minimum numbers of trailing zeros required to get a PoW score ≥ targetScore
	targetZeros := uint(math.Ceil(math.Log(float64(len(data)+nonceBytes)*targetScore) / ln3))

	workerWidth := math.MaxUint64 / uint64(w.numWorkers)
	for i := 0; i < w.numWorkers; i++ {
		startNonce := uint64(i) * workerWidth
		wg.Add(1)
		go func() {
			defer wg.Done()

			nonce, workerErr := w.worker(powDigest, startNonce, targetZeros, &done, &counter)
			if workerErr != nil {
				return
			}
			atomic.StoreUint32(&done, 1)
			results <- nonce
		}()
	}
	wg.Wait()
	close(results)
	close(closing)

	nonce, ok := <-results
	if !ok {
		return 0, ErrCancelled
	}
	return nonce, nil
}

func (w *Worker) worker(powDigest []byte, startNonce uint64, target uint, done *uint32, counter *uint64) (uint64, error) {
	if target > legacy.HashTrinarySize {
		panic("pow: invalid trailing zeros target")
	}

	// use batched Curl hashing
	c := bct.NewCurlP81()
	var l, h [legacy.HashTrinarySize]uint

	// allocate exactly one Curl block for each batch index and fill it with the encoded digest
	buf := make([]trinary.Trits, bct.MaxBatchSize)
	for i := range buf {
		buf[i] = make(trinary.Trits, legacy.HashTrinarySize)
		b1t6.Encode(buf[i], powDigest)
	}

	digestTritsLen := b1t6.EncodedLen(len(powDigest))
	for nonce := startNonce; atomic.LoadUint32(done) == 0; nonce += bct.MaxBatchSize {
		// add the nonce to each trit buffer
		for i := range buf {
			nonceBuf := buf[i][digestTritsLen:]
			encodeNonce(nonceBuf, nonce+uint64(i))
		}

		// process the batch
		c.Reset()
		if err := c.Absorb(buf, legacy.HashTrinarySize); err != nil {
			return 0, err
		}
		c.CopyState(l[:], h[:]) // the first 243 entries of the state correspond to the resulting hashes
		atomic.AddUint64(counter, bct.MaxBatchSize)

		// check the state whether it corresponds to a hash with sufficient amount of trailing zeros
		// this is equivalent to computing the hashes with Squeeze and then checking TrailingZeros of each
		if i := checkStateTrits(&l, &h, target); i < bct.MaxBatchSize {
			return nonce + uint64(i), nil
		}
	}
	return 0, ErrDone
}

func checkStateTrits(l, h *[legacy.HashTrinarySize]uint, n uint) int {
	var v uint
	for i := legacy.HashTrinarySize - n; i < legacy.HashTrinarySize; i++ {
		v |= l[i] ^ h[i] // 0 if trit is zero, 1 otherwise
	}
	// return the index of the first zero bit, this corresponds to the index of the hash with n trailing zero trits
	return bits.TrailingZeros(^v)
}
