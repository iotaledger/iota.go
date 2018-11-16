# IsTailTransaction()
IsTailTransaction checks if given transaction object is tail transaction. A tail transaction is one with currentIndex = 0.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| t | *Transaction | Required | The Transaction to validate.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether the Transaction is a tail Transaction. |



