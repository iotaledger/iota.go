package iotago

import (
	"fmt"
)

// SlotIndex is the ID of a slot.
type SlotIndex int64

func (i SlotIndex) String() string {
	return fmt.Sprintf("SlotIndex(%d)", i)
}

// Max returns the maximum of the two given slots.
func (i SlotIndex) Max(other SlotIndex) SlotIndex {
	if i > other {
		return i
	}

	return other
}

// Abs returns the absolute value of the SlotIndex.
func (i SlotIndex) Abs() (absolute SlotIndex) {
	if i < 0 {
		return -i
	}

	return i
}
