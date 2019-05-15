# account -> Shutdown()

> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.




## Output

| Return type     | Description |
|:---------------|:--------|
| error | Return when the account couldn't shutdown. |




## Example

```go
func ExampleShutdown() 
	if err := acc.Start(); err != nil {
		log.Fatal(err)
	}
}

```
