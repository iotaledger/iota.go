# API -> WereAddressesSpentFrom()
WereAddressesSpentFrom checks whether the given addresses were already spent.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| addresses | ...Hash | Required | The addresses to check for spent state.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []bool | The spent states of the addresses. |
| error | Returned for internal errors. |




## Example

```go
func ExampleWereAddressesSpentFrom() 
	spentStates, err := iotaAPI.WereAddressesSpentFrom("AETRKPXQNEK9GWM9ILSODEOZEFDDROCNKYQLWBDHWAEQJIGMSOJSETHNAMZOWDIVVMYPOPSFJRZYMDNRDQSGLFVZNY")
	if err != nil {
		// handle error
		return
	}
	fmt.Println("address spent?", spentStates[0])
}

```
