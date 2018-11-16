# ConvertUnits()
ConvertUnits converts the given value in the base Unit to the given new Unit.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| val | float64 | Required | The source value.  |
| from | Unit | Required | The Unit format of the source value.  |
| to | Unit | Required | The Unit format of the target value.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| float64 | The float64 representation of the target value. |




## Example

```go
func ExampleConvertUnits() 
	conv := units.ConvertUnits(float64(100), units.Mi, units.I)
	fmt.Println(conv)
	// output: 100000000
}

```
