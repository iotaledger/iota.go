# Curl -> MustSqueeze()
MustSqueeze squeezes out trits of the given length. Length has to be a multiple of HashTrinarySize. It panics if the length is not valid.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| length | int | Required | The length of the trits to squeeze out.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trits | The Trits representation of the hash. |



