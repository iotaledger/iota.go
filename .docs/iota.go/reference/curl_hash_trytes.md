# HashTrytes()
HashTrytes returns the hash of the given trytes.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| t | Trytes | Required | The Trytes of which to compute the hash of.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trytes | The Trytes representation of the hash. |
| error | Returned for internal errors. |




## Example

```go
func ExampleHashTrytes() 
	trytes := "PDFIDVWRXONZSPJJQVZVVMLGSVB"
	hash, err := curl.HashTrytes(trytes)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(hash) // output: UXBXSI9LHCPYFFZXOWALCBTUIVXYKMCEDDIFXXGXJ9ZLEWKOTXSGYHPEAD9SXSRAWM9TPPXWZMZSIEKGX
}

```
