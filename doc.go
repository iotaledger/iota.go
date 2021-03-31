// Package iotago provides IOTA data models, a node API client and builders to craft messages and transactions.
//
// Creating Messages
//
//	// create a new node API client
// 	nodeHTTPAPIClient := iotago.NewNodeHTTPAPIClient("https://example.com")
//
//	// fetch the node's info to know the min. required PoW score
//	info, err := nodeHTTPAPIClient.Info()
//	if err != nil {
//		return err
//	}
//
//	// craft an indexation payload
//	indexationPayload := &iotago.Indexation{
//		Index: []byte("hello world"),
//		Data:  []byte{1, 2, 3, 4},
//	}
//
//	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
//	defer cancelFunc()
//
//	// build a message by fetching tips via the node API client and then do local Proof-of-Work
//	msg, err := iotago.NewMessageBuilder().
//		Payload(indexationPayload).
//		Tips(nodeHTTPAPIClient).
//		ProofOfWork(ctx, info.MinPowScore).
//		Build()
//
//	// submit the message to the node
//	if _, err := nodeHTTPAPIClient.SubmitMessage(msg); err != nil {
//		return err
//	}
package iotago
