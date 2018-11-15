# HashTrytes()
HashTrytes returns hash of t.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | Trytes | true | The Trytes of which to compute the hash of.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trytes |  |




## Example

```go
func ExampleHashTrytes() 
	trytes := "PDFIDVWRXONZSPJJQVZVVMLGSVB"
	hash := curl.HashTrytes(trytes)
	fmt.Println(hash) // output: UXBXSI9LHCPYFFZXOWALCBTUIVXYKMCEDDIFXXGXJ9ZLEWKOTXSGYHPEAD9SXSRAWM9TPPXWZMZSIEKGX
}

```
