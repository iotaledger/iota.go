package iotago

type CongestionControlParameters struct {
	// MinReferenceManaCost is the minimum value of the reference Mana cost.
	MinReferenceManaCost Mana `serix:""`
	// Increase is the increase step size of the reference Mana cost.
	Increase Mana `serix:""`
	// Decrease is the decrease step size of the reference Mana cost.
	Decrease Mana `serix:""`
	// IncreaseThreshold is the threshold for increasing the reference Mana cost.
	// This value should be between 0 and SchedulerRate*SlotDurationInSeconds.
	IncreaseThreshold WorkScore `serix:""`
	// DecreaseThreshold is the threshold for decreasing the reference Mana cost.
	// This value should be between 0 and SchedulerRate*SlotDurationInSeconds and must be less than or equal to IncreaseThreshold.
	DecreaseThreshold WorkScore `serix:""`
	// SchedulerRate is the rate at which the scheduler runs in workscore units per second.
	SchedulerRate WorkScore `serix:""`
	// MaxBufferSize is the maximum number of blocks in the DRR buffer.
	MaxBufferSize uint32 `serix:""`
	// MaxValidaitonBufferSize is the maximum number of blocks in the validation buffer.
	MaxValidationBufferSize uint32 `serix:""`
}

func (c *CongestionControlParameters) Equals(other CongestionControlParameters) bool {
	return c.MinReferenceManaCost == other.MinReferenceManaCost &&
		c.Increase == other.Increase &&
		c.Decrease == other.Decrease &&
		c.IncreaseThreshold == other.IncreaseThreshold &&
		c.DecreaseThreshold == other.DecreaseThreshold &&
		c.SchedulerRate == other.SchedulerRate &&
		c.MaxBufferSize == other.MaxBufferSize &&
		c.MaxValidationBufferSize == other.MaxValidationBufferSize
}
