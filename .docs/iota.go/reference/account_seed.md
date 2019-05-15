# InMemorySeedProvider -> Seed()

> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.




## Output

| Return type     | Description |
|:---------------|:--------|
| Trytes | The seed in its Tryte representation. |
| error | Returned when retrieving the seed fails. |




## Example

```go
func ExampleSeed() 
	seedProv := account.NewInMemorySeedProvider("WEOIOSDFX...")
	seed, err := seedProv.Seed()
	if err != nil {
		log.Fatal(err)
	}
	// WEOIOSDFX...
	fmt.Println(seed)
}

```
