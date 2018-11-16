# API -> BroadcastBundle()
BroadcastBundle re-broadcasts all transactions in a bundle given the tail transaction hash. It might be useful when transactions did not properly propagate, particularly in the case of large bundles.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| tailTxHash | Hash | Required | The hash of the tail transaction of the bundle.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []Trytes | The Trytes of all transactions of the bundle. |
| error | Returned for invalid tail transaction hashes and internal error. |




## Example

```go
func ExampleBroadcastBundle() 
	hash := "CLXCQVSDAOHWLGKVLNUKKJOOANL9OVGEHSNGRQFLOZJUSJSSXBGJDROUHALTSNUPMTSAVFF9IQEEA9999"
	bundleTrytes, err := iotaAPI.BroadcastBundle(hash)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(bundleTrytes)
}

```
