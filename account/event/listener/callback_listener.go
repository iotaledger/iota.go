package listener

import (
	"github.com/iotaledger/iota.go/account/event"
	"github.com/iotaledger/iota.go/account/plugins/promoter"
	"github.com/iotaledger/iota.go/account/plugins/transfer/poller"
	"github.com/iotaledger/iota.go/bundle"
	"sync"
)

// CallbackEventListener handles callbacks and registration for events against an EventMachine.
type CallbackEventListener struct {
	em    event.EventMachine
	ids   []uint64
	idsMu sync.Mutex
}

// NewCallbackEventListener creates a new CallbackEventListener using the given EventMachine.
func NewCallbackEventListener(em event.EventMachine) *CallbackEventListener {
	return &CallbackEventListener{em: em, ids: []uint64{}}
}

// Close unregisters all callbacks from the EventMachine.
func (el *CallbackEventListener) Close() error {
	el.idsMu.Lock()
	defer el.idsMu.Unlock()
	for _, id := range el.ids {
		if err := el.em.UnregisterListener(id); err != nil {
			return err
		}
	}
	return nil
}

type PromotionReattachmentEventCallback func(*promoter.PromotionReattachmentEvent)
type BundleEventCallback func(bundle.Bundle)
type SignalEventCallback func()
type BoolEventCallback func(bool)
type ErrorCallback func(error)

// RegPromotions registers the given callback to execute on promotions.
func (el *CallbackEventListener) RegPromotions(f PromotionReattachmentEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f(data.(*promoter.PromotionReattachmentEvent))
	}, promoter.EventPromotion))
}

// RegReattachments registers the given callback to execute on reattachments.
func (el *CallbackEventListener) RegReattachments(f PromotionReattachmentEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f(data.(*promoter.PromotionReattachmentEvent))
	}, promoter.EventReattachment))
}

// RegSentTransfers registers the given callback to execute when a transfer is sent off.
func (el *CallbackEventListener) RegSentTransfers(f BundleEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f(data.(bundle.Bundle))
	}, event.EventSentTransfer))
}

// RegConfirmedTransfers registers the given callback to execute when a transfer got confirmed.
func (el *CallbackEventListener) RegConfirmedTransfers(f BundleEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f(data.(bundle.Bundle))
	}, poller.EventTransferConfirmed))
}

// RegReceivingDeposits registers the given callback to execute when a new deposit is being received.
func (el *CallbackEventListener) RegReceivingDeposits(f BundleEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f(data.(bundle.Bundle))
	}, poller.EventReceivingDeposit))
}

// RegReceivedDeposits registers this listener to listen for received (confirmed) deposits.
func (el *CallbackEventListener) RegReceivedDeposits(f BundleEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f(data.(bundle.Bundle))
	}, poller.EventReceivedDeposit))
}

// RegReceivedMessages registers this listener to listen for incoming messages.
func (el *CallbackEventListener) RegReceivedMessages(f BundleEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f(data.(bundle.Bundle))
	}, poller.EventReceivedMessage))
}

// RegAccountShutdown registers this listener to listen for shutdown messages.
// A shutdown signal is normally only signaled once by the account on it's graceful termination.
func (el *CallbackEventListener) RegAccountShutdown(f SignalEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f()
	}, event.EventShutdown))
}

// RegInputSelection registers this listener to listen to when input selection is executed.
func (el *CallbackEventListener) RegInputSelection(f BoolEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f(data.(bool))
	}, event.EventDoingInputSelection))
}

// RegPreparingTransfer registers this listener to listen to when a transfer is being prepared.
func (el *CallbackEventListener) RegPreparingTransfer(f SignalEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f()
	}, event.EventPreparingTransfer))
}

// RegGettingTransactionsToApprove registers this listener to listen to when transactions to approve are getting fetched.
func (el *CallbackEventListener) RegGettingTransactionsToApprove(f SignalEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f()
	}, event.EventGettingTransactionsToApprove))
}

// RegAttachingToTangle registers this listener to listen to when Proof-of-Work is executed.
func (el *CallbackEventListener) RegAttachingToTangle(f SignalEventCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f()
	}, event.EventAttachingToTangle))
}

// RegInternalErrors registers this listener to listen for internal account errors.
func (el *CallbackEventListener) RegInternalErrors(f ErrorCallback) {
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		f(data.(error))
	}, event.EventError))
}
