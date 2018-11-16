# API -> SendTrytes()
SendTrytes performs Proof-of-Work, stores and then broadcasts the given transactions and returns them.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | []Trytes | Required | The transaction Trytes to send.  |
| depth | uint64 | Required | The depth to use in GetTransactionsToApprove().  |
| mwm | uint64 | Required | The minimum weight magnitude to fulfill.  |
| reference | ...Hash | Optional | The optional reference to use in GetTransactionsToApprove().  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundle |  |
| error |  |



