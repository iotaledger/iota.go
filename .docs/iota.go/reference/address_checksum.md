# Checksum()
Checksum returns the checksum of the given address.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| address | Hash | Required | The address from which to compute the checksum.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trytes | The checksum of the address. |
| error | Returned for invalid addresses or checksum errors. |




## Example

```go
func ExampleChecksum() 
	addr := "ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC"
	addrWithChecksum, err := address.Checksum(addr)
	if err != nil {
		log.Fatalf("unable to compute checksum: %s", err.Error())
	}
	fmt.Println(addrWithChecksum) // output: JHCYLIGUW
}

```
