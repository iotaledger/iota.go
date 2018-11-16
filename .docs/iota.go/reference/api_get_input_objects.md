# API -> GetInputObjects()
GetInputObjects creates an Input object using the given addresses, balances, start index and security level.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| addresses | Hashes | Required | The addresses to convert.  |
| balances | []uint64 | Required | The balances of the addresses,  |
| start | uint64 | Required | The start index of the addresses.  |
| secLvl | SecurityLevel | Required | The used security level for generating the addresses.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Inputs | The computed Inputs from the given addresses. |



