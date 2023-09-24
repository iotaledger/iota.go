package iotago

type CongestionControlParameters struct {
	// MinReferenceManaCost is the minimum value of the reference Mana cost.
	MinReferenceManaCost Mana `serix:"0,mapKey=minReferenceManaCost"`
	// Increase is the increase step size of the reference Mana cost.
	Increase Mana `serix:"1,mapKey=increase"`
	// Decrease is the decrease step size of the reference Mana cost.
	Decrease Mana `serix:"2,mapKey=decrease"`
	// IncreaseThreshold is the threshold for increasing the reference Mana cost.
	// This value should be between 0 and SchedulerRate*SlotDurationInSeconds.
	IncreaseThreshold WorkScore `serix:"3,mapKey=increaseThreshold"`
	// DecreaseThreshold is the threshold for decreasing the reference Mana cost.
	// This value should be between 0 and SchedulerRate*SlotDurationInSeconds and must be less than or equal to IncreaseThreshold.
	DecreaseThreshold WorkScore `serix:"4,mapKey=decreaseThreshold"`
	// SchedulerRate is the rate at which the scheduler runs in workscore units per second.
	SchedulerRate WorkScore `serix:"5,mapKey=schedulerRate"`
	// MinMana is the minimum amount of Mana that an account must have to have a block scheduled.
	MinMana Mana `serix:"6,mapKey=minMana"`
	// MaxBufferSize is the maximum number of blocks in the DRR buffer.
	MaxBufferSize uint32 `serix:"7,mapKey=maxBufferSize"`
	// MaxValidaitonBufferSize is the maximum number of blocks in the validation buffer.
	MaxValidationBufferSize uint32 `serix:"8,mapKey=maxValidationBufferSize"`
}

func (c *CongestionControlParameters) Equals(other CongestionControlParameters) bool {
	return c.MinReferenceManaCost == other.MinReferenceManaCost &&
		c.Increase == other.Increase &&
		c.Decrease == other.Decrease &&
		c.IncreaseThreshold == other.IncreaseThreshold &&
		c.DecreaseThreshold == other.DecreaseThreshold &&
		c.SchedulerRate == other.SchedulerRate &&
		c.MinMana == other.MinMana &&
		c.MaxBufferSize == other.MaxBufferSize &&
		c.MaxValidationBufferSize == other.MaxValidationBufferSize
}
