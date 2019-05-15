# DefaultSettings()
DefaultSettings returns Settings initialized with default values: empty seed (81x "9" trytes), mwm: 14, depth: 3, security level: 2, no event machine, system clock as the time source, default input sel. strat, in-memory store, iota-api pointing to localhost, no transfer poller plugin, no promoter-reattacher plugin.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| setts | ...Settings | Optional |   |




## Output

| Return type     | Description |
|:---------------|:--------|
| *Settings |  |



