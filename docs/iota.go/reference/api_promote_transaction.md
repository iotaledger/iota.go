# API -> PromoteTransaction()
PromoteTransaction promotes a transaction by adding other transactions (spam by default) on top of it. If an optional Context is supplied, PromoteTransaction() will promote the given transaction until the Context is done/cancelled. If no Context is provided, PromoteTransaction() will create one promote transaction.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.

## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| tailTxHash | Hash | false |   |
| depth | uint64 | false |   |
| mwm | uint64 | false |   |
| spamTransfers | Transfers | false |   |
| options | PromoteTransactionOptions | false |   |


## Output

| Return type     | Description |
|:---------------|:--------|
| Transactions |  |
| error |  |


