package gocks

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/trinary"
	"gopkg.in/h2non/gock.v1"
	"strings"
)

var FindTransactionsByAddressesQuery = Hashes{"VIWGTBNSFOZDBZYRUMSFGHUJYURQHNYQMYVWGQOBNONDZRFJG9VQTAHPBMTWEEMRYIMQFRAC9VYBOLJVDBPTIELAWD"}
var FindTransactionsByAddresses = Hashes{"VIWGTBNSFOZDBZYRUMSFGHUJYURQHNYQMYVWGQOBNONDZRFJG9VQTAHPBMTWEEMRYIMQFRAC9VYBOLJVD"}
var FindTransactionsByBundles = DefaultHashes()
var FindTransactionsByTags = []Trytes{strings.Repeat("A", 27), strings.Repeat("B", 27)}
var FindTransactionsByApprovees = DefaultHashes()

func init() {

	// empty
	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(FindTransactionsCommand{
			Command:               Command{FindTransactionsCmd},
			FindTransactionsQuery: FindTransactionsQuery{Addresses: Hashes{}},
		}).
		Reply(200).
		JSON(FindTransactionsResponse{Hashes: Hashes{}})

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(FindTransactionsCommand{
			Command:               Command{FindTransactionsCmd},
			FindTransactionsQuery: FindTransactionsQuery{Addresses: FindTransactionsByAddresses},
		}).
		Reply(200).
		JSON(FindTransactionsResponse{Hashes: DefaultHashes()})

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(FindTransactionsCommand{
			Command:               Command{FindTransactionsCmd},
			FindTransactionsQuery: FindTransactionsQuery{Bundles: FindTransactionsByBundles},
		}).
		Reply(200).
		JSON(FindTransactionsResponse{Hashes: DefaultHashes()})

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(FindTransactionsCommand{
			Command:               Command{FindTransactionsCmd},
			FindTransactionsQuery: FindTransactionsQuery{Tags: FindTransactionsByTags},
		}).
		Reply(200).
		JSON(FindTransactionsResponse{Hashes: DefaultHashes()})

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(FindTransactionsCommand{
			Command:               Command{FindTransactionsCmd},
			FindTransactionsQuery: FindTransactionsQuery{Approvees: FindTransactionsByApprovees},
		}).
		Reply(200).
		JSON(FindTransactionsResponse{Hashes: DefaultHashes()})

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(FindTransactionsCommand{
			Command:               Command{FindTransactionsCmd},
			FindTransactionsQuery: FindTransactionsQuery{Addresses: Hashes{Bundle[0].Address}},
		}).
		Reply(200).
		JSON(FindTransactionsResponse{Hashes: Hashes{Bundle[0].Hash}})

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(FindTransactionsCommand{
			Command:               Command{FindTransactionsCmd},
			FindTransactionsQuery: FindTransactionsQuery{Addresses: SampleAddresses},
		}).
		Reply(200).
		JSON(FindTransactionsResponse{Hashes: DefaultHashes()})

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(FindTransactionsCommand{
			Command:               Command{FindTransactionsCmd},
			FindTransactionsQuery: FindTransactionsQuery{Addresses: Hashes{SampleAddresses[2]}},
		}).
		Reply(200).
		JSON(FindTransactionsResponse{Hashes: Hashes{}})

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(FindTransactionsCommand{
			Command:               Command{FindTransactionsCmd},
			FindTransactionsQuery: FindTransactionsQuery{Addresses: Hashes{SampleAddresses[0]}},
		}).
		Reply(200).
		JSON(FindTransactionsResponse{Hashes: Hashes{}})

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(FindTransactionsCommand{
			Command:               Command{FindTransactionsCmd},
			FindTransactionsQuery: FindTransactionsQuery{Bundles: Hashes{strings.Repeat("9", 81)}},
		}).
		Reply(200).
		JSON(FindTransactionsResponse{Hashes: DefaultHashes()})
}
