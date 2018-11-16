# API -> GetBundle()
GetBundle fetches and validates the bundle given a tail transaction hash by calling TraverseBundle and traversing through trunk transactions.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| tailTxHash | Hash | Required | The hash of the tail transaction of the bundle.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundle | The Bundle of the given tail transaction. |
| error | Returned for invalid parameters and internal errors. |




## Example

```go
func ExampleGetBundle() 
	hash := "CLXCQVSDAOHWLGKVLNUKKJOOANL9OVGEHSNGRQFLOZJUSJSSXBGJDROUHALTSNUPMTSAVFF9IQEEA9999"
	bundle, err := iotaAPI.GetBundle(hash)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(bundle)
}

```
