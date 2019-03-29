package builder

import (
	"github.com/iotaledger/iota.go/account"
	"github.com/iotaledger/iota.go/account/event"
	"github.com/iotaledger/iota.go/account/plugins/promoter"
	"github.com/iotaledger/iota.go/account/plugins/transfer/poller"
	"github.com/iotaledger/iota.go/account/store"
	"github.com/iotaledger/iota.go/account/timesrc"
	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	"time"
)

// NewBuilder creates a new Builder which uses the default settings
// provided by DefaultSettings().
func NewBuilder() *Builder {
	setts := account.DefaultSettings()
	setts.Plugins = make(map[string]account.Plugin)
	return &Builder{setts}
}

// Builder wraps a Settings object and provides a builder pattern around it.
type Builder struct {
	settings *account.Settings
}

// Build adds the given plugins and creates the account.
func (b *Builder) Build(plugins ...account.Plugin) (account.Account, error) {
	if b.settings.AddrGen == nil {
		b.settings.AddrGen = account.DefaultAddrGen(b.settings.SeedProv, true)
	}
	if b.settings.PrepareTransfers == nil {
		b.settings.PrepareTransfers = account.DefaultPrepareTransfers(b.settings.API, b.settings.SeedProv)
	}
	for _, p := range plugins {
		b.settings.Plugins[p.Name()] = p
	}
	settsCopy := *b.settings
	return account.NewAccount(&settsCopy)
}

// Settings returns the currently built settings.
func (b *Builder) Settings() *account.Settings {
	return b.settings
}

// API sets the underlying API to use.
func (b *Builder) WithAPI(api *api.API) *Builder {
	b.settings.API = api
	return b
}

// Store sets the underlying store to use.
func (b *Builder) WithStore(store store.Store) *Builder {
	b.settings.Store = store
	return b
}

// SeedProvider sets the underlying SeedProvider to use.
func (b *Builder) WithSeedProvider(seedProv account.SeedProvider) *Builder {
	b.settings.SeedProv = seedProv
	return b
}

// WithAddrGenFunc sets the address generation function to use.
func (b *Builder) WithAddrGenFunc(f account.AddrGenFunc) *Builder {
	b.settings.AddrGen = f
	return b
}

// WithPrepareTransfersFunc sets the prepare transfers function to use.
func (b *Builder) WithPrepareTransfersFunc(f account.PrepareTransfersFunc) *Builder {
	b.settings.PrepareTransfers = f
	return b
}

// Seed sets the underlying seed to use.
func (b *Builder) WithSeed(seed Trytes) *Builder {
	b.settings.SeedProv = account.NewInMemorySeedProvider(seed)
	return b
}

// MWM sets the minimum weight magnitude used to send transactions.
func (b *Builder) WithMWM(mwm uint64) *Builder {
	b.settings.MWM = mwm
	return b
}

// Depth sets the depth used when searching for transactions to approve.
func (b *Builder) WithDepth(depth uint64) *Builder {
	b.settings.Depth = depth
	return b
}

// The overall security level used by the account.
func (b *Builder) WithSecurityLevel(level consts.SecurityLevel) *Builder {
	b.settings.SecurityLevel = level
	return b
}

// TimeSource sets the TimeSource to use to get the current time.
func (b *Builder) WithTimeSource(timesource timesrc.TimeSource) *Builder {
	b.settings.TimeSource = timesource
	return b
}

// InputSelectionFunc sets the strategy to determine inputs and usable balance.
func (b *Builder) WithInputSelectionStrategy(strat account.InputSelectionFunc) *Builder {
	b.settings.InputSelectionStrat = strat
	return b
}

// WithDefaultPlugins adds a transfer poller and promoter-reattacher plugin with following settings:
//
// poll incoming/outgoing transfers every 30 seconds (filter by tail tx hash).
// promote/reattach each pending transfer every 30 seconds.
//
// This function must only be called after following settings are initialized:
// API, Store, MWM, Depth, SeedProvider or AddrGen+PrepareTransfers, TimeSource and EventMachine.
func (b *Builder) WithDefaultPlugins() *Builder {
	transferPoller := poller.NewTransferPoller(
		b.settings, poller.NewPerTailReceiveEventFilter(true),
		time.Duration(30)*time.Second,
	)

	promoterReattacher := promoter.NewPromoter(b.settings, time.Duration(30)*time.Second)

	b.settings.Plugins[transferPoller.Name()] = transferPoller
	b.settings.Plugins[promoterReattacher.Name()] = promoterReattacher
	return b
}

// WithEvents instructs the account to emit events using the given EventMachine.
func (b *Builder) WithEvents(em event.EventMachine) *Builder {
	b.settings.EventMachine = em
	return b
}
