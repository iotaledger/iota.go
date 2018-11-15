# API -> GetBundlesFromAddresses()
GetBundlesFromAddresses fetches all bundles from the given addresses and optionally sets the confirmed property on each transaction using GetLatestInclusion. This function does not validate the bundles.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.

## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| addresses | Hashes | true | The addresses of which to get the bundles of.  |
| inclusionState |  | false | Whether to set the persistence field on the transactions.  |


## Output

| Return type     | Description |
|:---------------|:--------|
| Bundles |  |
| error | Returned for invalid parameters and internal errors. |



## Example

```go
func ExampleGetBundlesFromAddresses() 
	addresses := trinary.Hashes{
		"PDEUDPV9GACEBLYZCQOMLMHOQWTBBMVMMYUDKJKVFVSLMUIXHUISQGFJKJABIMAVRNGOURDQBBRSCTWBCNYMIBWIZZ",
		"CUCCO99XUKMXHJQNGPZXGQOTWMACGCQHWPGKTCMC9IPOXTXNFTCDDXTUDXLOMDLSCRXKKLVMJSBSCTE9XRCB9FGUXX",
	}
	bundles, err := iotaAPI.GetBundlesFromAddresses(addresses)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(bundles)
}

```
