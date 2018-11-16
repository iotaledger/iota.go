# API -> FindTransactionObjects()
FindTransactionObjects searches for transactions given a query object with addresses, tags and approvees fields. Multiple query fields are supported and FindTransactionObjects returns the intersection of results.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| query | FindTransactionsQuery | Required | The object defining the transactions to search for.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Transactions | The Transactions of the query result. |
| error | Returned for invalid parameters and internal errors. |



