# API -> GetTransactionObjects()
GetTransactionObjects fetches transaction objects given an array of transaction hashes.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| hashes | ...Hash | Required | The hashes of the transaction to get.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Transactions | The Transactions of the given hashes. |
| error | Returned for invalid parameters and internal errors. |



