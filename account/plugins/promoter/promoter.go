package promoter

import (
	"github.com/iotaledger/iota.go/account"
	"github.com/iotaledger/iota.go/account/event"
	"github.com/iotaledger/iota.go/account/store"
	"github.com/iotaledger/iota.go/account/timesrc"
	"github.com/iotaledger/iota.go/account/util"
	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/transaction"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
	"strings"
	"time"
)

var ErrUnableToPromote = errors.New("unable to promote")
var ErrUnableToReattach = errors.New("unable to reattach")

// own emitted events
const (
	// emitted when a promotion occurred
	EventPromotion event.Event = 200
	// emitted when a reattachment occurred
	EventReattachment event.Event = 201
)

// PromotionReattachmentEvent is emitted when a promotion or reattachment happened.
type PromotionReattachmentEvent struct {
	// The tail tx hash of the first bundle broadcasted to the network.
	OriginTailTxHash Hash `json:"original_tail_tx_hash"`
	// The bundle hash of the promoted/reattached bundle.
	BundleHash Hash `json:"bundle_hash"`
	// The tail tx hash of the promotion transaction.
	PromotionTailTxHash Hash `json:"promotion_tail_tx_hash"`
	// The tail tx hash of the reattached bundle.
	ReattachmentTailTxHash Hash `json:"reattachment_tail_tx_hash"`
}

// NewPromoter creates a new Promoter. If the interval is set to 0, the Promoter will only
// promote/reattach through PromoteReattach().
func NewPromoter(setts *account.Settings, interval time.Duration) *Promoter {
	if setts.EventMachine == nil {
		setts.EventMachine = &event.DiscardEventMachine{}
	}
	promoter := &Promoter{
		interval: interval, setts: setts,
		tailCache: make(map[string]*transaction.Transaction),
	}
	promoter.syncTimer = util.NewSyncIntervalTimer(interval, promoter.promote, func(err error) {
		promoter.setts.EventMachine.Emit(err, event.EventError)
	})
	return promoter
}

// Promoter is an account plugin which takes care of trying to get pending transfers
// to get confirmed by issuing promotion transactions and creating reattachments.
// The Promoter will only run one promotion/reattachment task at any given time.
type Promoter struct {
	interval  time.Duration
	setts     *account.Settings
	acc       account.Account
	tailCache map[string]*transaction.Transaction
	syncTimer *util.SyncIntervalTimer
}

func (p *Promoter) Name() string {
	return "promoter-reattacher"
}

func (p *Promoter) Start(acc account.Account) error {
	p.acc = acc
	go p.syncTimer.Start()
	return nil
}

func (p *Promoter) Shutdown() error {
	p.syncTimer.Stop()
	return nil
}

// PromoteReattach awaits the current promotion/reattachment task to finish (in case it's ongoing),
// pauses the task, executes a manual task, resumes the repeated task and then returns.
// PromoteReattach will block infinitely if called after the account has been shutdown.
func (p *Promoter) PromoteReattach() error {
	p.syncTimer.Pause()
	err := p.promote()
	p.syncTimer.Resume()
	return err
}

const approxAboveMaxDepthMinutes = 5

func aboveMaxDepth(timesource timesrc.TimeSource, ts time.Time) (bool, error) {
	currentTime, err := timesource.Time()
	if err != nil {
		return false, err
	}
	return currentTime.Sub(ts).Minutes() < approxAboveMaxDepthMinutes, nil
}

const maxDepth = 15
const referenceTooOldMsg = "reference transaction is too old"

var emptySeed = strings.Repeat("9", 81)
var ErrUnpromotableTail = errors.New("tail is unpromoteable")

func (p *Promoter) promote() error {
	pendingTransfers, err := p.setts.Store.GetPendingTransfers(p.acc.ID())
	if err != nil {
		return err
	}
	if len(pendingTransfers) == 0 {
		return nil
	}
	send := func(preparedBundle []Trytes, tips *api.TransactionsToApprove) (Hash, error) {
		readyBundle, err := p.setts.API.AttachToTangle(tips.TrunkTransaction, tips.BranchTransaction, p.setts.MWM, preparedBundle)
		if err != nil {
			return "", errors.Wrap(err, "performing PoW for promote/reattach cycle bundle failed")
		}
		readyBundle, err = p.setts.API.StoreAndBroadcast(readyBundle)
		if err != nil {
			return "", errors.Wrap(err, "unable to store/broadcast bundle in promote/reattach cycle")
		}
		tailTx, err := transaction.AsTransactionObject(readyBundle[0])
		if err != nil {
			return "", err
		}
		return tailTx.Hash, nil
	}

	promote := func(tailTx Hash) (Hash, error) {
		depth := p.setts.Depth
		for {
			tips, err := p.setts.API.GetTransactionsToApprove(depth, tailTx)
			if err != nil {
				if err.Error() == referenceTooOldMsg {
					depth++
					if depth > maxDepth {
						return "", ErrUnpromotableTail
					}
					continue
				}
				return "", err
			}
			pTransfers := bundle.Transfers{bundle.EmptyTransfer}
			preparedBundle, err := p.setts.API.PrepareTransfers(emptySeed, pTransfers, api.PrepareTransfersOptions{})
			if err != nil {
				return "", errors.Wrap(err, "unable to prepare promotion bundle")
			}
			return send(preparedBundle, tips)
		}
	}

	reattach := func(essenceBndl bundle.Bundle) (Hash, error) {
		tips, err := p.setts.API.GetTransactionsToApprove(p.setts.Depth)
		if err != nil {
			return "", errors.Wrapf(err, "unable to GTTA for reattachment in promote/reattach cycle (bundle %s)", essenceBndl[0].Bundle)
		}
		essenceTrytes, err := transaction.TransactionsToTrytes(essenceBndl)
		if err != nil {
			return "", err
		}
		// reverse order of the trytes as PoW needs them from high to low index
		for left, right := 0, len(essenceTrytes)-1; left < right; left, right = left+1, right-1 {
			essenceTrytes[left], essenceTrytes[right] = essenceTrytes[right], essenceTrytes[left]
		}
		return send(essenceTrytes, tips)
	}

	storeTailTxHash := func(key string, tailTxHash string, msg string) error {
		if err := p.setts.Store.AddTailHash(p.acc.ID(), key, tailTxHash); err != nil {
			if err == store.ErrPendingTransferNotFound {
				// might have been removed by polling goroutine
				return nil
			}
			return errors.Wrap(err, msg)
		}
		return nil
	}

	tailsSet := map[string]struct{}{}
	for key, pendingTransfer := range pendingTransfers {
		// search for a tail transaction which is consistent and above max depth
		var tailToPromote string
		// go in reverse order to start from the most recent tails.
		// we iterate over all tails even if a promotable tail was found in between
		// in order to build the set of tail tx hashes, which is later used
		// to clean up the cache from confirmed transactions.
		for i := len(pendingTransfer.Tails) - 1; i >= 0; i-- {
			tailTx := pendingTransfer.Tails[i]

			// add tail to set for cleanup of cache
			tailsSet[tailTx] = struct{}{}

			if tailToPromote != "" {
				continue
			}

			tx, isCached := p.tailCache[tailTx]
			if !isCached {
				consistent, _, err := p.setts.API.CheckConsistency(tailTx)
				if err != nil {
					continue
				}

				if !consistent {
					continue
				}

				txTrytes, err := p.setts.API.GetTrytes(tailTx)
				if err != nil {
					continue
				}

				tx, err = transaction.AsTransactionObject(txTrytes[0])
				if err != nil {
					continue
				}
				p.tailCache[tailTx] = tx
			}

			if above, err := aboveMaxDepth(p.setts.TimeSource, time.Unix(int64(tx.Timestamp), 0)); !above || err != nil {
				continue
			}

			tailToPromote = tailTx
		}

		// promote as tail was found
		if len(tailToPromote) > 0 {
			promoteTailTxHash, err := promote(tailToPromote)
			if err != nil {
				return errors.Wrap(ErrUnableToPromote, err.Error())
			}
			// only load bundle once promotion was successful
			bndl, err := store.PendingTransferToBundle(pendingTransfer)
			if err != nil {
				return errors.Wrapf(err, "unable to translate pending transfer to bundle for reattachment")
			}
			p.setts.EventMachine.Emit(&PromotionReattachmentEvent{
				BundleHash:          bndl[0].Bundle,
				PromotionTailTxHash: promoteTailTxHash,
				OriginTailTxHash:    key,
			}, EventPromotion)
			continue
		}

		// load bundle as we need to reattach
		bndl, err := store.PendingTransferToBundle(pendingTransfer)
		if err != nil {
			return errors.Wrapf(err, "unable to translate pending transfer to bundle for reattachment")
		}

		// reattach
		reattachTailTxHash, err := reattach(bndl)
		if err != nil {
			return errors.Wrap(ErrUnableToReattach, err.Error())
		}
		p.setts.EventMachine.Emit(&PromotionReattachmentEvent{
			BundleHash:             bndl[0].Bundle,
			OriginTailTxHash:       key,
			ReattachmentTailTxHash: reattachTailTxHash,
		}, EventReattachment)

		if err := storeTailTxHash(key, reattachTailTxHash, "unable to store reattachment tx tail hash"); err != nil {
			return err
		}

		promoteTailTxHash, err := promote(reattachTailTxHash)
		if err != nil {
			return errors.Wrap(ErrUnableToPromote, err.Error())
		}
		p.setts.EventMachine.Emit(&PromotionReattachmentEvent{
			BundleHash:          bndl[0].Bundle,
			OriginTailTxHash:    key,
			PromotionTailTxHash: promoteTailTxHash,
		}, EventPromotion)
	}

	// clear cache
	for cachedTailTx := range p.tailCache {
		if _, has := tailsSet[cachedTailTx]; !has {
			delete(p.tailCache, cachedTailTx)
		}
	}
	return nil
}
