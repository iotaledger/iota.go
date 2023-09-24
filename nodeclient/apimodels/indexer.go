package apimodels

import (
	"github.com/pasztorpisti/qs"

	iotago "github.com/iotaledger/iota.go/v4"
)

// IndexerResponse is the standard successful response by the indexer.
type IndexerResponse struct {
	// The ledger index at which these outputs where available at.
	LedgerIndex iotago.SlotIndex `serix:"0,mapKey=ledgerIndex"`
	// The maximum count of results that are returned by the node.
	PageSize uint32 `serix:"1,mapKey=pageSize"`
	// The output IDs (transaction hash + output index) of the found outputs.
	Items iotago.HexOutputIDs `serix:"2,mapKey=items,lengthPrefixType=uint16"`
	// The cursor to use for getting the next results.
	Cursor string `serix:"3,mapKey=cursor,omitempty"`
}

// IndexerCursorParams define page size and cursor query parameters.
type IndexerCursorParams struct {
	// The maximum amount of items returned in one call.
	PageSize int `qs:"pageSize,omitempty"`
	// The offset from which to query from.
	Cursor *string `qs:"cursor,omitempty"`
}

// IndexerTimelockParams define timelock query parameters.
type IndexerTimelockParams struct {
	// Filters outputs based on the presence of timelock unlock condition.
	HasTimelock *bool `qs:"hasTimelock,omitempty"`
	// Return outputs that are timelocked before a certain Unix timestamp.
	TimelockedBefore uint32 `qs:"timelockedBefore,omitempty"`
	// Return outputs that are timelocked after a certain Unix timestamp.
	TimelockedAfter uint32 `qs:"timelockedAfter,omitempty"`
}

// IndexerExpirationParams define expiration query parameters.
type IndexerExpirationParams struct {
	// Filters outputs based on the presence of expiration unlock condition.
	HasExpiration *bool `qs:"hasExpiration,omitempty"`
	// Return outputs that expire before a certain Unix timestamp.
	ExpiresBefore uint32 `qs:"expiresBefore,omitempty"`
	// Return outputs that expire after a certain Unix timestamp.
	ExpiresAfter uint32 `qs:"expiresAfter,omitempty"`
	// Filter outputs based on the presence of a specific return address in the expiration unlock condition.
	ExpirationReturnAddressBech32 string `qs:"expirationReturnAddress,omitempty"`
}

// IndexerCreationParams define creation time query parameters.
type IndexerCreationParams struct {
	// Return outputs that were created before a certain Unix timestamp.
	CreatedBefore uint32 `qs:"createdBefore,omitempty"`
	// Return outputs that were created after a certain Unix timestamp.
	CreatedAfter uint32 `qs:"createdAfter,omitempty"`
}

// IndexerStorageDepositParams define storage deposit based query parameters.
type IndexerStorageDepositParams struct {
	// Filters outputs based on the presence of storage deposit return unlock condition.
	HasStorageDepositReturn *bool `qs:"hasStorageDepositReturn,omitempty"`
	// Filter outputs based on the presence of a specific return address in the storage deposit return unlock condition.
	StorageDepositReturnAddressBech32 string `qs:"storageDepositReturnAddress,omitempty"`
}

// IndexerNativeTokenParams define native token based query parameters.
type IndexerNativeTokenParams struct {
	// Filters outputs based on the presence of native tokens in the output.
	HasNativeTokens *bool `qs:"hasNativeTokens,omitempty"`
	// Filter outputs that have at least an amount of native tokens.
	MinNativeTokenCount *uint32 `qs:"minNativeTokenCount,omitempty"`
	// Filter outputs that have at the most an amount of native tokens.
	MaxNativeTokenCount *uint32 `qs:"maxNativeTokenCount,omitempty"`
}

// IndexerUnlockableByAddressParams define address unlock related query parameters.
type IndexerUnlockableByAddressParams struct {
	// Bech32-encoded address that should be searched for.
	UnlockableByAddressBech32 string `qs:"unlockableByAddress,omitempty"`
}

// BasicOutputsQuery defines parameters for an basic outputs query.
type BasicOutputsQuery struct {
	IndexerCursorParams
	IndexerTimelockParams
	IndexerExpirationParams
	IndexerCreationParams
	IndexerStorageDepositParams
	IndexerNativeTokenParams
	IndexerUnlockableByAddressParams

	// Bech32-encoded address that should be searched for.
	AddressBech32 string `qs:"address,omitempty"`
	// Filters outputs based on the presence of validated sender.
	SenderBech32 string `qs:"sender,omitempty"`
	// Filters outputs based on matching tag feature.
	Tag string `qs:"tag,omitempty"`
}

func (query *BasicOutputsQuery) SetOffset(cursor *string) {
	query.Cursor = cursor
}

func (query *BasicOutputsQuery) URLParams() (string, error) {
	return qs.Marshal(query)
}

// AccountsQuery defines parameters for an account outputs query.
type AccountsQuery struct {
	IndexerCursorParams
	IndexerCreationParams
	IndexerNativeTokenParams
	IndexerUnlockableByAddressParams

	// Bech32-encoded state controller address that should be searched for.
	StateControllerBech32 string `qs:"stateController,omitempty"`
	// Bech32-encoded governor address that should be searched for.
	GovernorBech32 string `qs:"governor,omitempty"`
	// Filters outputs based on the presence of validated sender.
	SenderBech32 string `qs:"sender,omitempty"`
	// Filters outputs based on the presence of validated issuer.
	IssuerBech32 string `qs:"issuer,omitempty"`
}

func (query *AccountsQuery) SetOffset(cursor *string) {
	query.Cursor = cursor
}

func (query *AccountsQuery) URLParams() (string, error) {
	return qs.Marshal(query)
}

// FoundriesQuery defines parameters for a foundry outputs query.
type FoundriesQuery struct {
	IndexerCursorParams
	IndexerCreationParams
	IndexerNativeTokenParams

	// Bech32-encoded address that should be searched for.
	AccountAddressBech32 string `qs:"accountAddress,omitempty"`
}

func (query *FoundriesQuery) SetOffset(cursor *string) {
	query.Cursor = cursor
}

func (query *FoundriesQuery) URLParams() (string, error) {
	return qs.Marshal(query)
}

// NFTsQuery defines parameters for an NFT outputs query.
type NFTsQuery struct {
	IndexerCursorParams
	IndexerTimelockParams
	IndexerExpirationParams
	IndexerStorageDepositParams
	IndexerNativeTokenParams
	IndexerCreationParams
	IndexerUnlockableByAddressParams

	// Bech32-encoded address that should be searched for.
	AddressBech32 string `qs:"address,omitempty"`
	// Filters outputs based on the presence of validated sender.
	SenderBech32 string `qs:"sender,omitempty"`
	// Filters outputs based on the presence of validated issuer.
	IssuerBech32 string `qs:"issuer,omitempty"`
	// Filters outputs based on matching tag feature.
	Tag string `qs:"tag,omitempty"`
}

func (query *NFTsQuery) SetOffset(cursor *string) {
	query.Cursor = cursor
}

func (query *NFTsQuery) URLParams() (string, error) {
	return qs.Marshal(query)
}

type OutputsQuery struct {
	IndexerCursorParams
	IndexerNativeTokenParams
	IndexerCreationParams
	IndexerUnlockableByAddressParams
}

func (query *OutputsQuery) SetOffset(cursor *string) {
	query.Cursor = cursor
}

func (query *OutputsQuery) URLParams() (string, error) {
	return qs.Marshal(query)
}

type DelegationOutputsQuery struct {
	IndexerCursorParams
	IndexerCreationParams

	// Bech32-encoded address that should be searched for.
	AddressBech32 string `qs:"address,omitempty"`
	// Filters outputs based on the presence of validator.
	ValidatorBech32 string `qs:"validator,omitempty"`
}

func (query *DelegationOutputsQuery) SetOffset(cursor *string) {
	query.Cursor = cursor
}

func (query *DelegationOutputsQuery) URLParams() (string, error) {
	return qs.Marshal(query)
}
