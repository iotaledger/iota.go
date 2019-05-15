# account -> AvailableBalance()

> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.




## Output

| Return type     | Description |
|:---------------|:--------|
| uint64 | The current available balance on the account. |
| error | Returned when storage or IRI API call problems arise. |




## Example

```go
func ExampleAvailableBalance() 
	balance, err := acc.AvailableBalance()
	if err != nil {
		log.Fatal(err)
	}

	// 1337
	fmt.Println(balance)
}

```
