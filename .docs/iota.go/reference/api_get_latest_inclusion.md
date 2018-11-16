# API -> GetLatestInclusion()
GetLatestInclusion fetches inclusion states of the given transactions by calling GetInclusionStates using the latest solid subtangle milestone from GetNodeInfo.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| txHashes | Hashes | Required | The hashes of the transactions to check for inclusion state.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []bool | The inclusion states. |
| error | Returned for invalid parameters and internal errors. |




## Example

```go
func ExampleGetLatestInclusion() 
	txHash := "CLXCQVSDAOHWLGKVLNUKKJOOANL9OVGEHSNGRQFLOZJUSJSSXBGJDROUHALTSNUPMTSAVFF9IQEEA9999"
	inclusionStates, err := iotaAPI.GetLatestInclusion(trinary.Hashes{txHash})
	if err != nil {
		// handle error
		return
	}
	fmt.Println("inclusion state by latest milestone?", inclusionStates[0])
}

```
