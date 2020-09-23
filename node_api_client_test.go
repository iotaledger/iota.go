package iota_test

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

const nodeAPIUrl = "http://127.0.0.1:8080"

func TestNodeAPI_Info(t *testing.T) {
	defer gock.Off()

	originInfo := &iota.NodeInfoResponse{
		Name:                      "HORNET",
		Version:                   "1.0.0",
		IsHealthy:                 true,
		OperatingNetwork:          "basement",
		Peers:                     9001,
		CoordinatorAddress:        "733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d",
		IsSynced:                  true,
		LatestMilestoneHash:       "5e4a89c549456dbec74ce3a21bde719e9cd84e655f3b1c5a09058d0fbf9417fe",
		LatestMilestoneIndex:      1337,
		LatestSolidMilestoneHash:  "598f7a3186bf7291b8199a3147bb2a81d19b89ac545788b4e5d8adbee7db0f13",
		LatestSolidMilestoneIndex: 666,
		PruningIndex:              142857,
		Time:                      133713371337,
		Features:                  []string{"Lazers"},
	}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteInfo).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originInfo})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	info, err := nodeAPI.Info()
	assert.NoError(t, err)
	assert.EqualValues(t, originInfo, info)
}

func TestNodeAPI_Tips(t *testing.T) {
	defer gock.Off()

	originRes := &iota.NodeTipsResponse{
		Tip1: "733ed2810f2333e9d6cd702c7d5c8264cd9f1ae454b61e75cf702c451f68611d",
		Tip2: "5e4a89c549456dbec74ce3a21bde719e9cd84e655f3b1c5a09058d0fbf9417fe",
	}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteTips).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	tips, err := nodeAPI.Tips()
	assert.NoError(t, err)
	assert.EqualValues(t, originRes, tips)
}

func TestNodeAPI_MessagesByHash(t *testing.T) {
	defer gock.Off()

	identifier := rand32ByteHash()
	queryHash := hex.EncodeToString(identifier[:])

	msg := &iota.Message{
		Version: 1,
		Parent1: rand32ByteHash(),
		Parent2: rand32ByteHash(),
		Payload: nil,
		Nonce:   16345984576234,
	}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteMessagesByHash).
		MatchParam("hashes", queryHash).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: []*iota.Message{msg}})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	msgs, err := nodeAPI.MessagesByHash(iota.MessageHashes{identifier})
	assert.NoError(t, err)
	assert.Len(t, msgs, 1)

	msgJson, err := json.Marshal(msgs[0])
	assert.NoError(t, err)

	originMsgJson, err := msg.MarshalJSON()
	assert.NoError(t, err)

	assert.EqualValues(t, originMsgJson, msgJson)
}

func TestNodeAPI_SubmitMessage(t *testing.T) {
	defer gock.Off()

	msgHash := rand32ByteHash()
	msgHashStr := hex.EncodeToString(msgHash[:])

	incompleteMsg := &iota.Message{Version: 1}
	completeMsg := &iota.Message{
		Version: 1,
		Parent1: rand32ByteHash(),
		Parent2: rand32ByteHash(),
		Payload: nil,
		Nonce:   3495721389537486,
	}

	gock.New(nodeAPIUrl).
		Post(iota.NodeAPIRouteMessageSubmit).
		MatchType("json").
		JSON(incompleteMsg).
		Reply(200).AddHeader("Location", msgHashStr)

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteMessagesByHash).
		MatchParam("hashes", msgHashStr).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: []*iota.Message{completeMsg}})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.SubmitMessage(incompleteMsg)
	assert.NoError(t, err)

	assert.EqualValues(t, completeMsg, resp)
}

func TestNodeAPI_AreMessagesReferencedByMilestone(t *testing.T) {
	defer gock.Off()

	id1 := rand32ByteHash()
	queryHash1 := hex.EncodeToString(id1[:])
	id2 := rand32ByteHash()
	queryHash2 := hex.EncodeToString(id2[:])

	originResp1 := iota.NodeObjectReferencedResponse{
		IsReferencedByMilestone: false,
		MilestoneIndex:          666,
		MilestoneTimestamp:      666666666,
	}

	originResp2 := iota.NodeObjectReferencedResponse{
		IsReferencedByMilestone: false,
		MilestoneIndex:          1337,
		MilestoneTimestamp:      133713371337,
	}

	originRes := []iota.NodeObjectReferencedResponse{originResp1, originResp2}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteMessagesReferencedByMilestone).
		MatchParam("hashes", strings.Join([]string{queryHash1, queryHash2}, ",")).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.AreMessagesReferencedByMilestone(iota.MessageHashes{id1, id2})
	assert.NoError(t, err)
	assert.EqualValues(t, []iota.NodeObjectReferencedResponse{originResp1, originResp2}, resp)
}

func TestNodeAPI_AreTransactionsReferencedByMilestone(t *testing.T) {
	defer gock.Off()

	id1 := rand32ByteHash()
	queryHash1 := hex.EncodeToString(id1[:])

	id2 := rand32ByteHash()
	queryHash2 := hex.EncodeToString(id2[:])

	originResp1 := iota.NodeObjectReferencedResponse{
		IsReferencedByMilestone: false,
		MilestoneIndex:          666,
		MilestoneTimestamp:      666666666,
	}

	originResp2 := iota.NodeObjectReferencedResponse{
		IsReferencedByMilestone: false,
		MilestoneIndex:          1337,
		MilestoneTimestamp:      133713371337,
	}

	originRes := []iota.NodeObjectReferencedResponse{originResp1, originResp2}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteTransactionReferencedByMilestone).
		MatchParam("hashes", strings.Join([]string{queryHash1, queryHash2}, ",")).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.AreTransactionsReferencedByMilestone(iota.MessageHashes{id1, id2})
	assert.NoError(t, err)
	assert.EqualValues(t, originRes, resp)
}

func TestNodeAPI_OutputsByHash(t *testing.T) {
	originOutput, _ := randSigLockedSingleDeposit(iota.AddressEd25519)
	sigDepJson, err := originOutput.MarshalJSON()
	assert.NoError(t, err)
	rawMsgSigDepJson := json.RawMessage(sigDepJson)
	originRes := []iota.NodeOutputResponse{{RawOutput: &rawMsgSigDepJson, Spent: true}}

	utxoInput := &iota.UTXOInput{
		TransactionID:          rand32ByteHash(),
		TransactionOutputIndex: 3,
	}
	utxoInputId := utxoInput.ID()

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteOutputsByID).
		MatchParam("ids", utxoInputId.ToHex()).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.OutputsByID(iota.UTXOInputIDs{utxoInputId})
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.EqualValues(t, originRes, resp)

	respOutput, err := resp[0].Output()
	assert.NoError(t, err)
	assert.EqualValues(t, originOutput, respOutput)
}

func TestNodeAPI_OutputsByAddress(t *testing.T) {
	originOutput, _ := randSigLockedSingleDeposit(iota.AddressEd25519)
	sigDepJson, err := originOutput.MarshalJSON()
	assert.NoError(t, err)
	rawMsgSigDepJson := json.RawMessage(sigDepJson)

	addr, _ := randEd25519Addr()
	addrHex := addr.String()
	originRes := map[string][]iota.NodeOutputResponse{
		addrHex: {{RawOutput: &rawMsgSigDepJson, Spent: true}},
	}

	gock.New(nodeAPIUrl).
		Get(iota.NodeAPIRouteOutputsByAddress).
		MatchParam("addresses", addrHex).
		Reply(200).
		JSON(&iota.HTTPOkResponseEnvelope{Data: originRes})

	nodeAPI := iota.NewNodeAPI(nodeAPIUrl)
	resp, err := nodeAPI.OutputsByAddress(addrHex)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.EqualValues(t, originRes, resp)

	respOutput, err := resp[addrHex][0].Output()
	assert.NoError(t, err)
	assert.EqualValues(t, originOutput, respOutput)
}
