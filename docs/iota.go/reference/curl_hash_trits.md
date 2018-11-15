# HashTrits()
HashTrits returns hash of the given trits.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.

## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trits | Trits | true | The Trits of which to compute the hash of.  |


## Output

| Return type     | Description |
|:---------------|:--------|
| Trits |  |



## Example

```go
func ExampleHashTrits() 
	trytes := "PDFIDVWRXONZSPJJQVZVVMLGSVB"
	trits := trinary.MustTrytesToTrits(trytes)
	tritsHash := curl.HashTrits(trits)
	fmt.Println(tritsHash) // output: [0 1 -1 0 -1 0 -1 1 ...]
}

```
