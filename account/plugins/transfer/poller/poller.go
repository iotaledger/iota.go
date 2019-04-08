package poller

import (
	"github.com/iotaledger/iota.go/account"
	"github.com/iotaledger/iota.go/account/event"
	"github.com/iotaledger/iota.go/account/store"
	"github.com/iotaledger/iota.go/account/util"
	"github.com/iotaledger/iota.go/bundle"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
	"sync"
	"time"
)

// StringSet is a set of strings.
type StringSet map[string]struct{}

// ReceiveEventFilter filters and creates events given the incoming bundles, deposit addresses and spent addresses.
// It's the job of the ReceiveEventFilter to emit the appropriate events through the given EventMachine.
type ReceiveEventFilter func(eventMachine event.EventMachine, bndls bundle.Bundles, depAddrs StringSet, spentAddrs StringSet)

const (
	// emitted when a broadcasted transfer got confirmed
	EventTransferConfirmed event.Event = 100
	// emitted when a deposit is being received
	EventReceivingDeposit event.Event = 101
	// emitted when a deposit is confirmed
	EventReceivedDeposit event.Event = 102
	// emitted when a zero value transaction is received
	EventReceivedMessage event.Event = 103
)

// NewTransferPoller creates a new TransferPoller. If the interval is set to 0, the TransferPoller will only
// poll through Poll().
func NewTransferPoller(setts *account.Settings, filter ReceiveEventFilter, interval time.Duration) *TransferPoller {
	if setts.EventMachine == nil {
		setts.EventMachine = &event.DiscardEventMachine{}
	}
	if filter == nil {
		filter = NewPerTailReceiveEventFilter(true)
	}
	poller := &TransferPoller{
		setts: setts, receiveEventFilter: filter,
	}
	poller.syncTimer = util.NewSyncIntervalTimer(interval, poller.pollTransfers, func(err error) {
		poller.setts.EventMachine.Emit(err, event.EventError)
	})
	return poller
}

// TransferPoller is an account plugin which takes care of checking pending transfers for confirmation
// and checking incoming transfers.
type TransferPoller struct {
	setts              *account.Settings
	receiveEventFilter ReceiveEventFilter
	acc                account.Account
	syncTimer          *util.SyncIntervalTimer
}

func (tp *TransferPoller) Name() string {
	return "transfer-poller"
}

func (tp *TransferPoller) Start(acc account.Account) error {
	tp.acc = acc
	go tp.syncTimer.Start()
	return nil
}

// Poll awaits the current transfer polling task to finish (in case it's ongoing),
// pauses the repeated task, does a manual polling, resumes the repeated task and then returns.
// Poll will block infinitely if called after the account has been shutdown.
func (tp *TransferPoller) Poll() error {
	tp.syncTimer.Pause()
	err := tp.pollTransfers()
	tp.syncTimer.Resume()
	return err
}

func (tp *TransferPoller) Shutdown() error {
	tp.syncTimer.Stop()
	return nil
}

func (tp *TransferPoller) pollTransfers() error {
	pendingTransfers, err := tp.setts.Store.GetPendingTransfers(tp.acc.ID())
	if err != nil {
		return errors.Wrap(err, "unable to load pending transfers for polling transfers")
	}

	depositAddresses, err := tp.setts.Store.GetDepositAddresses(tp.acc.ID())
	if err != nil {
		return errors.Wrap(err, "unable to load deposit addresses for polling transfers")
	}

	// nothing to do
	if len(pendingTransfers) == 0 && len(depositAddresses) == 0 {
		return nil
	}

	// poll incoming/outgoing in parallel
	var outErr, inErr error
	if len(pendingTransfers) > 0 {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			outErr = tp.checkOutgoingTransfers(pendingTransfers)
		}()
		inErr = tp.checkIncomingTransfers(depositAddresses, pendingTransfers)
		wg.Wait()
	} else {
		inErr = tp.checkIncomingTransfers(depositAddresses, pendingTransfers)
	}

	if outErr != nil {
		return outErr
	}
	return inErr
}

func (tp *TransferPoller) checkOutgoingTransfers(pendingTransfers map[string]*store.PendingTransfer) error {
	for tailTx, pendingTransfer := range pendingTransfers {
		if len(pendingTransfer.Tails) == 0 {
			continue
		}
		states, err := tp.setts.API.GetLatestInclusion(pendingTransfer.Tails)
		if err != nil {
			return errors.Wrapf(err, "unable to check latest inclusion state in outgoing transfers op. (first tail tx of bundle: %s)", tailTx)
		}
		// if any state is true we can remove the transfer as it got confirmed
		for i, state := range states {
			if !state {
				continue
			}
			// fetch bundle to emit it in the event
			bndl, err := tp.setts.API.GetBundle(pendingTransfer.Tails[i])
			if err != nil {
				return errors.Wrapf(err, "unable to get bundle in outgoing transfers op. (first tail tx of bundle: %s) of tail %s", tailTx, pendingTransfer.Tails[i])
			}
			tp.setts.EventMachine.Emit(bndl, EventTransferConfirmed)
			if err := tp.setts.Store.RemovePendingTransfer(tp.acc.ID(), tailTx); err != nil {
				return errors.Wrap(err, "unable to remove confirmed transfer from store in outgoing transfers op.")
			}
			break
		}
	}
	return nil
}

func (tp *TransferPoller) checkIncomingTransfers(storedDepositAddresses map[uint64]*store.StoredDepositAddress, pendingTransfers map[string]*store.PendingTransfer) error {
	if len(storedDepositAddresses) == 0 {
		return nil
	}

	depositAddresses := make(StringSet)
	depAddrs := make(Hashes, 0)
	for keyIndex, depositAddress := range storedDepositAddresses {
		// filter remainder address bundles
		if depositAddress.TimeoutAt == nil {
			continue
		}
		addr, err := tp.setts.AddrGen(keyIndex, depositAddress.SecurityLevel, true)
		if err != nil {
			return errors.Wrap(err, "unable to compute deposit address in incoming transfers op.")
		}
		depAddrs = append(depAddrs, addr)
		depositAddresses[addr] = struct{}{}
	}
	if len(depositAddresses) == 0 {
		return nil
	}

	spentAddresses := make(StringSet)
	for _, transfer := range pendingTransfers {
		bndl, err := store.PendingTransferToBundle(transfer)
		if err != nil {
			panic(err)
		}
		for j := range bndl {
			if bndl[j].Value < 0 {
				spentAddresses[bndl[j].Address] = struct{}{}
			}
		}
	}

	// get all bundles which operated on the current deposit addresses
	bndls, err := tp.setts.API.GetBundlesFromAddresses(depAddrs, true)
	if err != nil {
		return errors.Wrap(err, "unable to fetch bundles from deposit addresses in incoming transfers op.")
	}

	// create the events to emit in the event system
	tp.receiveEventFilter(tp.setts.EventMachine, bndls, depositAddresses, spentAddresses)
	return nil
}

type ReceiveEventTuple struct {
	Event  event.Event
	Bundle bundle.Bundle
}

// PerTailFilter filters receiving/received bundles by the bundle's tail transaction hash.
// Optionally takes in a bool flag indicating whether the first pass of the event filter
// should not emit any events.
func NewPerTailReceiveEventFilter(skipFirst ...bool) ReceiveEventFilter {
	var skipFirstEmitting bool
	if len(skipFirst) > 0 && skipFirst[0] {
		skipFirstEmitting = true
	}
	receivedFilter := map[string]struct{}{}
	receivingFilter := map[string]struct{}{}

	return func(em event.EventMachine, bndls bundle.Bundles, ownDepAddrs StringSet, ownSpentAddrs StringSet) {
		events := []ReceiveEventTuple{}

		receivingBundles := make(map[string]bundle.Bundle)
		receivedBundles := make(map[string]bundle.Bundle)

		// filter out transfers to own remainder addresses or where
		// a deposit is an input address (a spend from our own address)
		for _, bndl := range bndls {
			if err := bundle.ValidBundle(bndl); err != nil {
				continue
			}

			tailTx := bundle.TailTransactionHash(bndl)
			if *bndl[0].Persistence {
				receivedBundles[tailTx] = bndl
			} else {
				receivingBundles[tailTx] = bndl
			}
		}

		isValueTransfer := func(bndl bundle.Bundle) bool {
			isValue := false
			for _, tx := range bndl {
				if tx.Value > 0 || tx.Value < 0 {
					isValue = true
					break
				}
			}
			return isValue
		}

		// filter out bundles for which a previous event was emitted
		// and emit new events for the new bundles
		for tailTx, bndl := range receivingBundles {
			if _, has := receivingFilter[tailTx]; has {
				continue
			}
			receivingFilter[tailTx] = struct{}{}
			// determine whether the bundle is a value transfer.
			if isValueTransfer(bndl) {
				events = append(events, ReceiveEventTuple{EventReceivingDeposit, bndl})
				continue
			}
			events = append(events, ReceiveEventTuple{EventReceivedMessage, bndl})
		}

		for tailTx, bndl := range receivedBundles {
			if _, has := receivedFilter[tailTx]; has {
				continue
			}
			receivedFilter[tailTx] = struct{}{}
			if isValueTransfer(bndl) {
				events = append(events, ReceiveEventTuple{EventReceivedDeposit, bndl})
				continue
			}
			// if a message bundle got confirmed but an event was already emitted
			// during its receiving, we don't emit another event
			if _, has := receivingFilter[tailTx]; has {
				continue
			}
			events = append(events, ReceiveEventTuple{EventReceivedMessage, bndl})
		}

		// skip first emitting of events as multiple restarts of the same account
		// would yield the same events.
		if skipFirstEmitting {
			skipFirstEmitting = false
			return
		}

		for _, ev := range events {
			em.Emit(ev.Bundle, ev.Event)
		}
	}
}
