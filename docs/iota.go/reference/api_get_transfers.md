# API -> GetTransfers()
GetTransfers returns bundles which operated on the given address range specified by the supplied options.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| seed | Trytes | Required | The seed from which to derive the addresses of.  |
| options | GetTransfersOptions | Required | Options for addresses generation.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundles | The Bundles gathered from the given addresses. |
| error | Returned for invalid parameters and internal errors. |



