# ValidChecksum()
ValidChecksum checks whether the given checksum corresponds to the given address.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| address | Hash | Required | The address to check.  |
| checksum | Trytes | Required | The checksum to compare against.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| error | Returned if checksums don't match. |




## Example

```go
func ExampleValidChecksum() 
	addr := "ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC"
	checksum := "JHCYLIGUW"
	if err := address.ValidChecksum(addr, checksum); err != nil {
		log.Fatal(err.Error())
	}
}

```
