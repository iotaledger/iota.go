# API -> TraverseBundle()
TraverseBundle fetches the bundle of a given tail transaction by traversing through the trunk transactions. It does not validate the bundle.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trunkTxHash | Hash | Required | The hash of the tail transaction of the bundle.  |
| bndl | Bundle | Required | An empty Bundle in which transactions are added.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundle | The constructed Bundle by traversing through the trunk transactions. |
| error | Returned for invalid parameters and internal errors. |



