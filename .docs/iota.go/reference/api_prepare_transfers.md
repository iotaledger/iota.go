# API -> PrepareTransfers()
PrepareTransfers prepares the transaction trytes by generating a bundle, filling in transfers and inputs, adding remainder and signing all input transactions.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| seed | Trytes | Required | The seed from which to derive private keys and addresses of.  |
| transfers | Transfers | Required | The Transfers to prepare.  |
| options | PrepareTransfersOptions | Required | Options used for preparing the transfers.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []Trytes | The prepared Trytes, ready for Proof-of-Work. |
| error | Returned for invalid parameters and internal errors. |




## Example

```go
func ExamplePrepareTransfers() 
	seed := "IAFPAIDFNBRE..."

	// create a transfer to the given recipient address
	// optionally define a message and tag
	transfers := bundle.Transfers{
		{
			Address: "ASDEF...",
			Value:   80,
		},
	}

	// create inputs for the transfer
	inputs := []api.Input{
		{
			Address:  "BCEDFA...",
			Security: consts.SecurityLevelMedium,
			KeyIndex: 0,
			Balance:  100,
		},
	}

	// create an address for the remainder.
	// in this case we will have 20 iotas as the remainder, since we spend 100 from our input
	// address and only send 80 to the recipient.
	remainderAddress, err := address.GenerateAddress(seed, 1, consts.SecurityLevelMedium)
	if err != nil {
		// handle error
		return
	}

	// we don't need to set the security level or timestamp in the options because we supply
	// the input and remainder addresses.
	prepTransferOpts := api.PrepareTransfersOptions{Inputs: inputs, RemainderAddress: &remainderAddress}

	// prepare the transfer by creating a bundle with the given transfers and inputs.
	// the result are trytes ready for PoW.
	trytes, err := iotaAPI.PrepareTransfers(seed, transfers, prepTransferOpts)
	if err != nil {
		// handle error
		return
	}

	fmt.Println(trytes)
}

```
