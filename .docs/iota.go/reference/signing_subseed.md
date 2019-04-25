# Subseed()
Subseed takes a seed and an index and returns the given subseed. Optionally takes the SpongeFunction to use. Default is Kerl.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| seed | Trytes | Required | The seed from which to derive the subseed from.  |
| index | uint64 | Required | The index of the subseed.  |
| spongeFunc | ...SpongeFunction | Optional | The optional sponge function to use.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trits | The Trits representation of the subseed. |
| error | Returned for invalid seeds and internal errors. |




## Example

```go
func ExampleSubseed() 
	seed := "ZLNM9UHJWKTTDEZOTH9CXDEIFUJQCIACDPJIXPOWBDW9LTBHC9AQRIXTIHYLIIURLZCXNSTGNIVC9ISVB"
	subseed, err := signing.Subseed(seed, 0)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(trinary.MustTritsToTrytes(subseed))
	// output: CEFLDDLMF9TO9ZLLTYXIPVFIJKAOFRIQLGNYIDZCTDYSWMNXPYNGFAKHQDY9ABGGQZHEFTXKWKWZXEIUD
}

```
