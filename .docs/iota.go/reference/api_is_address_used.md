# API -> IsAddressUsed()
IsAddressUsed checks whether an address is used via FindTransactions and WereAddressesSpentFrom.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| address | Hash | Required | The address to check for used state.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether the address is used. |
| error | Returned for invalid parameters and errors. |



