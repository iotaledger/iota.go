package iotago

func BitMaskHasBit(bm []byte, bit uint) bool {
	byteIndex := bit / 8
	if uint(len(bm)) <= byteIndex {
		return false
	}
	bitIndex := bit % 8

	return bm[byteIndex]&(1<<bitIndex) > 0
}

func BitMaskSetBit(bm []byte, bit uint) []byte {
	newBitmask := bm
	byteIndex := bit / 8
	for uint(len(newBitmask)) <= byteIndex {
		newBitmask = append(newBitmask, 0)
	}
	bitIndex := bit % 8
	newBitmask[byteIndex] |= 1 << bitIndex

	return newBitmask
}
