# NewAccount()
NewAccount creates a new account. If settings are nil, the account is initialized with the default settings provided by DefaultSettings().
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| setts | *Settings | Optional | The settings to use for the account.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Account | The account itself. |
| error | Returned for misconfiguration or other types of errors. |




## Example

```go
func ExampleNewAccount() 
	newAccount, err := account.NewAccount(account.DefaultSettings())
	if err != nil {
		log.Fatal(err)
	}
	if err := newAccount.Start(); err != nil {
		log.Fatal(err)
	}
}

```
