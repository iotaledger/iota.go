# CARM64ProofOfWork()
CARM64ProofOfWork does proof of work on the given trytes using native C code and __int128 C type (ARM adjusted). This implementation follows common C standards and does not rely on SSE which is AMD64 specific.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | Trytes | Optional |   |
| mwm | int | Optional |   |
| parallelism | ...int | Optional |   |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trytes |  |
| error |  |



