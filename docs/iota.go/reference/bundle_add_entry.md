# AddEntry()
AddEntry adds a new entry to the bundle. It automatically adds additional transactions if the signature message fragments don't fit into one transaction.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| txs | Bundle | Required | The Bundle to which to add the entry to.  |
| bndlEntry | BundleEntry | Required | The BundleEntry to add.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundle | Returns a bundle with the newly added BundleEntry. |




## Example

```go
func ExampleAddEntry() 
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
}

```
