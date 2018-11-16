# HashTrits()
HashTrits returns the hash of the given trits.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trits | Trits | Required | The Trits of which to compute the hash of.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trits | The Trits representation of the hash. |
| error | Returned for internal errors. |




## Example

```go
func ExampleHashTrits() 
	trytes := "PDFIDVWRXONZSPJJQVZVVMLGSVB"
	trits := trinary.MustTrytesToTrits(trytes)
	tritsHash, err := curl.HashTrits(trits)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(tritsHash) // output: [0 1 -1 0 -1 0 -1 1 ...]
}

```
