# Curl -> SqueezeTrytes()
SqueezeTrytes squeezes out trytes of the given trit length. Length has to be a multiple of HashTrinarySize.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| length | int | Required | The length of the trits to squeeze out.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trytes | The Trytes representation of the hash. |
| error | Returned for invalid lengths. |



