# Finalize()
Finalize finalizes the bundle by calculating the bundle hash and setting it on each transaction bundle hash field.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| bundle | Bundle | Required | The Bundle to finalize.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundle | The finalized Bundle. |
| error | Returned for invalid finalization. |




## Example

```go
func ExampleFinalize() 
	// Unix epoch in seconds
	ts := uint64(time.Now().UnixNano() / int64(time.Second))
	transfers := bundle.Transfers{
		{
			Address: strings.Repeat("9", 81),
			Tag:     strings.Repeat("9", 27),
			Value:   0,
			Message: "",
		},
	}
	bundleEntries, err := bundle.TransfersToBundleEntries(ts, transfers...)
	if err != nil {
		// handle error
		return
	}

	bndl := bundle.Bundle{}
	for _, entry := range bundleEntries {
		bundle.AddEntry(bndl, entry)
	}

	fmt.Println(len(bndl)) // 1

	finalizedBundle, err := bundle.Finalize(bndl)
	if err != nil {
		// handle error
		return
	}

	// finalized bundle, ready for PoW
	_ = finalizedBundle
}

```
