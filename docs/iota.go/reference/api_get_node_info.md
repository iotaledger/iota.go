# API -> GetNodeInfo()
GetNodeInfo returns information about the connected node.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.




## Output

| Return type     | Description |
|:---------------|:--------|
| *GetNodeInfoResponse | The node info object describing the response. |
| error | Returned for internal errors. |




## Example

```go
func ExampleGetNodeInfo() 
	nodeInfo, err := iotaAPI.GetNodeInfo()
	if err != nil {
		// handle error
		return
	}
	fmt.Println("latest milestone index:", nodeInfo.LatestMilestoneIndex)
}

```
