# GenerateAddresses()
GenerateAddresses generates N new addresses from the given seed, indices and security level.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.

## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| seed | Trytes | true | The seed used for address generation.  |
| start | uint64 | true | The index from which to start to generate addresses from.  |
| count | uint64 | true | The amount of addresses to generate.  |
| secLvl | SecurityLevel | true | The security level used for address generation.  |
| addChecksum |  | false | Whether to append the checksum on the returned address.  |


## Output

| Return type     | Description |
|:---------------|:--------|
| Hashes | The generated addresses. |
| error | Returned for any error occurring during address generation. |



## Example

```go
func ExampleGenerateAddresses() 
	seed := strings.Repeat("9", 81)
	var index uint64 = 0
	secLvl := consts.SecurityLevelMedium
	addrs, err := address.GenerateAddresses(seed, index, 2, secLvl)
	if err != nil {
		log.Fatalf("unable to generate addresses: %s", err.Error())
	}
	fmt.Println(addrs)
	// output:
	// [
	// 	GPB9PBNCJTPGFZ9CCAOPCZBFMBSMMFMARZAKBMJFMTSECEBRWMGLPTYZRAFKUFOGJQVWVUPPABLTTLCIA,
	//  GMLRCFYRCWPZTORXSFCEGKXTVQGPFI9W9EJLERYJMEJGIPLNCLIKCCAOKQEFYUYCEUGIZKCSSJL9JD9SC,
	// ]
}

```
