# IsTransactionHash()
IsTransactionHash checks whether the given trytes can be a transaction hash.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | Trytes | Required | The Trytes to check.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether it passes the check. |




## Example

```go
func ExampleIsTransactionHash() 
	hash := "ZFPPXWSTIYJCPPMVCCBZR9TISFJALXEXVYMADGTERQLTHAZJMHGWWFIXVCVPJRBUYLKMTLLKMTWMA9999"
	fmt.Println(guards.IsTransactionHash(hash)) // output: true
}

```
