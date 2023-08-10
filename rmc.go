package iotago

type RMCParameters struct {
	// RMCMin is the minimum value of the reference Mana cost.
	RMCMin Mana `serix:"0,mapKey=rmcMin"`
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
}

func (r *RMCParameters) Equals(other RMCParameters) bool {
	return r.RMCMin == other.RMCMin &&
		r.Increase == other.Increase &&
		r.Decrease == other.Decrease &&
		r.IncreaseThreshold == other.IncreaseThreshold &&
		r.DecreaseThreshold == other.DecreaseThreshold
}
