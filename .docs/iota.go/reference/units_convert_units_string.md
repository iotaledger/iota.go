# ConvertUnitsString()
ConvertUnitsString converts the given string value in the base Unit to the given new Unit.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| val | string | Required | The source string value.  |
| from | Unit | Required | The Unit format of the source value.  |
| to | Unit | Required | The Unit format of the target value.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| float64 | The float64 representation of the target value. |
| error | Returned for invalid string values. |




## Example

```go
func ExampleConvertUnitsString() 
	conv, err := units.ConvertUnitsString("10.1", units.Gi, units.I)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(conv)
	// output: 10100000000
}

```
