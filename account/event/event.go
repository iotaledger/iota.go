package event

import (
	"sync"
)

type Callback func(data interface{})

// EventMachine defines an object which registers listeners and emits events.
type EventMachine interface {
	// Emit emits the given event with the given data.
	Emit(data interface{}, event Event)
	// Register registers the given channel to receive values for the given event.
	RegisterListener(cb Callback, event Event) uint64
	// Unregister unregisters the given channel to receive events.
	UnregisterListener(uint64) error
}

// NewEventMachine creates a new EventMachine.
func NewEventMachine() EventMachine {
	return &eventmachine{
		listeners: make(map[Event]map[uint64]Callback),
	}
}

type eventmachine struct {
	listenersMu           sync.Mutex
	listeners             map[Event]map[uint64]Callback
	nextID                uint64
	firstTransferPollMade bool
}

func (em *eventmachine) Emit(payload interface{}, event Event) {
	em.listenersMu.Lock()
	defer em.listenersMu.Unlock()
	listenersMap, ok := em.listeners[event]
	if !ok {
		return
	}
	for _, listener := range listenersMap {
		listener(payload)
	}
}

func (em *eventmachine) RegisterListener(cb Callback, event Event) uint64 {
	em.listenersMu.Lock()
	id := em.nextID
	listenersMap, ok := em.listeners[event]
	if !ok {
		// allocate map
		listenersMap = make(map[uint64]Callback)
		em.listeners[event] = listenersMap
	}
	listenersMap[id] = cb
	em.nextID++
	em.listenersMu.Unlock()
	return id
}

func (em *eventmachine) UnregisterListener(id uint64) error {
	em.listenersMu.Lock()
	for _, listenersMap := range em.listeners {
		if _, has := listenersMap[id]; has {
			delete(listenersMap, id)
		}
	}
	em.listenersMu.Unlock()
	return nil
}

// Event is an event emitted by the account or account plugin.
type Event int32

const (
	// emitted when a transfer was broadcasted.
	EventSentTransfer Event = iota
	// emitted when input selection is executed.
	EventDoingInputSelection
	// emitted when a new transfer is being prepared.
	EventPreparingTransfer
	// emitted when transactions to approve are fetched.
	EventGettingTransactionsToApprove
	// emitted when Proof-of-Work is being done.
	EventAttachingToTangle
	// emitted for internal errors of all kinds.
	EventError
	// emitted when the account got shutdown cleanly.
	EventShutdown
)

// DiscardEventMachine is an EventMachine which discards all emits.
type DiscardEventMachine struct{}

func (*DiscardEventMachine) Emit(data interface{}, event Event)               {}
func (*DiscardEventMachine) RegisterListener(cb Callback, event Event) uint64 { return 0 }
func (*DiscardEventMachine) UnregisterListener(id uint64) error               { return nil }
