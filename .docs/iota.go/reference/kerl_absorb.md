# Kerl -> Absorb()
Absorb fills the internal state of the sponge with the given trits. This is only defined for Trit slices that are a multiple of HashTrinarySize long.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| in | Trits | Required | The Trits slice to absorb. Must be a multiple of 243 in length.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| error | Returned for invalid slice lengths and internal errors. |



