# Kerl -> Squeeze()
Squeeze out length trits. Length has to be a multiple of HashTrinarySize.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| length | int | Required | The length of the Trits to squeeze out. Must be a multiple of 243.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trits | The squeezed out Trits. |
| error | Returned for invalid lengths and internal errors. |



