# IsTrytesOfExactLength()
IsTrytesOfExactLength checks if input is correct trytes consisting of [9A-Z] and given length
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | Trytes | Required | The Trytes to check.  |
| length | int | Required | The length the Trytes must have.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether it passes the check. |




## Example

```go
func ExampleIsTrytesOfExactLength() 
	fmt.Println(guards.IsTrytesOfExactLength("ABCD", 4)) // output: true
	fmt.Println(guards.IsTrytesOfExactLength("ABCD", 2)) // output: false
}

```
