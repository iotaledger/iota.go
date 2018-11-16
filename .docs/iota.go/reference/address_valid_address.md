# ValidAddress()
ValidAddress checks whether the given address is valid.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| address | Hash | Required | The address to check.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| error | Returned if the address is invalid. |




## Example

```go
func ExampleValidAddress() 
	addr := "ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC"
	if err := address.ValidAddress(addr); err != nil {
		log.Fatalf("invalid address: %s", err.Error())
	}
}

```
