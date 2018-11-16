# API -> GetBalances()
GetBalances fetches confirmed balances of the given addresses at the latest solid milestone.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| addresses | Hashes | Required | The addresses of which to get the balances of.  |
| threshold | uint64 | Required | The threshold of the query, must be less than or equal 100.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| *Balances | The object describing the result of the balance query. |
| error | Returned for invalid addresses and internal errors. |




## Example

```go
func ExampleGetBalances() 
	balances, err := iotaAPI.GetBalances(trinary.Hashes{"DJDMZD9G9VMGR9UKMEYJWYRLUDEVWTPQJXIQAAXFGMXXSCONBGCJKVQQZPXFMVHAAPAGGBMDXESTZ9999"}, 100)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(balances.Balances[0])
}

```
