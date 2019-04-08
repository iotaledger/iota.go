package account

import (
	"github.com/iotaledger/iota.go/account/event"
	"github.com/iotaledger/iota.go/account/store"
	"github.com/iotaledger/iota.go/account/store/inmemory"
	"github.com/iotaledger/iota.go/account/timesrc"
	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
	"strings"
	"sync"
)

// InputSelectionFunc defines a function which given the account, transfer value and the flag balance check,
// computes the inputs for fulfilling the transfer or the usable balance of the account.
// The InputSelectionFunc must obey to the rules of conditional deposit addresses to ensure consistency.
// It returns the computed balance/transfer value, inputs and the key indices to remove from the store.
type InputSelectionFunc func(acc *account, transferValue uint64, balanceCheck bool) (uint64, []api.Input, []uint64, error)

// AddrGenFunc defines a function which given the index, security level and addChecksum flag, generates a new address.
type AddrGenFunc func(index uint64, secLvl consts.SecurityLevel, addChecksum bool) (trinary.Hash, error)

// PrepareTransfersFunc defines a function which prepares the transaction trytes by generating a bundle,
// filling in transfers and inputs, adding the remainder address and signing all input transactions.
type PrepareTransfersFunc func(transfers bundle.Transfers, options api.PrepareTransfersOptions) ([]trinary.Trytes, error)

// DefaultAddrGen is the default address generation function used by the account, if non is specified.
// withCache creates a function which caches the computed addresses by the index and security level for subsequent calls.
func DefaultAddrGen(provider SeedProvider, withCache bool) AddrGenFunc {
	var cacheMu sync.Mutex
	var cache map[uint64]map[consts.SecurityLevel]trinary.Hash

	if withCache {
		cache = map[uint64]map[consts.SecurityLevel]trinary.Hash{}
	}

	read := func(index uint64, secLvl consts.SecurityLevel, addChecksum bool) trinary.Hash {
		cacheMu.Lock()
		defer cacheMu.Unlock()
		m, hasEntry := cache[index]
		if !hasEntry {
			return ""
		}
		cachedAddr, hasAddr := m[secLvl]
		if !hasAddr {
			return ""
		}
		if addChecksum {
			return cachedAddr
		}
		return cachedAddr[:consts.HashTrytesSize]
	}

	write := func(index uint64, secLvl consts.SecurityLevel, addr trinary.Hash) {
		cacheMu.Lock()
		defer cacheMu.Unlock()
		m, hasEntry := cache[index]
		if !hasEntry {
			m = map[consts.SecurityLevel]trinary.Hash{}
			m[secLvl] = addr
			cache[index] = m
			return
		}
		m[secLvl] = addr
	}

	generate := func(index uint64, secLvl consts.SecurityLevel) (trinary.Hash, error) {
		seed, err := provider.Seed()
		if err != nil {
			return "", err
		}

		addr, err := address.GenerateAddress(seed, index, secLvl, true)
		return addr, err
	}

	if !withCache {
		return func(index uint64, secLvl consts.SecurityLevel, addChecksum bool) (trinary.Hash, error) {
			addr, err := generate(index, secLvl)
			if err != nil {
				return "", err
			}
			if addChecksum {
				return addr, nil
			}
			return addr[:consts.HashTrytesSize], nil
		}
	}

	return func(index uint64, secLvl consts.SecurityLevel, addChecksum bool) (trinary.Hash, error) {
		if hash := read(index, secLvl, addChecksum); hash != "" {
			return hash, nil
		}

		addr, err := generate(index, secLvl)
		if err != nil {
			return "", err
		}

		write(index, secLvl, addr)
		if addChecksum {
			return addr, nil
		}
		return addr[:consts.HashTrytesSize], nil
	}
}

// DefaultPrepareTransfers is the default prepare transfers function used by the account, if non is specified.
func DefaultPrepareTransfers(a *api.API, provider SeedProvider) PrepareTransfersFunc {
	return func(transfers bundle.Transfers, options api.PrepareTransfersOptions) ([]trinary.Trytes, error) {
		seed, err := provider.Seed()
		if err != nil {
			return nil, err
		}
		return a.PrepareTransfers(seed, transfers, options)
	}
}

// Settings defines settings used by an account.
// The settings must not be directly mutated after an account was started.
type Settings struct {
	API                 *api.API
	Store               store.Store
	SeedProv            SeedProvider
	MWM                 uint64
	Depth               uint64
	SecurityLevel       consts.SecurityLevel
	TimeSource          timesrc.TimeSource
	InputSelectionStrat InputSelectionFunc
	EventMachine        event.EventMachine
	Plugins             map[string]Plugin
	AddrGen             AddrGenFunc
	PrepareTransfers    PrepareTransfersFunc
}

var emptySeed = strings.Repeat("9", 81)

// DefaultSettings returns Settings initialized with default values:
// empty seed (81x "9" trytes), mwm: 14, depth: 3, security level: 2, no event machine,
// system clock as the time source, default input sel. strat, in-memory store, iota-api pointing to localhost,
// no transfer poller plugin, no promoter-reattacher plugin.
func DefaultSettings(setts ...Settings) *Settings {
	if len(setts) == 0 {
		iotaAPI, _ := api.ComposeAPI(api.HTTPClientSettings{})
		return &Settings{
			MWM:                 14,
			Depth:               3,
			SeedProv:            NewInMemorySeedProvider(emptySeed),
			SecurityLevel:       consts.SecurityLevelMedium,
			TimeSource:          &timesrc.SystemClock{},
			EventMachine:        &event.DiscardEventMachine{},
			API:                 iotaAPI,
			Store:               inmemory.NewInMemoryStore(),
			InputSelectionStrat: DefaultInputSelection,
		}
	}
	defaultValue := func(val uint64, should uint64) uint64 {
		if val == 0 {
			return should
		}
		return val
	}
	sett := setts[0]
	if sett.SecurityLevel == 0 {
		sett.SecurityLevel = consts.SecurityLevelMedium
	}
	sett.Depth = defaultValue(sett.Depth, 3)
	sett.MWM = defaultValue(sett.MWM, 14)
	if sett.TimeSource == nil {
		sett.TimeSource = &timesrc.SystemClock{}
	}
	if sett.InputSelectionStrat == nil {
		sett.InputSelectionStrat = DefaultInputSelection
	}
	if sett.API == nil {
		iotaAPI, _ := api.ComposeAPI(api.HTTPClientSettings{})
		sett.API = iotaAPI
	}
	if sett.Store == nil {
		sett.Store = inmemory.NewInMemoryStore()
	}
	return &sett
}
