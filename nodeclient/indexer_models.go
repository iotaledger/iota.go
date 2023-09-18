package nodeclient

import (
	"github.com/pasztorpisti/qs"

	iotago "github.com/iotaledger/iota.go/v3"
)

// IndexerResponse is the standard successful response by the indexer.
type IndexerResponse struct {
	// The ledger index at which these outputs where available at.
	LedgerIndex iotago.MilestoneIndex `json:"ledgerIndex"`
	// The maximum count of results that are returned by the node.
	PageSize int `json:"pageSize"`
	// The output IDs (transaction hash + output index) of the found outputs.
	Items iotago.HexOutputIDs `json:"items"`
	// The cursor to use for getting the next results.
	Cursor *string `json:"cursor"`
}

// IndexerCursorParas define page size and cursor query parameters.
type IndexerCursorParas struct {
	// The maximum amount of items returned in one call.
	PageSize int `qs:"pageSize,omitempty"`
	// The offset from which to query from.
	Cursor *string `qs:"cursor,omitempty"`
}

// IndexerTimelockParas define timelock query parameters.
type IndexerTimelockParas struct {
	// Filters outputs based on the presence of timelock unlock condition.
	HasTimelock *bool `qs:"hasTimelock,omitempty"`
	// Return outputs that are timelocked before a certain Unix timestamp.
	TimelockedBefore uint32 `qs:"timelockedBefore,omitempty"`
	// Return outputs that are timelocked after a certain Unix timestamp.
	TimelockedAfter uint32 `qs:"timelockedAfter,omitempty"`
}

// IndexerExpirationParas define expiration query parameters.
type IndexerExpirationParas struct {
	// Filters outputs based on the presence of expiration unlock condition.
	HasExpiration *bool `qs:"hasExpiration,omitempty"`
	// Return outputs that expire before a certain Unix timestamp.
	ExpiresBefore uint32 `qs:"expiresBefore,omitempty"`
	// Return outputs that expire after a certain Unix timestamp.
	ExpiresAfter uint32 `qs:"expiresAfter,omitempty"`
	// Filter outputs based on the presence of a specific return address in the expiration unlock condition.
	ExpirationReturnAddressBech32 string `qs:"expirationReturnAddress,omitempty"`
}

// IndexerCreationParas define creation time query parameters.
type IndexerCreationParas struct {
	// Return outputs that were created before a certain Unix timestamp.
	CreatedBefore uint32 `qs:"createdBefore,omitempty"`
	// Return outputs that were created after a certain Unix timestamp.
	CreatedAfter uint32 `qs:"createdAfter,omitempty"`
}

// IndexerStorageDepositParas define storage deposit based query parameters.
type IndexerStorageDepositParas struct {
	// Filters outputs based on the presence of storage deposit return unlock condition.
	HasStorageDepositReturn *bool `qs:"hasStorageDepositReturn,omitempty"`
	// Filter outputs based on the presence of a specific return address in the storage deposit return unlock condition.
	StorageDepositReturnAddressBech32 string `qs:"storageDepositReturnAddress,omitempty"`
}

// IndexerNativeTokenParas define native token based query parameters.
type IndexerNativeTokenParas struct {
	// Filters outputs based on the presence of native tokens in the output.
	HasNativeTokens *bool `qs:"hasNativeTokens,omitempty"`
	// Filter outputs that have at least an amount of native tokens.
	MinNativeTokenCount *uint32 `qs:"minNativeTokenCount,omitempty"`
	// Filter outputs that have at the most an amount of native tokens.
	MaxNativeTokenCount *uint32 `qs:"maxNativeTokenCount,omitempty"`
}

// IndexerUnlockableByAddressParas define address unlock related query parameters.
type IndexerUnlockableByAddressParas struct {
	// Bech32-encoded address that should be searched for.
	UnlockableByAddressBech32 string `qs:"unlockableByAddress,omitempty"`
}

// BasicOutputsQuery defines parameters for an basic outputs query.
type BasicOutputsQuery struct {
	IndexerCursorParas
	IndexerTimelockParas
	IndexerExpirationParas
	IndexerCreationParas
	IndexerStorageDepositParas
	IndexerNativeTokenParas
	IndexerUnlockableByAddressParas

	// Bech32-encoded address that should be searched for.
	AddressBech32 string `qs:"address,omitempty"`
	// Filters outputs based on the presence of validated sender.
	SenderBech32 string `qs:"sender,omitempty"`
	// Filters outputs based on matching tag feature.
	Tag string `qs:"tag,omitempty"`
}

func (query *BasicOutputsQuery) BaseRoute() string {
	return IndexerAPIRouteBasicOutputs
}

func (query *BasicOutputsQuery) SetOffset(cursor *string) {
	query.Cursor = cursor
}

func (query *BasicOutputsQuery) URLParas() (string, error) {
	return qs.Marshal(query)
}

// AliasesQuery defines parameters for an alias outputs query.
type AliasesQuery struct {
	IndexerCursorParas
	IndexerCreationParas
	IndexerNativeTokenParas
	IndexerUnlockableByAddressParas

	// Bech32-encoded state controller address that should be searched for.
	StateControllerBech32 string `qs:"stateController,omitempty"`
	// Bech32-encoded governor address that should be searched for.
	GovernorBech32 string `qs:"governor,omitempty"`
	// Filters outputs based on the presence of validated sender.
	SenderBech32 string `qs:"sender,omitempty"`
	// Filters outputs based on the presence of validated issuer.
	IssuerBech32 string `qs:"issuer,omitempty"`
}

func (query *AliasesQuery) BaseRoute() string {
	return IndexerAPIRouteAliases
}

func (query *AliasesQuery) SetOffset(cursor *string) {
	query.Cursor = cursor
}

func (query *AliasesQuery) URLParas() (string, error) {
	return qs.Marshal(query)
}

// FoundriesQuery defines parameters for a foundry outputs query.
type FoundriesQuery struct {
	IndexerCursorParas
	IndexerCreationParas
	IndexerNativeTokenParas

	// Bech32-encoded address that should be searched for.
	AliasAddressBech32 string `qs:"aliasAddress,omitempty"`
}

func (query *FoundriesQuery) BaseRoute() string {
	return IndexerAPIRouteFoundries
}

func (query *FoundriesQuery) SetOffset(cursor *string) {
	query.Cursor = cursor
}

func (query *FoundriesQuery) URLParas() (string, error) {
	return qs.Marshal(query)
}

// NFTsQuery defines parameters for an NFT outputs query.
type NFTsQuery struct {
	IndexerCursorParas
	IndexerTimelockParas
	IndexerExpirationParas
	IndexerStorageDepositParas
	IndexerNativeTokenParas
	IndexerCreationParas
	IndexerUnlockableByAddressParas

	// Bech32-encoded address that should be searched for.
	AddressBech32 string `qs:"address,omitempty"`
	// Filters outputs based on the presence of validated sender.
	SenderBech32 string `qs:"sender,omitempty"`
	// Filters outputs based on the presence of validated issuer.
	IssuerBech32 string `qs:"issuer,omitempty"`
	// Filters outputs based on matching tag feature.
	Tag string `qs:"tag,omitempty"`
}

func (query *NFTsQuery) BaseRoute() string {
	return IndexerAPIRouteNFTs
}

func (query *NFTsQuery) SetOffset(cursor *string) {
	query.Cursor = cursor
}

func (query *NFTsQuery) URLParas() (string, error) {
	return qs.Marshal(query)
}

type OutputsQuery struct {
	IndexerCursorParas
	IndexerNativeTokenParas
	IndexerCreationParas
	IndexerUnlockableByAddressParas
}

func (query *OutputsQuery) BaseRoute() string {
	return IndexerAPIRouteOutputs
}

func (query *OutputsQuery) SetOffset(cursor *string) {
	query.Cursor = cursor
}

func (query *OutputsQuery) URLParas() (string, error) {
	return qs.Marshal(query)
}
