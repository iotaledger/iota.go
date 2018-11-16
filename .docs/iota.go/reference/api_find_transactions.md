# API -> FindTransactions()
FindTransactions searches for transaction hashes. It allows to search for transactions by passing a query object with addresses, bundle hashes, tags and/or approvees fields. Multiple query fields are supported and FindTransactions returns the intersection of the results.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| query | FindTransactionsQuery | Required | The object defining the transactions to search for.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Hashes | The Hashes of the query result. |
| error | Returned for invalid query objects and internal errors. |




## Example

```go
func ExampleFindTransactions() 
	txHashes, err := iotaAPI.FindTransactionObjects(api.FindTransactionsQuery{
		Approvees: []trinary.Trytes{
			"DJDMZD9G9VMGR9UKMEYJWYRLUDEVWTPQJXIQAAXFGMXXSCONBGCJKVQQZPXFMVHAAPAGGBMDXESTZ9999",
		},
	})
	if err != nil {
		// handle error
		return
	}
	fmt.Println(txHashes)
}

```
