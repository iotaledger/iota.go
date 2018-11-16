# API -> GetTransactionsToApprove()
GetTransactionsToApprove does the tip selection via the connected node.  Returns a pair of approved transactions which are chosen randomly after validating the transaction trytes, the signatures and cross-checking for conflicting transactions.  Tip selection is executed by a Random Walk (RW) starting at random point in the given depth, ending up to the pair of selected tips. For more information about tip selection please refer to the whitepaper (http://iotatoken.com/IOTA_Whitepaper.pdf).  The reference option allows to select tips in a way that the reference transaction is being approved too. This is useful for promoting transactions, for example with PromoteTransaction().
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| depth | uint64 | Required | How many milestones back to begin the Random Walk from.  |
| reference | ...Hash | Optional | A hash of a transaction which should be approved by the returned tips.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| *TransactionsToApprove | Trunk and branch transaction hashes selected by the Random Walk. |
| error | Returned for internal errors. |




## Example

```go
func ExampleGetTransactionsToApprove() 
	tips, err := iotaAPI.GetTransactionsToApprove(3)
	if err != nil {
		// handle error
		return
	}
	fmt.Println("trunk", tips.TrunkTransaction)
	fmt.Println("branch", tips.BranchTransaction)
}

```
