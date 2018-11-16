# AddChecksum()
AddChecksum computes the checksum and returns the given trytes with the appended checksum. If isAddress is true, then the input trytes must be of length HashTrytesSize. Specified checksum length must be at least MinChecksumTrytesSize long or it must be AddressChecksumTrytesSize if isAddress is true.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| input | Trytes | Required | The Trytes of which to compute the checksum of.  |
| isAddress | bool | Required | Whether to validate the input as an address.  |
| checksumLength | uint64 | Required | The wanted length of the checksum. Must be 9 when isAddress is true.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trytes | The input Trytes with the appended checksum. |
| error | Returned for invalid addresses and other inputs. |




## Example

```go
func ExampleAddChecksum() 
	addr := "ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC"
	addrWithChecksum, err := checksum.AddChecksum(addr, true, consts.AddressChecksumTrytesSize)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(addrWithChecksum) // output: JHCYLIGUW
}

```
