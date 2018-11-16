# API -> GetAccountData()
GetAccountData returns an AccountData object containing account information about addresses, transactions, inputs and total account balance. Deprecated: Use a solution which uses local persistence to keep the account data.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| seed | Trytes | Required | The seed of the account.  |
| options | GetAccountDataOptions | Required | Options used for gathering the account data.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| *AccountData | An object describing the current state of the account. |
| error | Returned for invalid parameters and internal errors. |




## Example

```go
func ExampleGetAccountData() 
	seed := "CLXCQVSDAOHWLGKVLNUKKJOOANL9OVGEHSNGRQFLOZJUSJSSXBGJDROUHALTSNUPMTSAVFF9IQEEA9999"
	accountData, err := iotaAPI.GetAccountData(seed, api.GetAccountDataOptions{})
	if err != nil {
		// handle error
		return
	}
	fmt.Println(accountData)
}

```
