# IsTrytes()
IsTrytes checks if input is correct trytes consisting of [9A-Z]
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | Trytes | Required | The Trytes to check.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether the Trytes are valid. |




## Example

```go
func ExampleIsTrytes() 
	fmt.Println(guards.IsTrytes("ABCD")) // output: true
}

```
