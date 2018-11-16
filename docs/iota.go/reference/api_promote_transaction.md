# API -> PromoteTransaction()
PromoteTransaction promotes a transaction by adding other transactions (spam by default) on top of it. If an optional Context is supplied, PromoteTransaction() will promote the given transaction until the Context is done/cancelled. If no Context is provided, PromoteTransaction() will create one promote transaction.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| tailTxHash | Hash | Required | The hash of the tail transaction.  |
| depth | uint64 | Required | The depth used in GetTransactionsToApprove().  |
| mwm | uint64 | Required | The minimum weight magnitude to fulfill.  |
| spamTransfers | Transfers | Required | The spam transaction used for promoting the given tail transaction.  |
| options | PromoteTransactionOptions | Required | Options used during promotion.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Transactions | The promote transactions. |
| error | Returned for inconsistent tail transactions, invalid parameters and internal errors. |




## Example

```go
func ExamplePromoteTransaction() 
	tailTxHash := "SLFKTBMXWQPWF..."

	promotionTransfers := bundle.Transfers{bundle.EmptyTransfer}

	// this will create one promotion transaction
	promotionTx, err := iotaAPI.PromoteTransaction(tailTxHash, 3, 14, promotionTransfers, api.PromoteTransactionOptions{})
	if err != nil {
		// handle error
		return
	}

	fmt.Println("promoted tx with new tx:", promotionTx[0].Hash)

	// options for promotion
	delay := time.Duration(5) * time.Second
	// stop promotion after one minute
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*time.Minute)
	opts := api.PromoteTransactionOptions{
		Ctx: ctx,
		// wait for 5 seconds before each promotion
		Delay: &delay,
	}

	// this promotion will stop until the passed in Context is done
	promotionTx, err = iotaAPI.PromoteTransaction(tailTxHash, 3, 14, promotionTransfers, opts)
	if err != nil {
		// handle error
		return
	}
	fmt.Println("promoted tx with new tx:", promotionTx[0].Hash)
}

```
