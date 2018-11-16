# API -> StoreAndBroadcast()
StoreAndBroadcast first stores and the broadcasts the given transactions.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | []Trytes | Required | The Trytes to store and broadcast.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []Trytes | The stored and broadcasted Trytes. |
| error | Returned for invalid parameters and internal errors. |



