# account -> IsNew()

> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.




## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether the account is new or not (has actual account data in the store). |
| error | Returned when storage problems arise during the check. |




## Example

```go
func ExampleIsNew() 
	isNew, err := acc.IsNew()
	if err != nil {
		log.Fatal(err)
	}
	switch isNew {
	case true:
		fmt.Println("the account is new")
	case false:
		fmt.Println("the account is not new")
	}
}

```
