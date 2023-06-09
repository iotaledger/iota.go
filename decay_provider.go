package iotago

// TODO: fixed point arithmetic
type DecayProvider struct {
	manaGenerationRate        uint8
	storedManaDecayFactors    []float64
	potentialManaDecayFactors []float64 // mana generation rate is built into these values
}

func NewDecayProvider(manaGenerationRate uint8, stored []float64, potential []float64) *DecayProvider {
	return &DecayProvider{
		manaGenerationRate:        manaGenerationRate,
		storedManaDecayFactors:    stored,
		potentialManaDecayFactors: potential,
	}
}

func (d *DecayProvider) StoredManaDecayFactor(deltaT SlotIndex) float64 {
	// TODO: implement decay factor table
	totalDecay := 1.0
	for i := len(d.storedManaDecayFactors) - 1; i >= 0; i-- {
		totalDecay *= float64(int(deltaT)/i) * d.storedManaDecayFactors[i]
		deltaT -= SlotIndex(int(deltaT) / i)
	}
	return 1.0
}

func (d *DecayProvider) PotentialManaDecayFactor(deltaT SlotIndex) float64 {
	// This factor incorporates generation of Mana and the decay to be applied to IOTA token amount
	// TODO: implement decay factor table for this
	totalDecay := 1.0
	for i := len(d.potentialManaDecayFactors) - 1; i >= 0; i-- {
		totalDecay *= float64(int(deltaT)/i) * d.potentialManaDecayFactors[i]
		deltaT -= SlotIndex(int(deltaT) / i)
	}
	return float64(uint64(deltaT) * uint64(d.manaGenerationRate))
}

func (d *DecayProvider) StoredManaWithDecay(storedMana uint64, deltaT SlotIndex) uint64 {
	// TODO: implement fixed point arithmetic for applying decay factors
	return uint64(float64(storedMana) * d.StoredManaDecayFactor(deltaT))
}

func (d *DecayProvider) PotentialManaWithDecay(deposit uint64, deltaT SlotIndex) uint64 {
	// TODO: implement fixed point arithmetic for applying decay factors
	return uint64(float64(deposit) * d.PotentialManaDecayFactor(deltaT))
}
