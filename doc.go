// Package iotago provides IOTA data models, a node API client and builders to craft messages and transactions.
//
// Creating Messages:
//
//	import (
//		"context"
//		"time"
//
//		iotago "github.com/iotaledger/iota.go/v3"
//		"github.com/iotaledger/iota.go/v3/builder"
//		"github.com/iotaledger/iota.go/v3/nodeclient"
//	)
//
//	func sendMessageExample() error {
//		// create a new node API client
//		nodeHTTPAPIClient := nodeclient.New("https://example.com")
//
//		ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
//		defer cancelFunc()
//
//		// fetch the node's info to know the min. required PoW score
//		info, err := nodeHTTPAPIClient.Info(ctx)
//		if err != nil {
//			return err
//		}
//
//		// craft a tagged data payload
//		taggedDataPayload := &iotago.TaggedData{
//			Tag:  []byte("hello world"),
//			Data: []byte{1, 2, 3, 4},
//		}
//
//		// get some tips from the node
//		tipsResponse, err := nodeHTTPAPIClient.Tips(ctx)
//		if err != nil {
//			return err
//		}
//
//		tips, err := tipsResponse.Tips()
//		if err != nil {
//			return err
//		}
//
//		// get the parameters for the storage deposit
//		nodeRentParas := &iotago.DeSerializationParameters{RentStructure: &info.Protocol.RentStructure}
//
//		// build a message by adding the paylod and the tips and then do local Proof-of-Work
//		msg, err := builder.NewMessageBuilder().
//			Payload(taggedDataPayload).
//			ParentsMessageIDs(tips).
//			ProofOfWork(ctx, nodeRentParas, info.Protocol.MinPowScore).
//			Build()
//
//		// submit the message to the node
//		if _, err := nodeHTTPAPIClient.SubmitMessage(ctx, msg, nodeRentParas); err != nil {
//			return err
//		}
//
//		return nil
//	}
package iotago
