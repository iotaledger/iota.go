# Recipients -> AsTransfers()
AsTransfers converts the recipients to transfers.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.




## Output

| Return type     | Description |
|:---------------|:--------|
| Transfers |  |




## Example

```go
func ExampleAsTransfers() 
	recipients := account.Recipients{
		{
			Address: "PWBJRKWJX...",
			Value:   200,
		},
		{
			Address: "ASXMROTUF...",
			Value:   400,
		},
	}

	transfers := recipients.AsTransfers()
	// PWBJRKWJX...
	fmt.Println(transfers[0].Address)
}

```
