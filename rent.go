package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// VByteCostFactor defines the type of the virtual byte cost factor.
type VByteCostFactor byte

const (
	// VByteCostFactorData defines the multiplier for data fields.
	VByteCostFactorData VByteCostFactor = 1
	// VByteCostFactorKey defines the multiplier for fields which can act as keys for lookups.
	VByteCostFactorKey VByteCostFactor = 10
)

var (
	// ErrVByteRentNotCovered gets returned when a NonEphemeralObject does not cover the state rent
	// cost which are calculated from its virtual byte costs.
	ErrVByteRentNotCovered = errors.New("virtual byte rent costs not covered")
	// ErrTypeIsNotSupportedRentStructure gets returned when a serializable was found to not be a supported RentStructure.
	ErrTypeIsNotSupportedRentStructure = errors.New("serializable is not a supported rent structure")
)

// Multiply multiplies in with this factor.
func (factor VByteCostFactor) Multiply(in uint64) uint64 {
	return uint64(factor) * in
}

// With joins two factors with each other.
func (factor VByteCostFactor) With(other VByteCostFactor) VByteCostFactor {
	return factor + other
}

// RentStructure defines the parameters of rent cost calculations on objects which take node resources.
type RentStructure struct {
	// Defines the rent of a single virtual byte denoted in IOTA tokens.
	VByteCost uint32
	// Defines the factor to be used for data only fields.
	VBFactorData VByteCostFactor
	// defines the factor to be used for key/lookup generating fields.
	VBFactorKey VByteCostFactor
}

func (r *RentStructure) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	var factorData uint8
	var factorKey uint8

	return serializer.NewDeserializer(data).
		ReadNum(&r.VByteCost, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize virtual byte cost within rent structure", err)
		}).
		ReadNum(&factorData, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize virtual byte factor data within rent structure", err)
		}).
		Do(func() {
			r.VBFactorData = VByteCostFactor(factorData)
		}).
		ReadNum(&factorKey, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize virtual byte factor key within rent structure", err)
		}).
		Do(func() {
			r.VBFactorKey = VByteCostFactor(factorKey)
		}).
		Done()
}

func (r *RentStructure) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(&r.VByteCost, func(err error) error {
			return fmt.Errorf("%w: unable to serialize virtual byte cost within rent structure", err)
		}).
		WriteNum(&r.VBFactorData, func(err error) error {
			return fmt.Errorf("%w: unable to serialize virtual byte factor data within rent structure", err)
		}).
		WriteNum(&r.VBFactorKey, func(err error) error {
			return fmt.Errorf("%w: unable to serialize virtual byte factor key within rent structure", err)
		}).
		Serialize()
}

func (r *RentStructure) MarshalJSON() ([]byte, error) {
	jRentStructure := &jsonRentStructure{
		VByteCost:    r.VByteCost,
		VBFactorData: uint8(r.VBFactorData),
		VBFactorKey:  uint8(r.VBFactorKey),
	}
	return json.Marshal(jRentStructure)
}

func (r *RentStructure) UnmarshalJSON(data []byte) error {
	jRentStructure := &jsonRentStructure{}
	if err := json.Unmarshal(data, jRentStructure); err != nil {
		return err
	}
	seri, err := jRentStructure.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*RentStructure)
	return nil
}

// jsonRentStructure defines the json representation of a RentStructure.
type jsonRentStructure struct {
	VByteCost    uint32 `json:"vByteCost"`
	VBFactorData uint8  `json:"vByteFactorData"`
	VBFactorKey  uint8  `json:"vByteFactorKey"`
}

func (j *jsonRentStructure) ToSerializable() (serializer.Serializable, error) {
	return &RentStructure{
		VByteCost:    j.VByteCost,
		VBFactorData: VByteCostFactor(j.VBFactorData),
		VBFactorKey:  VByteCostFactor(j.VBFactorKey),
	}, nil
}

// CoversStateRent tells whether given this NonEphemeralObject, the given rent fulfills the renting costs
// by examining the virtual bytes cost of the object.
// Returns the minimum rent computed and an error if it is not covered by rent.
func (r *RentStructure) CoversStateRent(object NonEphemeralObject, rent uint64) (uint64, error) {
	minRent := r.MinRent(object)
	if rent < minRent {
		return 0, fmt.Errorf("%w: needed %d but only got %d", ErrVByteRentNotCovered, minRent, rent)
	}
	return minRent, nil
}

// MinRent returns the minimum rent to cover a given object.
func (r *RentStructure) MinRent(object NonEphemeralObject) uint64 {
	return uint64(r.VByteCost) * object.VBytes(r, nil)
}

// MinStorageDepositForReturnOutput returns the minimum renting costs for an BasicOutput which returns
// a StorageDepositReturnUnlockCondition amount back to the origin sender.
func (r *RentStructure) MinStorageDepositForReturnOutput(sender Address) uint64 {
	return uint64(r.VByteCost) * uint64((&BasicOutput{Conditions: UnlockConditions{&AddressUnlockCondition{Address: sender}}, Amount: 0}).VBytes(r, nil))
}

// NonEphemeralObject is an object which can not be pruned by nodes as it
// makes up an integral part to execute the IOTA protocol. This kind of objects are associated
// with costs in terms of the resources they take up.
type NonEphemeralObject interface {
	// VBytes returns the cost this object has in terms of taking up
	// virtual and physical space within the data set needed to implement the IOTA protocol.
	// The override parameter acts as an escape hatch in case the cost needs to be adjusted
	// according to some external properties outside the NonEphemeralObject.
	VBytes(rentStruct *RentStructure, override VBytesFunc) uint64
}

// VBytesFunc is a function which computes the virtual byte cost of a NonEphemeralObject.
type VBytesFunc func(rentStruct *RentStructure) uint64
