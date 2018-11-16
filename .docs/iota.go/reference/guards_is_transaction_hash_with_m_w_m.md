# IsTransactionHashWithMWM()
IsTransactionHashWithMWM checks if input is correct transaction hash (81 trytes) with given MWM
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | Trytes | Required | The Trytes to check.  |
| mwm | uint | Optional |   |




## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether it passes the check. |




## Example

```go
func ExampleIsTransactionHashWithMWM() 
	hash := "ZFPPXWSTIYJCPPMVCCBZR9TISFJALXEXVYMADGTERQLTHAZJMHGWWFIXVCVPJRBUYLKMTLLKMTWMA9999"
	fmt.Println(guards.IsTransactionTrytesWithMWM(hash, 14)) // output: true
}

```
