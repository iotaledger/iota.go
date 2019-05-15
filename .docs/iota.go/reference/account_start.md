# account -> Start()

> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.




## Output

| Return type     | Description |
|:---------------|:--------|
| error | Returned when the account couldn't be started because of a misconfiguration or faulty plugin. |




## Example

```go
func ExampleStart() 
	if err := acc.Start(); err != nil {
		log.Fatal(err)
	}
}

```
