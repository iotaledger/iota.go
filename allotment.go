package iotago

// Allotments is a slice of Allotment.
type Allotments []Allotment

// Allotment is a struct that represents a list of account IDs and an allotted value.
type Allotment struct {
	AccountID AccountID `serix:"0"`
	Amount    uint64    `serix:"1"`
}
