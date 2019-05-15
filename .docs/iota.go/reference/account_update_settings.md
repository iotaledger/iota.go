# account -> UpdateSettings()

> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| setts | *Settings | Optional | The new settings to be applied to the account.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| error | Returned when any error occurs while applying the node settings. |




## Example

```go
func ExampleUpdateSettings() 
	newSetts := account.DefaultSettings()
	newSetts.Depth = 1
	newSetts.MWM = 9
	if err := acc.UpdateSettings(newSetts); err != nil {
		log.Fatal(err)
	}
	fmt.Println("updated account settings")
}

```
