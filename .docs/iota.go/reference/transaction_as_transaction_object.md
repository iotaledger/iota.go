# AsTransactionObject()
AsTransactionObject makes a new transaction from the given trytes. Optionally the computed transaction hash can be overwritten by supplying an own hash.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | Trytes | Required | The transaction Trytes to convert.  |
| hash | ...Hash | Optional | The hash to add to the transaction.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| *Transaction | The parsed Transaction. |
| error | Returned for schematically wrong transactions. |



