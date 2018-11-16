# API -> GetNeighbors()
GetNeighbors returns the list of connected neighbors of the connected node.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.




## Output

| Return type     | Description |
|:---------------|:--------|
| Neighbors | The Neighbors of the connected node. |
| error | Returned for internal errors. |




## Example

```go
func ExampleGetNeighbors() 
	neighbors, err := iotaAPI.GetNeighbors()
	if err != nil {
		// handle error
		return
	}
	fmt.Println("address of neighbor 1:", neighbors[0].Address)
}

```
