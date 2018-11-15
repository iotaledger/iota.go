# API -> SendTransfer()
SendTransfer calls PrepareTransfers and then sends off the bundle via SendTrytes.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.

## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| seed | Trytes | false |   |
| depth | uint64 | false |   |
| mwm | uint64 | false |   |
| transfers | Transfers | false |   |
| options | *SendTransfersOptions | false |   |


## Output

| Return type     | Description |
|:---------------|:--------|
| Bundle |  |
| error |  |


