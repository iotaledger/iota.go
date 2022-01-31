package nodeclient

import (
	"github.com/iotaledger/iota.go/v3"
	"github.com/pasztorpisti/qs"
)

// IndexerResponse is the standard successful response by the indexer.
type IndexerResponse struct {
	// The milestone index of the ledger while the query was executed.
	LedgerIndex int `json:"ledgerIndex"`
	// The limit of the query.
	Limit int `json:"limit"`
	// The count of result elements.
	Count int `json:"count"`
	// The result output IDs from the query.
	OutputIds iotago.HexOutputIDs `json:"outputIds"`
	// The offset by which to query for the next page.
	Offset string `json:"offset"`
}

// IndexerTimelockParas define timelock query parameters.
type IndexerTimelockParas struct {
	// Filters outputs based on the presence of timelock unlock condition.
	HasTimelockCondition bool `qs:"hasTimelockCondition,omitempty"`
	// Return outputs that are timelocked before a certain Unix timestamp.
	TimelockedBefore int `qs:"timelockedBefore,omitempty"`
	// Return outputs that are timelocked after a certain Unix timestamp.
	TimelockedAfter int `qs:"timelockedAfter,omitempty"`
	// Return outputs that are timelocked before a certain milestone index.
	TimelockedBeforeMilestone int `qs:"timelockedBeforeMilestone,omitempty"`
	// Return outputs that are timelocked after a certain milestone index.
	TimelockedAfterMilestone int `qs:"timelockedAfterMilestone,omitempty"`
}

// IndexerExpirationParas define expiration query parameters.
type IndexerExpirationParas struct {
	// Filters outputs based on the presence of expiration unlock condition.
	HasExpirationCondition bool `qs:"hasExpirationCondition,omitempty"`
	// Return outputs that expire before a certain Unix timestamp.
	ExpiresBefore int `qs:"expiresBefore,omitempty"`
	// Return outputs that expire after a certain Unix timestamp.
	ExpiresAfter int `qs:"expiresAfter,omitempty"`
	// Return outputs that expire before a certain milestone index.
	ExpiresBeforeMilestone int `qs:"expiresBeforeMilestone,omitempty"`
	// Return outputs that expire after a certain milestone index.
	ExpiresAfterMilestone int `qs:"expiresAfterMilestone,omitempty"`
	// Filter outputs based on the presence of a specific return address in the expiration unlock condition.
	ExpirationReturnAddressBech32 string `qs:"expirationReturnAddress,omitempty"`
}

// IndexerCreationParas define creation time query parameters.
type IndexerCreationParas struct {
	// Return outputs that were created before a certain Unix timestamp.
	CreatedBefore int `qs:"createdBefore,omitempty"`
	// Return outputs that were created after a certain Unix timestamp.
	CreatedAfter int `qs:"createdAfter,omitempty"`
}

// IndexerDustParas define dust deposit based query parameters.
type IndexerDustParas struct {
	// Filters outputs based on the presence of dust return unlock condition.
	RequiresDustReturn bool `qs:"requiresDustReturn,omitempty"`
	// Filter outputs based on the presence of a specific return address in the dust deposit return unlock condition.
	DustReturnAddressBech32 string `qs:"dustReturnAddress,omitempty"`
}

// OutputsQuery defines parameters for an outputs query.
type OutputsQuery struct {
	IndexerTimelockParas
	IndexerExpirationParas
	IndexerCreationParas
	IndexerDustParas

	// Bech32-encoded address that should be searched for.
	AddressBech32 string `qs:"address,omitempty"`
	// Filters outputs based on the presence of validated sender.
	SenderBech32 string `qs:"sender,omitempty"`
	// Filters outputs based on matching tag feature block.
	Tag string `qs:"tag,omitempty"`
	// The offset from which to query from.
	Offset string `qs:"offset,omitempty"`
}

func (query *OutputsQuery) OutputType() iotago.OutputType {
	return iotago.OutputExtended
}

func (query *OutputsQuery) SetOffset(offset string) {
	query.Offset = offset
}

func (query *OutputsQuery) URLParas() (string, error) {
	return qs.Marshal(query)
}

// AliasesQuery defines parameters for an alias outputs query.
type AliasesQuery struct {
	// Bech32-encoded state controller address that should be searched for.
	StateControllerBech32 string `qs:"stateController,omitempty"`
	// Bech32-encoded governor address that should be searched for.
	GovernorBech32 string `qs:"governor,omitempty"`
	// Filters outputs based on the presence of validated sender.
	SenderBech32 string `qs:"sender,omitempty"`
	// Filters outputs based on the presence of validated issuer.
	IssuerBech32 string `qs:"issuer,omitempty"`
	// Filters outputs based on matching tag feature block.
	Tag string `qs:"tag,omitempty"`
	// The offset from which to query from.
	Offset string `qs:"offset,omitempty"`
}

func (query *AliasesQuery) OutputType() iotago.OutputType {
	return iotago.OutputAlias
}

func (query *AliasesQuery) SetOffset(offset string) {
	query.Offset = offset
}

func (query *AliasesQuery) URLParas() (string, error) {
	return qs.Marshal(query)
}

// FoundriesQuery defines parameters for a foundry outputs query.
type FoundriesQuery struct {
	IndexerCreationParas
	// Bech32-encoded address that should be searched for.
	AddressBech32 string `qs:"address,omitempty"`
	// The offset from which to query from.
	Offset string `qs:"offset,omitempty"`
}

func (query *FoundriesQuery) OutputType() iotago.OutputType {
	return iotago.OutputFoundry
}

func (query *FoundriesQuery) SetOffset(offset string) {
	query.Offset = offset
}

func (query *FoundriesQuery) URLParas() (string, error) {
	return qs.Marshal(query)
}

// NFTsQuery defines parameters for an NFT outputs query.
type NFTsQuery struct {
	IndexerTimelockParas
	IndexerExpirationParas
	IndexerDustParas
	IndexerCreationParas

	// Bech32-encoded address that should be searched for.
	AddressBech32 string `qs:"address,omitempty"`
	// Filters outputs based on the presence of validated sender.
	SenderBech32 string `qs:"sender,omitempty"`
	// Filters outputs based on the presence of validated issuer.
	IssuerBech32 string `qs:"issuer,omitempty"`
	// Filters outputs based on matching tag feature block.
	Tag string `qs:"tag,omitempty"`
	// The offset from which to query from.
	Offset string `qs:"offset,omitempty"`
}

func (query *NFTsQuery) OutputType() iotago.OutputType {
	return iotago.OutputNFT
}

func (query *NFTsQuery) SetOffset(offset string) {
	query.Offset = offset
}

func (query *NFTsQuery) URLParas() (string, error) {
	return qs.Marshal(query)
}
