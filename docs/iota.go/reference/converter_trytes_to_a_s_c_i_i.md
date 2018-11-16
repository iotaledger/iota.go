# TrytesToASCII()
TrytesToASCII converts trytes of even length to an ascii string.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| trytes | Trytes | Required | The input Trytes to convert to an ASCII string.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| string | The computed ASCII string. |
| error | Returned for invalid Trytes or odd length inputs. |




## Example

```go
func ExampleTrytesToASCII() 
	ascii, err := converter.TrytesToASCII("SBYBCCKB")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(ascii) // output: IOTA
}

```
