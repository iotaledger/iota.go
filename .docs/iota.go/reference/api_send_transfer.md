# API -> SendTransfer()
SendTransfer calls PrepareTransfers and then sends off the bundle via SendTrytes.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| seed | Trytes | Required | The seed from which to derive private keys and addresses of.  |
| depth | uint64 | Required | The depth used in GetTransactionsToApprove().  |
| mwm | uint64 | Required | The minimum weight magnitude to fufill.  |
| transfers | Transfers | Required | The Transfers to prepare and send off.  |
| options | *SendTransfersOptions | Required | The options used for preparing and sending of the bundle.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundle | The sent of Bundle. |
| error | Returned for invalid parameters and internal errors. |



