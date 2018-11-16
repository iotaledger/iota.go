# API -> GetInputs()
GetInputs creates and returns an Inputs object by generating addresses and fetching their latest balance.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| seed | Trytes | Required | The seed from which to derive the addresses of.  |
| options | GetInputsOptions | Required | The options used for getting the Inputs.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| *Inputs | The Inputs gathered of the given seed. |
| error | Returned for invalid parameters and internal errors. |



