package iotago

type RewardsParameters struct {
	// ValidatorBlocksPerSlot is the number of validation blocks that should be issued by a selected validator per slot during its epoch duties.
	ValidatorBlocksPerSlot uint8 `serix:"0,mapKey=validatorBlocksPerSlot"`
	// ProfitMarginExponent is used for shift operation for calculation of profit margin.
	ProfitMarginExponent uint8 `serix:"1,mapKey=profitMarginExponent"`
}

func (r RewardsParameters) Equals(other RewardsParameters) bool {
	return r.ValidatorBlocksPerSlot == other.ValidatorBlocksPerSlot &&
		r.ProfitMarginExponent == other.ProfitMarginExponent
}
