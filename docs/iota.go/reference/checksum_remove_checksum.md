# RemoveChecksum()
RemoveChecksum removes the checksum from the given trytes. The input trytes must be of length HashTrytesSize or AddressWithChecksumTrytesSize.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| input | Trytes | Required | The Trytes (which must be 81/90 in length) of which to remove the checksum.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trytes | The input Trytes without the checksum. |
| error | Returned for inputs of invalid length. |




## Example

```go
func ExampleRemoveChecksum() 
	cs := "JHCYLIGUW"
	addr := "ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC"
	addr += cs
	addrWithoutChecksum, err := checksum.RemoveChecksum(addr)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(addrWithoutChecksum) // output: ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC
}

```
