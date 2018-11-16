# AsTransactionObjects()
AsTransactionObjects constructs new transactions from the given raw trytes.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| rawTrytes | []Trytes | Required | The transaction Trytes to convert.  |
| hashes | Hashes | Optional | The hashes to add to the parsed transactions.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Transactions | The parsed Transactions. |
| error | Returned for schematically wrong transactions. |



