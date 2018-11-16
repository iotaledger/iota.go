# IsHash()
IsHash checks if input is correct hash (81 trytes or 90)
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
func ExampleIsHash() 
	hash := "ZFPPXWSTIYJCPPMVCCBZR9TISFJALXEXVYMADGTERQLTHAZJMHGWWFIXVCVPJRBUYLKMTLLKMTWMA9999 "
	fmt.Println(guards.IsHash(hash)) // output: true
}

```
