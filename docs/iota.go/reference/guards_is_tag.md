# IsTag()
IsTag checks that input is valid tag trytes.
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
func ExampleIsTag() 
	tag := "CD9999999999999999999999999"
	fmt.Println(guards.IsTag(tag)) // output: true
}

```
