# API -> AddNeighbors()
AddNeighbors adds a list of neighbors to the connected IRI node. Assumes addNeighbors command is available on the node. AddNeighbors has only a temporary effect until the node relaunches.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| uris | ...string | Required | The URIs of the neighbors to add. Must be in udp:// or tcp:// format.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| int64 | The actual amount of added neighbors to the connected node. |
| error | Returned for API and internal errors. |




## Example

```go
func ExampleAddNeighbors() 
	iotaAPI, _ := api.ComposeAPI(api.HTTPClientSettings{URI: endpoint})
	added, err := iotaAPI.AddNeighbors("udp://iota.node:14600")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(added)
	// output: 1
}

```
