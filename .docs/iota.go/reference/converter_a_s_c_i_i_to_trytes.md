# ASCIIToTrytes()
ASCIIToTrytes converts an ascii encoded string to trytes.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| s | string | Required | The ASCII string to convert to Trytes.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Trytes | The Trytes representation of the input ASCII string. |
| error | Returned for non ASCII string inputs. |




## Example

```go
func ExampleASCIIToTrytes() 
	trytes, err := converter.ASCIIToTrytes("IOTA")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(trytes) // output: "SBYBCCKB"
}

```
