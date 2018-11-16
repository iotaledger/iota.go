# TailTransactionHash()
TailTransactionHash returns the tail transaction's hash.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| bndl | Bundle | Required | The Bundle from which to get the tail transaction of.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Hash |  |




## Example

```go
func ExampleTailTransactionHash() 
	bndl := bundle.Bundle{
		{
			Hash:         "AAAA...",
			CurrentIndex: 0,
			// ...
		},
		{
			Hash:         "BBBB...",
			CurrentIndex: 1,
			// ...
		},
	}
	tailTxHash := bundle.TailTransactionHash(bndl)
	fmt.Println(tailTxHash) // "AAAA..."
}

```
