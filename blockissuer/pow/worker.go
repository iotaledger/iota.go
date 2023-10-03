package pow

import (
	"context"
	"math"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/ierrors"
)

var (
	// ErrCanceled gets returned when the context for the PoW was canceled.
	ErrCanceled = ierrors.New("canceled")

	// ErrDone gets returned when the PoW was done but no valid nonce was found.
	ErrDone = ierrors.New("done")
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

// Mine performs the PoW for data.
// It returns a nonce that appended to data results in at least targetTrailingZeros.
// The computation can be canceled anytime using ctx.
func (w *Worker) Mine(ctx context.Context, data []byte, targetTrailingZeros int) (uint64, error) {
	var (
		done    uint32
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

	workerWidth := math.MaxUint64 / uint64(w.numWorkers)
	for i := 0; i < w.numWorkers; i++ {
		startNonce := uint64(i) * workerWidth
		wg.Add(1)
		go func() {
			defer wg.Done()

			nonce, workerErr := w.worker(powDigest, startNonce, targetTrailingZeros, &done)
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
		return 0, ErrCanceled
	}

	return nonce, nil
}

func (w *Worker) worker(powDigest []byte, startNonce uint64, targetTrailingZeros int, done *uint32) (uint64, error) {
	if targetTrailingZeros > MaxTrailingZeros {
		panic("pow: invalid target trailing zeros")
	}

	for nonce := startNonce; atomic.LoadUint32(done) == 0; nonce++ {
		if trailingZeros := TrailingZeros(powDigest, nonce); trailingZeros >= targetTrailingZeros {
			return nonce, nil
		}
	}

	return 0, ErrDone
}
