# NormalizedBundleHash()
NormalizedBundleHash normalizes the given bundle hash, with resulting digits summing to zero. It returns a slice with the tryte decimal representation without any 13/M values.
> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.


## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| bundleHash | Hash | Required | The bundle hash to normalize.  |




## Output

| Return type     | Description |
|:---------------|:--------|
| []int8 | The normalized int8 slice without any 13/M values. |




## Example

```go
func ExampleNormalizedBundleHash() 
	bundleHash := "VAJOHANFEOTRSIPCLG9MIPENDFPLQQUGSBLBHMKZ9XVCUSWIKJOOHSPWJAXVLPTAKMPURYAYD9ONODVOW"
	normBundleHash := signing.NormalizedBundleHash(bundleHash)
	fmt.Println(normBundleHash)
	// output: []int8{8, 1, 10, -12, 8, 1, -13, 6, 5, -12, -7, -9, -8, 9, -11, 3, 12, 7, 0, 13, 9, -11, 5, -13, 4, 6, -11, -3, -10, -10, -6, 7, -8, 2, 12, 2, 8, 13, 11, -1, 0, -3, -5, 3, -6, -8, -4, 9, 11, 10, -12, -12, 8, -8, 13, 13, 13, 13, 13, -5, 12, -11, -7, 1, 11, 13, -11, -6, -9, -2, 1, -2, 4, 0, -12, -13, -12, 4, -5, -12, -4}
}

```
