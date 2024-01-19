//nolint:scopelint
package iotago_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

const (
	OneIOTA iotago.BaseToken = 1_000_000
)

type deSerializeTest struct {
	name      string
	source    any
	target    any
	seriErr   error
	deSeriErr error
}

func (test *deSerializeTest) assertBinaryEncodeDecode(t *testing.T) {
	t.Helper()

	serixData, err := tpkg.ZeroCostTestAPI.Encode(test.source, serix.WithValidation())
	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr, "binary encoding")

		// Encode again without validation so we can check that deserialization would also fail.
		serixData, err = tpkg.ZeroCostTestAPI.Encode(test.source)
		require.NoError(t, err, "binary encoding")
	} else {
		require.NoError(t, err, "binary encoding")
	}

	if src, ok := test.source.(iotago.Sizer); ok {
		require.Equal(t, src.Size(), len(serixData), "binary encoding")
	}

	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := tpkg.ZeroCostTestAPI.Decode(serixData, serixTarget, serix.WithValidation())
	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr, "binary decoding")

		return
	}
	require.NoError(t, err, "binary decoding")
	require.Len(t, serixData, bytesRead, "binary decoding")
	require.EqualValues(t, test.source, serixTarget, "binary decoding")
}

func (test *deSerializeTest) assertJSONEncodeDecode(t *testing.T) {
	t.Helper()

	sourceJSON, err := tpkg.ZeroCostTestAPI.JSONEncode(test.source, serix.WithValidation())
	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr, "JSON encoding")

		// Encode again without validation so we can check that deserialization would also fail.
		sourceJSON, err = tpkg.ZeroCostTestAPI.JSONEncode(test.source)
		require.NoError(t, err, "JSON encoding")
	} else {
		require.NoError(t, err, "JSON encoding")
	}

	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	err = tpkg.ZeroCostTestAPI.JSONDecode(sourceJSON, jsonDest, serix.WithValidation())
	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr, "JSON decoding")

		return
	}
	require.NoError(t, err, "JSON decoding")
	require.EqualValues(t, test.source, jsonDest, "JSON decoding")
}

func (test *deSerializeTest) deSerialize(t *testing.T) {
	t.Helper()

	if reflect.TypeOf(test.target).Kind() != reflect.Ptr {
		// This is required for the serixTarget creation hack to work.
		t.Fatal("test target must be a pointer")
	}

	test.assertBinaryEncodeDecode(t)
	test.assertJSONEncodeDecode(t)
}

func TestProtocolParameters_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandProtocolParameters(),
			target: &iotago.V3ProtocolParameters{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestProtocolParametersJSONMarshalling(t *testing.T) {
	var protoParams = tpkg.FixedGenesisV3TestProtocolParameters

	protoParamsJSON := `{"type":0,"version":3,"networkName":"testnet","bech32Hrp":"rms","storageScoreParameters":{"storageCost":"100","factorData":1,"offsetOutputOverhead":"10","offsetEd25519BlockIssuerKey":"100","offsetStakingFeature":"100","offsetDelegation":"100"},"workScoreParameters":{"dataByte":0,"block":1,"input":0,"contextInput":0,"output":0,"nativeToken":0,"staking":0,"blockIssuer":0,"allotment":0,"signatureEd25519":0},"manaParameters":{"bitsCount":63,"generationRate":1,"generationRateExponent":17,"decayFactors":[4290989755,4287015898,4283045721,4279079221,4275116394,4271157237,4267201747,4263249920,4259301752,4255357241,4251416383,4247479175,4243545613,4239615693,4235689414,4231766770,4227847759,4223932377,4220020622,4216112489,4212207975,4208307077,4204409792,4200516116,4196626046,4192739579,4188856710,4184977438,4181101758,4177229668,4173361163,4169496241,4165634898,4161777132,4157922938,4154072313,4150225254,4146381758,4142541822,4138705441,4134872614,4131043336,4127217604,4123395415,4119576766,4115761654,4111950074,4108142024,4104337501,4100536502,4096739022,4092945060,4089154610,4085367672,4081584240,4077804312,4074027884,4070254954,4066485518,4062719573,4058957115,4055198142,4051442650,4047690636,4043942097,4040197029,4036455429,4032717295,4028982622,4025251408,4021523650,4017799344,4014078486,4010361075,4006647106,4002936577,3999229484,3995525824,3991825594,3988128791,3984435412,3980745453,3977058911,3973375783,3969696066,3966019757,3962346853,3958677350,3955011245,3951348535,3947689218,3944033289,3940380746,3936731586,3933085805,3929443400,3925804369,3922168708,3918536413,3914907483,3911281913,3907659701,3904040843,3900425337,3896813179,3893204366,3889598896,3885996764,3882397968,3878802505,3875210372,3871621566,3868036083,3864453920,3860875075,3857299544,3853727325,3850158414,3846592808,3843030504,3839471499,3835915790,3832363374,3828814248,3825268408,3821725853,3818186578,3814650580,3811117858,3807588407,3804062225,3800539308,3797019654,3793503259,3789990121,3786480237,3782973602,3779470216,3775970074,3772473173,3768979511,3765489084,3762001889,3758517924,3755037186,3751559671,3748085377,3744614300,3741146437,3737681787,3734220344,3730762108,3727307074,3723855240,3720406602,3716961158,3713518905,3710079840,3706643960,3703211262,3699781742,3696355399,3692932229,3689512229,3686095396,3682681728,3679271221,3675863872,3672459679,3669058639,3665660748,3662266004,3658874404,3655485944,3652100623,3648718437,3645339383,3641963459,3638590661,3635220986,3631854432,3628490996,3625130675,3621773465,3618419365,3615068371,3611720480,3608375690,3605033997,3601695399,3598359893,3595027476,3591698145,3588371897,3585048730,3581728640,3578411625,3575097682,3571786808,3568479000,3565174255,3561872571,3558573944,3555278373,3551985853,3548696383,3545409959,3542126578,3538846238,3535568936,3532294669,3529023435,3525755230,3522490051,3519227897,3515968763,3512712648,3509459548,3506209461,3502962384,3499718314,3496477248,3493239183,3490004118,3486772048,3483542972,3480316886,3477093788,3473873674,3470656543,3467442391,3464231216,3461023014,3457817784,3454615522,3451416225,3448219892,3445026518,3441836102,3438648641,3435464131,3432282571,3429103957,3425928286,3422755557,3419585766,3416418910,3413254987,3410093995,3406935929,3403780789,3400628570,3397479270,3394332887,3391189418,3388048860,3384911211,3381776467,3378644627,3375515686,3372389644,3369266496,3366146241,3363028875,3359914396,3356802802,3353694089,3350588256,3347485298,3344385214,3341288001,3338193657,3335102178,3332013562,3328927806,3325844909,3322764866,3319687675,3316613335,3313541841,3310473192,3307407385,3304344417,3301284286,3298226988,3295172522,3292120885,3289072074,3286026086,3282982919,3279942570,3276905037,3273870317,3270838408,3267809306,3264783010,3261759516,3258738822,3255720926,3252705824,3249693515,3246683996,3243677263,3240673315,3237672149,3234673763,3231678153,3228685317,3225695253,3222707958,3219723430,3216741666,3213762662,3210786418,3207812930,3204842196,3201874213,3198908979,3195946490,3192986746,3190029742,3187075477,3184123947,3181175151,3178229086,3175285749,3172345138,3169407251,3166472084,3163539635,3160609902,3157682882,3154758573,3151836972,3148918077,3146001885,3143088393,3140177600,3137269503,3134364098,3131461384,3128561359,3125664019,3122769362,3119877387,3116988089,3114101467,3111217518,3108336240,3105457631,3102581687,3099708407,3096837788,3093969827,3091104522,3088241871,3085381870,3082524519,3079669813,3076817752,3073968331,3071121550,3068277404,3065435893,3062597013,3059760763,3056927139,3054096139,3051267761,3048442002,3045618860,3042798333,3039980417,3037165112,3034352413,3031542320,3028734829,3025929938,3023127644,3020327946,3017530840,3014736325,3011944398,3009155056],"decayFactorsExponent":32,"decayFactorEpochsSum":2262417561,"decayFactorEpochsSumExponent":21,"annualDecayFactorPercentage":70},"tokenSupply":"1813620509061365","genesisSlot":65898,"genesisUnixTimestamp":"1690879505","slotDurationInSeconds":10,"slotsPerEpochExponent":13,"stakingUnbondingPeriod":10,"validationBlocksPerSlot":10,"punishmentEpochs":10,"livenessThresholdLowerBound":15,"livenessThresholdUpperBound":30,"minCommittableAge":10,"maxCommittableAge":20,"epochNearingThreshold":60,"congestionControlParameters":{"minReferenceManaCost":"1","increase":"0","decrease":"0","increaseThreshold":800000,"decreaseThreshold":500000,"schedulerRate":100000,"maxBufferSize":1000,"maxValidationBufferSize":100},"versionSignalingParameters":{"windowSize":7,"windowTargetRatio":5,"activationOffset":7},"rewardsParameters":{"profitMarginExponent":8,"bootstrappingDuration":1079,"manaShareCoefficient":"2","decayBalancingConstantExponent":8,"decayBalancingConstant":"1","poolCoefficientExponent":11,"retentionPeriod":384},"targetCommitteeSize":32,"chainSwitchingThreshold":3}`

	jsonProtoParams, err := tpkg.ZeroCostTestAPI.JSONEncode(protoParams)
	require.NoError(t, err)
	require.Equal(t, protoParamsJSON, string(jsonProtoParams))

	var decodedProtoParams iotago.ProtocolParameters
	err = tpkg.ZeroCostTestAPI.JSONDecode([]byte(protoParamsJSON), &decodedProtoParams)
	require.NoError(t, err)

	require.Equal(t, protoParams, decodedProtoParams)
}
