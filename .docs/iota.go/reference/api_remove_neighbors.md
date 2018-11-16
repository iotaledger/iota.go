# API -> RemoveNeighbors()
RemoveNeighbors removes a list of neighbors from the connected IRI node. This method has a temporary effect until the IRI node relaunches.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| uris | ...string | Required | The neighbors to remove.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| int64 | The amount of neighbors which got removed. |
| error | Returned for internal errors. |




## Example

```go
func ExampleRemoveNeighbors() 
	removed, err := iotaAPI.RemoveNeighbors("udp://example-iri.io:14600")
	if err != nil {
		// handle error
		return
	}
	fmt.Println("neighbors removed:", removed)
}

```
