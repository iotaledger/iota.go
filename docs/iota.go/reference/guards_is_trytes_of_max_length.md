# IsTrytesOfMaxLength()
IsTrytesOfMaxLength checks if input is correct trytes consisting of [9A-Z] and length <= maxLength
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | Trytes | Required | The Trytes to check.  |
| max | int | Required | The max length the Trytes can have.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether it passes the check. |




## Example

```go
func ExampleIsTrytesOfMaxLength() 
	fmt.Println(guards.IsTrytesOfMaxLength("ABCD", 5)) // output: true
	fmt.Println(guards.IsTrytesOfMaxLength("ABCD", 2)) // output: false
}

```
