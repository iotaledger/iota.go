# RemoveChecksums()
RemoveChecksums is a wrapper function around RemoveChecksum for multiple trytes strings.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| inputs | []Trytes | Required | The Trytes (which must be 81/90 in length) slice of which to remove the checksums.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []Trytes | The input Trytes slice without the checksums. |
| error | Returned for inputs of invalid length. |



