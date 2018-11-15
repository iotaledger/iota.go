# API -> GetTips()
GetTips returns a list of tips (transactions not referenced by other transactions) as seen by the connected node.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.




## Output

| Return type     | Description |
|:---------------|:--------|
| Hashes | A set of transaction hashes of tips as seen by the connected node. |
| error | Returned for internal errors. |




## Example

```go
func ExampleGetTips() 
	tips, err := iotaAPI.GetTips()
	if err != nil {
		// handle error
		return
	}
	fmt.Println(tips)
}

```
