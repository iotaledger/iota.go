# DefaultAddrGen()
DefaultAddrGen is the default address generation function used by the account, if non is specified. withCache creates a function which caches the computed addresses by the index and security level for subsequent calls.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| provider | SeedProvider | Optional |   |
| withCache | bool | Optional |   |




## Output

| Return type     | Description |
|:---------------|:--------|
| AddrGenFunc |  |



