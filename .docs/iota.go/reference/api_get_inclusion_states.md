# API -> GetInclusionStates()
GetInclusionStates fetches inclusion states of a given list of transactions.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| txHashes | Hashes | Required | The transaction hashes to check for inclusion state.  |
| tips | ...Hash | Required | The reference tips of which to check whether the transactions were included in or not.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []bool | The inclusion states in the same order as the passed in transaction hashes. |
| error | Returned for invalid transaction/tip hashes and internal errors. |




## Example

```go
func ExampleGetInclusionStates() 
	txHash := "DJDMZD9G9VMGR9UKMEYJWYRLUDEVWTPQJXIQAAXFGMXXSCONBGCJKVQQZPXFMVHAAPAGGBMDXESTZ9999"
	info, err := iotaAPI.GetNodeInfo()
	if err != nil {
		// handle error
		return
	}
	states, err := iotaAPI.GetInclusionStates(trinary.Hashes{txHash}, info.LatestMilestone)
	if err != nil {
		// handle error
		return
	}

	fmt.Println("inclusion?", states[0])
}

```
