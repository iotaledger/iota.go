# Recipients -> Sum()
Sum returns the sum of all amounts.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.




## Output

| Return type     | Description |
|:---------------|:--------|
| uint64 | Returns the sum of the transfer value. |




## Example

```go
func ExampleSum() 
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

	// 600
	fmt.Println(recipients.Sum())
}

```
