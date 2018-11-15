# API -> SendTrytes()
SendTrytes performs Proof-of-Work, stores and then broadcasts the given transactions and returns them.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | []Trytes | false |   |
| depth | uint64 | false |   |
| mwm | uint64 | false |   |
| reference |  | false |   |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundle |  |
| error |  |



