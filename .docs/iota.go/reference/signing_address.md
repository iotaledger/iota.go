# Address()
Address generates the address trits from the given digests. Optionally takes the SpongeFunction to use. Default is Kerl.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| digests | Trits | Required | The digests from which to derive the address from.  |
| spongeFunc | ...SpongeFunction | Optional | The optional sponge function to use.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trits | The Trits representation of the address. |
| error | Returned for internal errors. |




## Example

```go
func ExampleAddress() 
	digests := "MUVADERKIZGMEYJVHGVWBKMQMMXOPWYVOXYPNAGDNKBLHWIBUALWLWSSNDXLYAIIWX9NQRRAOQIVIHWLAIRTWWSF9TGEIKFGMCDWNIXPIYKRTSBHJIONSTSSVUCBYHS9SOZB9PSAOSJUIYQYTUV9NXLZCZWHUALYWW"
	digestsTrits := trinary.MustTrytesToTrits(digests)
	address, err := signing.Address(digestsTrits)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(address)
	// output: CLAAFXEY9AHHCSZCXNKDRZEJHIAFVKYORWNOZAGFPAZYNTSLCXUAG9WBSXBRXYEDPVPLXYVDCBCEKRUBD
}

```
