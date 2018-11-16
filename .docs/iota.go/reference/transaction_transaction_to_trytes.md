# TransactionToTrytes()
TransactionToTrytes converts the transaction to trytes.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| t | *Transaction | Required | The transaction to convert to Trytes.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trytes | The Trytes representation of the transaction. |
| error | Returned for schematically wrong transactions. |



