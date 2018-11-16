# IsEmptyTrytes()
IsEmptyTrytes checks if input is null (all 9s trytes)
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
func ExampleIsEmptyTrytes() 
	fmt.Println(guards.IsEmptyTrytes("99999999")) // output: true
	fmt.Println(guards.IsEmptyTrytes("ABCD"))     // output: false
}

```
