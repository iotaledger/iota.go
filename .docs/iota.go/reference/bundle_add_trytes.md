# AddTrytes()
AddTrytes adds the given fragments to the txs in the bundle starting from the specified offset.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| bndl | Bundle | Required | The Bundle to add the Trytes to.  |
| fragments | []Trytes | Required | The Trytes fragments to add to the Bundle,  |
| offset | int | Required | The offset at which to start to add the Trytes into the Bundle.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundle | The Bundle with the added fragments. |




## Example

```go
func ExampleAddTrytes() 
	bndl := bundle.Bundle{}
	// fragments get automatically padded
	bndl = bundle.AddTrytes(bndl, []trinary.Trytes{"ASDFEF..."}, 0)
}

```
