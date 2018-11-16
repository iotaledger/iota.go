# API -> ReplayBundle()
ReplayBundle reattaches a transfer to the Tangle by selecting tips & performing the Proof-of-Work again. Reattachments are useful in case original transactions are pending and can be done securely as many times as needed.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| tailTxHash | Hash | Required | The hash of the tail transaction of the bundle to reattach.  |
| depth | uint64 | Required | The depth used in GetTransactionstoApprove().  |
| mwm | uint64 | Required | The minimum weight magnitude to fulfill.  |
| reference | ...Hash | Optional | The optional reference to use in GetTransactionsToApprove().  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundle | The newly attached Bundle. |
| error | Returned for invalid parameters and internal errors. |



