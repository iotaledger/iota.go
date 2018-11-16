# API -> GetNewAddress()
GetNewAddress generates and returns a new address by calling FindTransactions and WereAddressesSpentFrom until the first unused address is detected. This stops working after a snapshot.  If the "total" parameter is supplied in the options, then this function simply generates the specified address range without doing any I/O.  It is suggested that the library user keeps track of used addresses and directly generates addresses from the stored information instead of relying on GetNewAddress.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| seed | Trytes | Required | The seed from which to compute the addresses.  |
| options | GetNewAddressOptions | Required | Options used during address generation.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Hashes | The generated address(es). |
| error | Returned for invalid parameters and internal errors. |




## Example

```go
func ExampleGetNewAddress() 
	seed := "HZVEINVKVIKGFRAWRTRXWD9JLIYLCQNCXZRBLDETPIQGKZJRYKZXLTV9JNUVBIAHAGUZVIQWIAWDZ9ACW"
	addr, err := iotaAPI.GetNewAddress(seed, api.GetNewAddressOptions{Index: 0})
	if err != nil {
		// handle error
		return
	}
	fmt.Println(addr)
	// output: PERXVBEYBJFPNEVPJNTCLWTDVOTEFWVGKVHTGKEOYRTZWYTPXGJJGZZZ9MQMHUNYDKDNUIBWINWB9JQLD
}

```
