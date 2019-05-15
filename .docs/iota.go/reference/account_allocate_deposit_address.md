# account -> AllocateDepositAddress()

> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| conds | Conditions | Optional | The conditions for the newly allocated deposit address.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| CDA | The generated conditional deposit address. |
| error | Returned when there's any issue persisting or creating the CDA. |




## Example

```go
func ExampleAllocateDepositAddress() 
	timeoutInOneHour := time.Now().Add(time.Duration(1) * time.Hour)
	conds := &deposit.Conditions{
		TimeoutAt: &timeoutInOneHour,
	}
	cda, err := acc.AllocateDepositAddress(conds)
	if err != nil {
		log.Fatal(err)
	}

	// EQSAUZXULTTYZCLNJNTXQTQHOMOFZERHTCGTXOLTVAHKSA9OGAZDEKECURBRIXIJWNPFCQIOVFVVXJVD9DGCJRJTHZ
	fmt.Println(cda.AsMagnetLink())
}

```
