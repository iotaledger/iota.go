# TransactionsToTrytes()
TransactionsToTrytes returns a slice of transaction trytes from the given transactions.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| txs | Transactions | Required | The transactions to convert to Trytes.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []Trytes | The Trytes representation of the transactions. |
| error | Returned for schematically wrong transactions. |



