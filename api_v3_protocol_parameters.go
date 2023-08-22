package iotago

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/runtime/options"
)

// V3ProtocolParameters defines the parameters of the protocol.
type V3ProtocolParameters struct {
	basicProtocolParameters `serix:"0"`
}

func NewV3ProtocolParameters(opts ...options.Option[V3ProtocolParameters]) *V3ProtocolParameters {
	var schedulerRate WorkScore = 100000
	return options.Apply(
		new(V3ProtocolParameters),
		append([]options.Option[V3ProtocolParameters]{
			WithVersion(apiV3Version),
			WithNetworkOptions("testnet", PrefixTestnet),
			WithSupplyOptions(1813620509061365, 100, 1, 10, 100, 100),
			WithWorkScoreOptions(0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0),
			WithTimeProviderOptions(time.Now().Unix(), 10, 13),
			WithManaOptions(1,
				17,
				[]uint32{
					4291249941, 4287535805, 4283824883, 4280117173, 4276412671, 4272711377, 4269013285, 4265318395, 4261626702, 4257938205, 4254252900, 4250570785, 4246891856, 4243216112, 4239543550, 4235874166, 4232207957, 4228544922, 4224885058, 4221228361, 4217574829, 4213924459, 4210277249, 4206633195, 4202992295, 4199354547, 4195719947, 4192088493, 4188460182, 4184835011, 4181212978, 4177594080, 4173978314, 4170365677, 4166756168, 4163149782, 4159546518, 4155946372, 4152349343, 4148755427, 4145164621, 4141576923, 4137992331, 4134410840, 4130832450, 4127257157, 4123684959, 4120115852, 4116549834, 4112986903, 4109427055, 4105870289, 4102316601, 4098765988, 4095218449, 4091673981, 4088132580, 4084594244, 4081058971, 4077526757, 4073997601, 4070471499, 4066948449, 4063428449, 4059911495, 4056397585, 4052886716, 4049378886, 4045874092, 4042372332, 4038873602, 4035377901, 4031885225, 4028395572, 4024908939, 4021425325, 4017944725, 4014467138, 4010992560, 4007520990, 4004052425, 4000586862, 3997124298, 3993664731, 3990208159, 3986754578, 3983303986, 3979856381, 3976411760, 3972970120, 3969531459, 3966095774, 3962663063, 3959233323, 3955806551, 3952382745, 3948961903, 3945544021, 3942129098, 3938717130, 3935308116, 3931902052, 3928498936, 3925098765, 3921701537, 3918307250, 3914915900, 3911527486, 3908142004, 3904759453, 3901379829, 3898003131, 3894629355, 3891258499, 3887890560, 3884525537, 3881163426, 3877804224, 3874447931, 3871094542, 3867744056, 3864396469, 3861051780, 3857709986, 3854371084, 3851035072, 3847701948, 3844371708, 3841044351, 3837719873, 3834398273, 3831079548, 3827763695, 3824450713, 3821140597, 3817833347, 3814528959, 3811227431, 3807928760, 3804632945, 3801339982, 3798049869, 3794762604, 3791478184, 3788196607, 3784917870, 3781641970, 3778368907, 3775098676, 3771831275, 3768566702, 3765304955, 3762046031, 3758789928, 3755536643, 3752286174, 3749038518, 3745793673, 3742551636, 3739312405, 3736075978, 3732842352, 3729611525, 3726383494, 3723158258, 3719935812, 3716716156, 3713499286, 3710285201, 3707073897, 3703865373, 3700659626, 3697456653, 3694256453, 3691059023, 3687864360, 3684672462, 3681483326, 3678296951, 3675113334, 3671932472, 3668754363, 3665579005, 3662406395, 3659236531, 3656069411, 3652905032, 3649743392, 3646584488, 3643428318, 3640274880, 3637124172, 3633976190, 3630830933, 3627688398, 3624548583, 3621411486, 3618277104, 3615145434, 3612016476, 3608890225, 3605766680, 3602645839, 3599527699, 3596412257, 3593299512, 3590189461, 3587082102, 3583977433, 3580875450, 3577776153, 3574679537, 3571585602, 3568494345, 3565405764, 3562319855, 3559236618, 3556156049, 3553078146, 3550002907, 3546930330, 3543860413, 3540793152, 3537728546, 3534666593, 3531607290, 3528550634, 3525496624, 3522445258, 3519396533, 3516350446, 3513306995, 3510266179, 3507227995, 3504192440, 3501159513, 3498129210, 3495101531, 3492076472, 3489054031, 3486034206, 3483016995, 3480002395, 3476990404, 3473981020, 3470974241, 3467970065, 3464968488, 3461969510, 3458973127, 3455979337, 3452988139, 3449999530, 3447013507, 3444030069, 3441049213, 3438070937, 3435095238, 3432122115, 3429151566, 3426183587, 3423218178, 3420255335, 3417295056, 3414337339, 3411382183, 3408429584, 3405479541, 3402532051, 3399587112, 3396644722, 3393704878, 3390767579, 3387832823, 3384900606, 3381970927, 3379043784, 3376119175, 3373197097, 3370277548, 3367360525, 3364446028, 3361534053, 3358624598, 3355717662, 3352813241, 3349911335, 3347011940, 3344115054, 3341220676, 3338328803, 3335439433, 3332552563, 3329668193, 3326786318, 3323906939, 3321030051, 3318155653, 3315283743, 3312414319, 3309547378, 3306682918, 3303820938, 3300961435, 3298104407, 3295249852, 3292397767, 3289548151, 3286701001, 3283856315, 3281014092, 3278174328, 3275337023, 3272502173, 3269669777, 3266839832, 3264012336, 3261187288, 3258364685, 3255544525, 3252726806, 3249911526, 3247098682, 3244288273, 3241480296, 3238674749, 3235871631, 3233070939, 3230272671, 3227476825, 3224683399, 3221892391, 3219103798, 3216317619, 3213533851, 3210752492, 3207973541, 3205196995, 3202422853, 3199651111, 3196881768, 3194114823, 3191350272, 3188588114, 3185828346, 3183070967, 3180315975, 3177563367, 3174813142, 3172065297, 3169319830, 3166576739, 3163836023, 3161097679, 3158361705, 3155628099, 3152896859, 3150167982, 3147441468, 3144717314, 3141995517, 3139276076, 3136558989, 3133844253, 3131131867,
				},
				32,
				2420916375,
				21,
			),
			WithLivenessOptions(3, 10, 20, 24),
			WithCongestionControlOptions(1, 0, 0, 8*schedulerRate, 5*schedulerRate, schedulerRate, 1, 100*MaxBlockSize),
			WithStakingOptions(10),
			WithVersionSignalingOptions(7, 5, 7),
			WithRewardsOptions(10, 8, 8, 31, 1154, 2, 1),
		},
			opts...,
		),
	)
}

var _ ProtocolParameters = &V3ProtocolParameters{}

func (p *V3ProtocolParameters) Version() Version {
	return p.basicProtocolParameters.Version
}

func (p *V3ProtocolParameters) Bech32HRP() NetworkPrefix {
	return p.basicProtocolParameters.Bech32HRP
}

func (p *V3ProtocolParameters) NetworkName() string {
	return p.basicProtocolParameters.NetworkName
}

func (p *V3ProtocolParameters) RentStructure() *RentStructure {
	return &p.basicProtocolParameters.RentStructure
}

func (p *V3ProtocolParameters) WorkScoreStructure() *WorkScoreStructure {
	return &p.basicProtocolParameters.WorkScoreStructure
}

func (p *V3ProtocolParameters) TokenSupply() BaseToken {
	return p.basicProtocolParameters.TokenSupply
}

func (p *V3ProtocolParameters) NetworkID() NetworkID {
	return NetworkIDFromString(p.basicProtocolParameters.NetworkName)
}

func (p *V3ProtocolParameters) SlotsPerEpochExponent() uint8 {
	return p.basicProtocolParameters.SlotsPerEpochExponent
}

func (p *V3ProtocolParameters) TimeProvider() *TimeProvider {
	return NewTimeProvider(p.basicProtocolParameters.GenesisUnixTimestamp, int64(p.basicProtocolParameters.SlotDurationInSeconds), p.basicProtocolParameters.SlotsPerEpochExponent)
}

// ParamEpochDurationInSlots defines the amount of slots in an epoch.
func (p *V3ProtocolParameters) ParamEpochDurationInSlots() SlotIndex {
	return 1 << p.basicProtocolParameters.SlotsPerEpochExponent
}

func (p *V3ProtocolParameters) StakingUnbondingPeriod() EpochIndex {
	return p.basicProtocolParameters.StakingUnbondingPeriod
}

func (p *V3ProtocolParameters) LivenessThreshold() SlotIndex {
	return p.basicProtocolParameters.LivenessThreshold
}

func (p *V3ProtocolParameters) MinCommittableAge() SlotIndex {
	return p.basicProtocolParameters.MinCommittableAge
}

func (p *V3ProtocolParameters) MaxCommittableAge() SlotIndex {
	return p.basicProtocolParameters.MaxCommittableAge
}

func (p *V3ProtocolParameters) EpochNearingThreshold() SlotIndex {
	return p.basicProtocolParameters.EpochNearingThreshold
}

func (p *V3ProtocolParameters) CongestionControlParameters() *CongestionControlParameters {
	return &p.basicProtocolParameters.CongestionControlParameters
}

func (p *V3ProtocolParameters) ManaParameters() *ManaParameters {
	return &p.basicProtocolParameters.ManaParameters
}

func (p *V3ProtocolParameters) VersionSignaling() *VersionSignaling {
	return &p.basicProtocolParameters.VersionSignaling
}

func (p *V3ProtocolParameters) RewardsParameters() *RewardsParameters {
	return &p.basicProtocolParameters.RewardsParameters
}

func (p *V3ProtocolParameters) Bytes() ([]byte, error) {
	return CommonSerixAPI().Encode(context.TODO(), p)
}

func (p *V3ProtocolParameters) Hash() (Identifier, error) {
	bytes, err := p.Bytes()
	if err != nil {
		return Identifier{}, err
	}

	return IdentifierFromData(bytes), nil
}

func (p *V3ProtocolParameters) String() string {
	return fmt.Sprintf("ProtocolParameters: {\n\tVersion: %d\n\tNetwork Name: %s\n\tBech32 HRP Prefix: %s\n"+
		"\tRent Structure: %v\n\tWorkScore Structure: %v\n\tToken Supply: %d\n\tGenesis Unix Timestamp: %d\n"+
		"\tSlot Duration in Seconds: %d\n\tSlots per Epoch Exponent: %d\n\tMana Generation Rate: %d\n"+
		"\tMana Generation Rate Exponent: %d\t\nMana Decay Factors: %v\n\tMana Decay Factors Exponent: %d\n"+
		"\tMana Decay Factor Epochs Sum: %d\n\tMana Decay Factor Epochs Sum Exponent: %d\n\tStaking Unbonding Period: %d\n"+
		"\tLiveness Threshold: %d\n\tMin Committable Age: %d\n\tMax Committable Age: %d\n}"+
		"\tEpoch Nearing Threshold: %d\n\tRMC parameters: %v\n\tVersion Signaling: %v\n\tRewardsParameters: %v\n",
		p.basicProtocolParameters.Version, p.basicProtocolParameters.NetworkName, p.basicProtocolParameters.Bech32HRP,
		p.basicProtocolParameters.RentStructure, p.basicProtocolParameters.WorkScoreStructure, p.basicProtocolParameters.TokenSupply, p.basicProtocolParameters.GenesisUnixTimestamp,
		p.basicProtocolParameters.SlotDurationInSeconds, p.basicProtocolParameters.SlotsPerEpochExponent, p.basicProtocolParameters.ManaParameters.ManaGenerationRate,
		p.basicProtocolParameters.ManaParameters.ManaGenerationRateExponent, p.basicProtocolParameters.ManaParameters.ManaDecayFactors, p.basicProtocolParameters.ManaParameters.ManaDecayFactorsExponent,
		p.basicProtocolParameters.ManaParameters.ManaDecayFactorEpochsSum, p.basicProtocolParameters.ManaParameters.ManaDecayFactorEpochsSumExponent, p.basicProtocolParameters.StakingUnbondingPeriod,
		p.basicProtocolParameters.LivenessThreshold, p.basicProtocolParameters.MinCommittableAge, p.basicProtocolParameters.MaxCommittableAge,
		p.basicProtocolParameters.EpochNearingThreshold, p.basicProtocolParameters.CongestionControlParameters, p.basicProtocolParameters.VersionSignaling, p.basicProtocolParameters.RewardsParameters)
}

func (p *V3ProtocolParameters) ManaDecayProvider() *ManaDecayProvider {
	return NewManaDecayProvider(p.TimeProvider(), p.basicProtocolParameters.SlotsPerEpochExponent, p.basicProtocolParameters.ManaParameters.ManaGenerationRate, p.basicProtocolParameters.ManaParameters.ManaGenerationRateExponent, p.basicProtocolParameters.ManaParameters.ManaDecayFactors, p.basicProtocolParameters.ManaParameters.ManaDecayFactorsExponent, p.basicProtocolParameters.ManaParameters.ManaDecayFactorEpochsSum, p.basicProtocolParameters.ManaParameters.ManaDecayFactorEpochsSumExponent)
}

func (p *V3ProtocolParameters) Equals(other ProtocolParameters) bool {
	otherV3Params, matches := other.(*V3ProtocolParameters)
	if !matches {
		return false
	}

	return p.basicProtocolParameters.Equals(otherV3Params.basicProtocolParameters)
}

func WithVersion(version Version) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.Version = version
	}
}

func WithNetworkOptions(networkName string, bech32HRP NetworkPrefix) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.NetworkName = networkName
		p.basicProtocolParameters.Bech32HRP = bech32HRP
	}
}

func WithSupplyOptions(totalSupply BaseToken, vByteCost uint32, vBFactorData, vBFactorKey, vBFactorIssuerKeys, vBFactorStakingFeature VByteCostFactor) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.TokenSupply = totalSupply
		p.basicProtocolParameters.RentStructure = RentStructure{
			VByteCost:              vByteCost,
			VBFactorData:           vBFactorData,
			VBFactorKey:            vBFactorKey,
			VBFactorIssuerKeys:     vBFactorIssuerKeys,
			VBFactorStakingFeature: vBFactorStakingFeature,
		}
	}
}

func WithWorkScoreOptions(
	dataKilobyte WorkScore,
	block WorkScore,
	missingParent WorkScore,
	input WorkScore,
	contextInput WorkScore,
	output WorkScore,
	nativeToken WorkScore,
	staking WorkScore,
	blockIssuer WorkScore,
	allotment WorkScore,
	signatureEd25519 WorkScore,
	minStrongParentsThreshold byte,
) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.WorkScoreStructure = WorkScoreStructure{
			DataKilobyte:              dataKilobyte,
			Block:                     block,
			MissingParent:             missingParent,
			Input:                     input,
			ContextInput:              contextInput,
			Output:                    output,
			NativeToken:               nativeToken,
			Staking:                   staking,
			BlockIssuer:               blockIssuer,
			Allotment:                 allotment,
			SignatureEd25519:          signatureEd25519,
			MinStrongParentsThreshold: minStrongParentsThreshold,
		}
	}
}

func WithTimeProviderOptions(genesisTimestamp int64, slotDuration uint8, slotsPerEpochExponent uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.GenesisUnixTimestamp = genesisTimestamp
		p.basicProtocolParameters.SlotDurationInSeconds = slotDuration
		p.basicProtocolParameters.SlotsPerEpochExponent = slotsPerEpochExponent
	}
}

func WithManaOptions(manaGenerationRate uint8, manaGenerationRateExponent uint8, manaDecayFactors []uint32, manaDecayFactorsExponent uint8, manaDecayFactorEpochsSum uint32, manaDecayFactorEpochsSumExponent uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.ManaParameters.ManaGenerationRate = manaGenerationRate
		p.basicProtocolParameters.ManaParameters.ManaGenerationRateExponent = manaGenerationRateExponent
		p.basicProtocolParameters.ManaParameters.ManaDecayFactors = manaDecayFactors
		p.basicProtocolParameters.ManaParameters.ManaDecayFactorsExponent = manaDecayFactorsExponent
		p.basicProtocolParameters.ManaParameters.ManaDecayFactorEpochsSum = manaDecayFactorEpochsSum
		p.basicProtocolParameters.ManaParameters.ManaDecayFactorEpochsSumExponent = manaDecayFactorEpochsSumExponent
	}
}

func WithLivenessOptions(livenessThreshold SlotIndex, minCommittableAge SlotIndex, maxCommittableAge SlotIndex, epochNearingThreshold SlotIndex) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.LivenessThreshold = livenessThreshold
		p.basicProtocolParameters.MinCommittableAge = minCommittableAge
		p.basicProtocolParameters.MaxCommittableAge = maxCommittableAge
		p.basicProtocolParameters.EpochNearingThreshold = epochNearingThreshold
	}
}

func WithCongestionControlOptions(rmcMin Mana, rmcIncrease Mana, rmcDecrease Mana, rmcIncreaseThreshold WorkScore, rmcDecreaseThreshold WorkScore, schedulerRate WorkScore, minMana Mana, maxBufferSize uint32) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.CongestionControlParameters.RMCMin = rmcMin
		p.basicProtocolParameters.CongestionControlParameters.Increase = rmcIncrease
		p.basicProtocolParameters.CongestionControlParameters.Decrease = rmcDecrease
		p.basicProtocolParameters.CongestionControlParameters.IncreaseThreshold = rmcIncreaseThreshold
		p.basicProtocolParameters.CongestionControlParameters.DecreaseThreshold = rmcDecreaseThreshold
		p.basicProtocolParameters.CongestionControlParameters.SchedulerRate = schedulerRate
		p.basicProtocolParameters.CongestionControlParameters.MinMana = minMana
		p.basicProtocolParameters.CongestionControlParameters.MaxBufferSize = maxBufferSize
	}
}

func WithStakingOptions(unbondingPeriod EpochIndex) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.StakingUnbondingPeriod = unbondingPeriod
	}
}

func WithVersionSignalingOptions(windowSize uint8, windowTargetRatio uint8, activationOffset uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.VersionSignaling = VersionSignaling{
			WindowSize:        windowSize,
			WindowTargetRatio: windowTargetRatio,
			ActivationOffset:  activationOffset,
		}
	}
}
