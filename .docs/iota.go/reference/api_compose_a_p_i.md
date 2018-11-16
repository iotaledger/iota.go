# ComposeAPI()
ComposeAPI composes a new API from the given settings and provider. If no provider function is supplied, then the default http provider is used. Settings must not be nil.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| settings | Settings | Required | The settings used for creating the Provider.  |
| createProvider | ...CreateProviderFunc | Optional | A function which creates a new Provider given the Settings.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| *API | The composed API object. |
| error | Returned for invalid settings and internal errors. |




## Example

```go
func ExampleComposeAPI() 
	endpoint := "https://example-iri-node.io:14265"

	// a new API object using HTTP connecting to https://example-iri-node.io:14265.
	// this API object will use AttachToTangle() on the remote node.
	iotaAPI, err := api.ComposeAPI(api.HTTPClientSettings{URI: endpoint})
	if err != nil {
		// handle error
		return
	}

	// this API object will perform Proof-of-Work locally
	_, powFunc := pow.GetFastestProofOfWorkImpl()
	iotaAPI, err = api.ComposeAPI(api.HTTPClientSettings{
		URI:                  endpoint,
		LocalProofOfWorkFunc: powFunc,
	})
	if err != nil {
		// handle error
		return
	}

	// this API object will perform Proof-of-Work locally
	// and have a default timeout of 10 seconds
	httpClient := &http.Client{Timeout: time.Duration(10) * time.Second}
	iotaAPI, err = api.ComposeAPI(api.HTTPClientSettings{
		URI:                  endpoint,
		LocalProofOfWorkFunc: powFunc,
		Client:               httpClient,
	})
	if err != nil {
		// handle error
		return
	}

	_ = iotaAPI
}

```
