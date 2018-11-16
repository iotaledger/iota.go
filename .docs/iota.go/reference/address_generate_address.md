# GenerateAddress()
GenerateAddress generates an address deterministically, according to the given seed, index and security level.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| seed | Trytes | Required | The seed used for address generation.  |
| index | uint64 | Required | The index from which to generate the address from.  |
| secLvl | SecurityLevel | Required | The security level used for address generation.  |
| addChecksum | ...bool | Optional | Whether to append the checksum on the returned address.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Hash | The generated address. |
| error | Returned for any error occurring during address generation. |




## Example

```go
func ExampleGenerateAddress() 
	seed := strings.Repeat("9", 81)
	var index uint64 = 0
	secLvl := consts.SecurityLevelMedium
	addr, err := address.GenerateAddress(seed, index, secLvl)
	if err != nil {
		log.Fatalf("unable to generate address: %s", err.Error())
	}
	fmt.Println(addr)
	// output: GPB9PBNCJTPGFZ9CCAOPCZBFMBSMMFMARZAKBMJFMTSECEBRWMGLPTYZRAFKUFOGJQVWVUPPABLTTLCIA
}

```
