package listener

import (
	"github.com/iotaledger/iota.go/account/event"
	"github.com/iotaledger/iota.go/account/plugins/promoter"
	"github.com/iotaledger/iota.go/account/plugins/transfer/poller"
	"github.com/iotaledger/iota.go/bundle"
	"sync"
)

// ChannelEventListener handles channels and registration for events against an EventMachine.
// Use the builder methods to register this listener against certain events.
// Once registered for events, you must listen for incoming events on the specific channel.
type ChannelEventListener struct {
	em                           event.EventMachine
	ids                          []uint64
	idsMu                        sync.Mutex
	Promoted                     chan *promoter.PromotionReattachmentEvent
	Reattached                   chan *promoter.PromotionReattachmentEvent
	SentTransfer                 chan bundle.Bundle
	TransferConfirmed            chan bundle.Bundle
	ReceivingDeposit             chan bundle.Bundle
	ReceivedDeposit              chan bundle.Bundle
	ReceivedMessage              chan bundle.Bundle
	InternalError                chan error
	ExecutingInputSelection      chan bool
	PreparingTransfers           chan struct{}
	GettingTransactionsToApprove chan struct{}
	AttachingToTangle            chan struct{}
	Shutdown                     chan struct{}
}

// NewChannelEventListener creates a new ChannelEventListener using the given EventMachine.
func NewChannelEventListener(em event.EventMachine) *ChannelEventListener {
	return &ChannelEventListener{em: em, ids: []uint64{}}
}

// Close frees up all underlying channels from the EventMachine.
func (el *ChannelEventListener) Close() error {
	el.idsMu.Lock()
	defer el.idsMu.Unlock()
	for _, id := range el.ids {
		if err := el.em.UnregisterListener(id); err != nil {
			return err
		}
	}
	return nil
}

// RegPromotions registers this listener to listen for promotions.
func (el *ChannelEventListener) RegPromotions() *ChannelEventListener {
	el.Promoted = make(chan *promoter.PromotionReattachmentEvent)
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.Promoted <- data.(*promoter.PromotionReattachmentEvent)
	}, promoter.EventPromotion))
	return el
}

// RegReattachments registers this listener to listen for reattachments.
func (el *ChannelEventListener) RegReattachments() *ChannelEventListener {
	el.Reattached = make(chan *promoter.PromotionReattachmentEvent)
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.Reattached <- data.(*promoter.PromotionReattachmentEvent)
	}, promoter.EventReattachment))
	return el
}

// RegSentTransfers registers this listener to listen for sent off transfers.
func (el *ChannelEventListener) RegSentTransfers() *ChannelEventListener {
	el.SentTransfer = make(chan bundle.Bundle)
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.SentTransfer <- data.(bundle.Bundle)
	}, event.EventSentTransfer))
	return el
}

// RegConfirmedTransfers registers this listener to listen for sent off confirmed transfers.
func (el *ChannelEventListener) RegConfirmedTransfers() *ChannelEventListener {
	el.TransferConfirmed = make(chan bundle.Bundle)
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.TransferConfirmed <- data.(bundle.Bundle)
	}, poller.EventTransferConfirmed))
	return el
}

// RegReceivingDeposits registers this listener to listen for incoming deposits which are not yet confirmed.
func (el *ChannelEventListener) RegReceivingDeposits() *ChannelEventListener {
	el.ReceivingDeposit = make(chan bundle.Bundle)
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.ReceivingDeposit <- data.(bundle.Bundle)
	}, poller.EventReceivingDeposit))
	return el
}

// RegReceivedDeposits registers this listener to listen for received (confirmed) deposits.
func (el *ChannelEventListener) RegReceivedDeposits() *ChannelEventListener {
	el.ReceivedDeposit = make(chan bundle.Bundle)
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.ReceivedDeposit <- data.(bundle.Bundle)
	}, poller.EventReceivedDeposit))
	return el
}

// RegReceivedMessages registers this listener to listen for incoming messages.
func (el *ChannelEventListener) RegReceivedMessages() *ChannelEventListener {
	el.ReceivedMessage = make(chan bundle.Bundle)
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.ReceivedMessage <- data.(bundle.Bundle)
	}, poller.EventReceivedMessage))
	return el
}

// RegAccountShutdown registers this listener to listen for shutdown messages.
// A shutdown signal is normally only signaled once by the account on it's graceful termination.
func (el *ChannelEventListener) RegAccountShutdown() *ChannelEventListener {
	el.Shutdown = make(chan struct{})
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.Shutdown <- struct{}{}
	}, event.EventShutdown))
	return el
}

// RegInputSelection registers this listener to listen to when input selection is executed.
func (el *ChannelEventListener) RegInputSelection() *ChannelEventListener {
	el.ExecutingInputSelection = make(chan bool)
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.ExecutingInputSelection <- data.(bool)
	}, event.EventDoingInputSelection))
	return el
}

// RegPreparingTransfer registers this listener to listen to when a transfer is being prepared.
func (el *ChannelEventListener) RegPreparingTransfer() *ChannelEventListener {
	el.PreparingTransfers = make(chan struct{})
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.PreparingTransfers <- struct{}{}
	}, event.EventPreparingTransfer))
	return el
}

// RegGettingTransactionsToApprove registers this listener to listen to when transactions to approve are getting fetched.
func (el *ChannelEventListener) RegGettingTransactionsToApprove() *ChannelEventListener {
	el.GettingTransactionsToApprove = make(chan struct{})
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.GettingTransactionsToApprove <- struct{}{}
	}, event.EventGettingTransactionsToApprove))
	return el
}

// RegAttachingToTangle registers this listener to listen to when Proof-of-Work is executed.
func (el *ChannelEventListener) RegAttachingToTangle() *ChannelEventListener {
	el.AttachingToTangle = make(chan struct{})
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.AttachingToTangle <- struct{}{}
	}, event.EventAttachingToTangle))
	return el
}

// RegInternalErrors registers this listener to listen for internal account errors.
func (el *ChannelEventListener) RegInternalErrors() *ChannelEventListener {
	el.InternalError = make(chan error)
	el.ids = append(el.ids, el.em.RegisterListener(func(data interface{}) {
		el.InternalError <- data.(error)
	}, event.EventError))
	return el
}

// All sets this listener up to listen to all account events.
func (el *ChannelEventListener) All() *ChannelEventListener {
	return el.
		RegAccountShutdown().
		RegPromotions().
		RegReattachments().
		RegSentTransfers().
		RegConfirmedTransfers().
		RegReceivedDeposits().
		RegReceivingDeposits().
		RegReceivedMessages().
		RegInputSelection().
		RegPreparingTransfer().
		RegGettingTransactionsToApprove().
		RegAttachingToTangle().
		RegInternalErrors()
}
