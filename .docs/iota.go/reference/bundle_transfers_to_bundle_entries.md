# TransfersToBundleEntries()
TransfersToBundleEntries translates transfers to bundle entries.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| timestamp | uint64 | Required | The timestamp (Unix epoch/seconds) for each entry/transaction.  |
| transfers | ...Transfer | Required | Transfer objects to convert to BundleEntries.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| BundleEntries | The converted bundle entries. |
| error | Returned for invalid addresses etc. |




## Example

```go
func ExampleTransfersToBundleEntries() 
	// Unix epoch in seconds
	ts := uint64(time.Now().UnixNano() / int64(time.Second))
	transfers := bundle.Transfers{
		{
			Address: strings.Repeat("9", 81),
			Tag:     strings.Repeat("9", 27),
			Value:   0,
			// if the message of the transfer would not fit
			// into one transaction, then TransfersToBundleEntries()
			// will create multiple entries for that transaction
			Message: "",
		},
	}
	bundleEntries, err := bundle.TransfersToBundleEntries(ts, transfers...)
	if err != nil {
		// handle error
		return
	}
	// add bundle entries to bundle with bundle.AddEntry()
	_ = bundleEntries
}

```
