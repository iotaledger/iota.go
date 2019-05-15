# NewInMemorySeedProvider()
NewInMemorySeedProvider creates a new InMemorySeedProvider providing the given seed.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| seed | Trytes | Optional | The seed to keep in memory.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| SeedProvider | The SeedProvider which will provide the seed. |




## Example

```go
func ExampleNewInMemorySeedProvider() 
	seedProv := account.NewInMemorySeedProvider("WEOIOSDFX...")
	seed, err := seedProv.Seed()
	if err != nil {
		log.Fatal(err)
	}
	// WEOIOSDFX...
	fmt.Println(seed)
}

```
