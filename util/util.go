package util

// NumByteLen returns the number of bytes a given number will occupy when serialized
func NumByteLen(n interface{}) int {
	switch n.(type) {
	case bool, int8, uint8, *bool, *int8, *uint8:
		return 1
	case int16, uint16, *int16, *uint16:
		return 2
	case int32, uint32, *int32, *uint32:
		return 4
	case int64, uint64, *int64, *uint64:
		return 8
	case float32, *float32:
		return 4
	case float64, *float64:
		return 8
	}
	panic("NumByteLen type not supported")
}
