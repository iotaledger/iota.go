# HasValidNonce()
HasValidNonce checks if the transaction has the valid MinWeightMagnitude. MWM corresponds to the amount of zero trits at the end of the transaction hash.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.

## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| t | *Transaction | true | The Transaction to validate.  |
| mwm | uint64 | true | The Minimum Weight Magnitude to check against.  |


## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether the MWM is fulfilled or not. |


