# account -> Send()

> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| recipients | ...Recipient | Optional | The recipients to which to send funds or messages to.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| Bundle | The bundle which got sent off |
| error | Returned when any error while gathering inputs, IRI API calls or storage related issues occur. |




## Example

```go
func ExampleSend() 
	// the Send() method expects Recipient(s), which are just
	// simple transfer objects found in the bundle package
	recipient := account.Recipient{
		Address: "SDOFSKOEFSDFKLG...",
		Value:   100,
	}

	bndl, err := acc.Send(recipient)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("sent off bundle %s\n", bndl[0].Hash)
}

```
