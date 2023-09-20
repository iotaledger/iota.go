package iotago

func SimpleFromBytes[T any](versionedFromBytesFunc func(APIProvider, []byte) (*T, int, error), apiProvider APIProvider) func(bytes []byte) (obj *T, consumedBytes int, err error) {
	return func(bytes []byte) (*T, int, error) {
		return versionedFromBytesFunc(apiProvider, bytes)
	}
}
