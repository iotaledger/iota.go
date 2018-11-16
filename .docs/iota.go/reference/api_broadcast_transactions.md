# API -> BroadcastTransactions()
BroadcastTransactions broadcasts a list of attached transaction trytes to the network. Tip-selection and Proof-of-Work must be done first by calling GetTransactionsToApprove and AttachToTangle or an equivalent attach method.  You may use this method to increase odds of effective transaction propagation.  Persist the transaction trytes in local storage before calling this command for first time, to ensure that reattachment is possible, until your bundle has been included.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | ...Trytes | Required | The Trytes to broadcast.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []Trytes | The broadcasted Trytes. |
| error | Returned for invalid Trytes and internal errors. |




## Example

```go
func ExampleBroadcastTransactions() 
	// trytes which are chained together and had Proof-of-Work done on them
	var finalTrytes []trinary.Trytes
	_, err := iotaAPI.BroadcastTransactions(finalTrytes...)
	if err != nil {
		// handle error
		return
	}
}

```
