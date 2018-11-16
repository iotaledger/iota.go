# ValidateBundleSignatures()
ValidateBundleSignatures validates all signatures of the given bundle. Use ValidBundle() if you want to validate the overall structure of the bundle and the signatures.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| bundle | Bundle | Required | The Bundle to validate.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| bool | Whether the signatures are valid or not. |
| error | Returned if an error occurs during validation. |




## Example

```go
func ExampleValidateBundleSignatures() 
	bndl := bundle.Bundle{} // hypothetical finalized Bundle
	valid, err := bundle.ValidateBundleSignatures(bndl)
	if err != nil {
		// handle error
		return
	}
	switch valid {
	case true:
		fmt.Println("bundle is valid")
	case false:
		fmt.Println("bundle is invalid")
	}
}

```
