package iotago

import "math"

// TODO: fixed point arithmetic
type DecayProvider struct {
	storedManaDecayFactors    map[SlotIndex]float64
	potentialManaDecayFactors map[SlotIndex]float64 // mana generation rate is built into these values
}

func NewDecayProvider(stored map[SlotIndex]float64, potential map[SlotIndex]float64) *DecayProvider {
	return &DecayProvider{
		storedManaDecayFactors:    stored,
		potentialManaDecayFactors: potential,
	}
}

func (d *DecayProvider) StoredManaDecayFactor(deltaT SlotIndex) float64 {
	// TODO: implement decay factor table
	totalDecay := 1.0
	for i, factor := range d.storedManaDecayFactors {
		totalDecay *= math.Floor(float64(deltaT)/float64(i)) * factor
	}
	return totalDecay
}

func (d *DecayProvider) PotentialManaDecayFactor(deltaT SlotIndex) float64 {
	// This factor incorporates generation of Mana and the decay to be applied to IOTA token amount
	// TODO: implement decay factor table for this
	totalDecay := 1.0
	for i, factor := range d.potentialManaDecayFactors {
		totalDecay *= math.Floor(float64(deltaT)/float64(i)) * factor
	}
	return totalDecay
}

func (d *DecayProvider) StoredManaWithDecay(storedMana uint64, deltaT SlotIndex) uint64 {
	// TODO: implement fixed point arithmetic for applying decay factors
	return uint64(float64(storedMana) * d.StoredManaDecayFactor(deltaT))
}

func (d *DecayProvider) PotentialManaWithDecay(deposit uint64, deltaT SlotIndex) uint64 {
	// TODO: implement fixed point arithmetic for applying decay factors
	return uint64(float64(deposit) * d.PotentialManaDecayFactor(deltaT))
}
