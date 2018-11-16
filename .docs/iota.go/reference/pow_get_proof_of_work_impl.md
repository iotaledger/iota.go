# GetProofOfWorkImpl()
GetProofOfWorkImpl returns the specified Proof-of-Work implementation given a name.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| name | string | Required | The name of the Proof-of-Work implementation to get.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| ProofOfWorkFunc | The Proof-of-Work implementation. |
| error | Returned if the Proof-of-Work implementation is not known. |



