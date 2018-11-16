# KerlTritsToBytes()
KerlTritsToBytes is only defined for hashes, i.e. slices of trits of length 243. It returns 48 bytes.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trits | Trits | Required | The Trits to convert to []byte.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []byte | The converted bytes. |
| error | Returned for invalid lengths and internal errors. |



