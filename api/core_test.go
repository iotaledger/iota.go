package api_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func testAPI() iotago.API {
	params := tpkg.FixedGenesisV3TestProtocolParameters

	return iotago.V3API(params)
}

func Test_InfoResponse(t *testing.T) {
	testAPI := testAPI()
	{
		response := &api.InfoResponse{
			Name:    "test",
			Version: "2.0.0",
			Status: &api.InfoResNodeStatus{
				IsHealthy:                   false,
				AcceptedTangleTime:          time.Unix(1690879505, 0).UTC(),
				RelativeAcceptedTangleTime:  time.Unix(1690879505, 0).UTC(),
				ConfirmedTangleTime:         time.Unix(1690879505, 0).UTC(),
				RelativeConfirmedTangleTime: time.Unix(1690879505, 0).UTC(),
				LatestCommitmentID:          iotago.CommitmentID{},
				LatestFinalizedSlot:         1,
				LatestAcceptedBlockSlot:     2,
				LatestConfirmedBlockSlot:    3,
				PruningEpoch:                4,
			},
			Metrics: &api.InfoResNodeMetrics{
				BlocksPerSecond:          1.1,
				ConfirmedBlocksPerSecond: 2.2,
				ConfirmationRate:         3.3,
			},
			ProtocolParameters: []*api.InfoResProtocolParameters{
				{
					StartEpoch: 0,
					Parameters: testAPI.ProtocolParameters(),
				},
			},
			BaseToken: &api.InfoResBaseToken{
				Name:         "Shimmer",
				TickerSymbol: "SMR",
				Unit:         "SMR",
				Subunit:      "glow",
				Decimals:     6,
			},
			Features: []string{"test"},
		}

		jsonResponse, err := testAPI.JSONEncode(response)
		require.NoError(t, err)

		expected := `{"name":"test","version":"2.0.0","status":{"isHealthy":false,"acceptedTangleTime":"1690879505000000000","relativeAcceptedTangleTime":"1690879505000000000","confirmedTangleTime":"1690879505000000000","relativeConfirmedTangleTime":"1690879505000000000","latestCommitmentId":"0x000000000000000000000000000000000000000000000000000000000000000000000000","latestFinalizedSlot":1,"latestAcceptedBlockSlot":2,"latestConfirmedBlockSlot":3,"pruningEpoch":4},"metrics":{"blocksPerSecond":"1.1E+00","confirmedBlocksPerSecond":"2.2E+00","confirmationRate":"3.3E+00"},"protocolParameters":[{"startEpoch":0,"parameters":{"type":0,"version":3,"networkName":"testnet","bech32Hrp":"rms","storageScoreParameters":{"storageCost":"100","factorData":1,"offsetOutputOverhead":"10","offsetEd25519BlockIssuerKey":"100","offsetStakingFeature":"100","offsetDelegation":"100"},"workScoreParameters":{"dataByte":1,"block":100,"input":10,"contextInput":20,"output":20,"nativeToken":20,"staking":5000,"blockIssuer":1000,"allotment":1000,"signatureEd25519":1000},"manaParameters":{"bitsCount":63,"generationRate":1,"generationRateExponent":17,"decayFactors":[4290989755,4287015898,4283045721,4279079221,4275116394,4271157237,4267201747,4263249920,4259301752,4255357241,4251416383,4247479175,4243545613,4239615693,4235689414,4231766770,4227847759,4223932377,4220020622,4216112489,4212207975,4208307077,4204409792,4200516116,4196626046,4192739579,4188856710,4184977438,4181101758,4177229668,4173361163,4169496241,4165634898,4161777132,4157922938,4154072313,4150225254,4146381758,4142541822,4138705441,4134872614,4131043336,4127217604,4123395415,4119576766,4115761654,4111950074,4108142024,4104337501,4100536502,4096739022,4092945060,4089154610,4085367672,4081584240,4077804312,4074027884,4070254954,4066485518,4062719573,4058957115,4055198142,4051442650,4047690636,4043942097,4040197029,4036455429,4032717295,4028982622,4025251408,4021523650,4017799344,4014078486,4010361075,4006647106,4002936577,3999229484,3995525824,3991825594,3988128791,3984435412,3980745453,3977058911,3973375783,3969696066,3966019757,3962346853,3958677350,3955011245,3951348535,3947689218,3944033289,3940380746,3936731586,3933085805,3929443400,3925804369,3922168708,3918536413,3914907483,3911281913,3907659701,3904040843,3900425337,3896813179,3893204366,3889598896,3885996764,3882397968,3878802505,3875210372,3871621566,3868036083,3864453920,3860875075,3857299544,3853727325,3850158414,3846592808,3843030504,3839471499,3835915790,3832363374,3828814248,3825268408,3821725853,3818186578,3814650580,3811117858,3807588407,3804062225,3800539308,3797019654,3793503259,3789990121,3786480237,3782973602,3779470216,3775970074,3772473173,3768979511,3765489084,3762001889,3758517924,3755037186,3751559671,3748085377,3744614300,3741146437,3737681787,3734220344,3730762108,3727307074,3723855240,3720406602,3716961158,3713518905,3710079840,3706643960,3703211262,3699781742,3696355399,3692932229,3689512229,3686095396,3682681728,3679271221,3675863872,3672459679,3669058639,3665660748,3662266004,3658874404,3655485944,3652100623,3648718437,3645339383,3641963459,3638590661,3635220986,3631854432,3628490996,3625130675,3621773465,3618419365,3615068371,3611720480,3608375690,3605033997,3601695399,3598359893,3595027476,3591698145,3588371897,3585048730,3581728640,3578411625,3575097682,3571786808,3568479000,3565174255,3561872571,3558573944,3555278373,3551985853,3548696383,3545409959,3542126578,3538846238,3535568936,3532294669,3529023435,3525755230,3522490051,3519227897,3515968763,3512712648,3509459548,3506209461,3502962384,3499718314,3496477248,3493239183,3490004118,3486772048,3483542972,3480316886,3477093788,3473873674,3470656543,3467442391,3464231216,3461023014,3457817784,3454615522,3451416225,3448219892,3445026518,3441836102,3438648641,3435464131,3432282571,3429103957,3425928286,3422755557,3419585766,3416418910,3413254987,3410093995,3406935929,3403780789,3400628570,3397479270,3394332887,3391189418,3388048860,3384911211,3381776467,3378644627,3375515686,3372389644,3369266496,3366146241,3363028875,3359914396,3356802802,3353694089,3350588256,3347485298,3344385214,3341288001,3338193657,3335102178,3332013562,3328927806,3325844909,3322764866,3319687675,3316613335,3313541841,3310473192,3307407385,3304344417,3301284286,3298226988,3295172522,3292120885,3289072074,3286026086,3282982919,3279942570,3276905037,3273870317,3270838408,3267809306,3264783010,3261759516,3258738822,3255720926,3252705824,3249693515,3246683996,3243677263,3240673315,3237672149,3234673763,3231678153,3228685317,3225695253,3222707958,3219723430,3216741666,3213762662,3210786418,3207812930,3204842196,3201874213,3198908979,3195946490,3192986746,3190029742,3187075477,3184123947,3181175151,3178229086,3175285749,3172345138,3169407251,3166472084,3163539635,3160609902,3157682882,3154758573,3151836972,3148918077,3146001885,3143088393,3140177600,3137269503,3134364098,3131461384,3128561359,3125664019,3122769362,3119877387,3116988089,3114101467,3111217518,3108336240,3105457631,3102581687,3099708407,3096837788,3093969827,3091104522,3088241871,3085381870,3082524519,3079669813,3076817752,3073968331,3071121550,3068277404,3065435893,3062597013,3059760763,3056927139,3054096139,3051267761,3048442002,3045618860,3042798333,3039980417,3037165112,3034352413,3031542320,3028734829,3025929938,3023127644,3020327946,3017530840,3014736325,3011944398,3009155056],"decayFactorsExponent":32,"decayFactorEpochsSum":2262417561,"decayFactorEpochsSumExponent":21,"annualDecayFactorPercentage":70},"tokenSupply":"1813620509061365","genesisSlot":65898,"genesisUnixTimestamp":"1690879505","slotDurationInSeconds":10,"slotsPerEpochExponent":13,"stakingUnbondingPeriod":10,"validationBlocksPerSlot":10,"punishmentEpochs":10,"livenessThresholdLowerBound":15,"livenessThresholdUpperBound":30,"minCommittableAge":10,"maxCommittableAge":20,"epochNearingThreshold":60,"congestionControlParameters":{"minReferenceManaCost":"1","increase":"10","decrease":"10","increaseThreshold":800000,"decreaseThreshold":500000,"schedulerRate":100000,"maxBufferSize":1000,"maxValidationBufferSize":100},"versionSignalingParameters":{"windowSize":7,"windowTargetRatio":5,"activationOffset":7},"rewardsParameters":{"profitMarginExponent":8,"bootstrappingDuration":1079,"manaShareCoefficient":"2","decayBalancingConstantExponent":8,"decayBalancingConstant":"1","poolCoefficientExponent":11,"retentionPeriod":384},"targetCommitteeSize":32,"chainSwitchingThreshold":3}}],"baseToken":{"name":"Shimmer","tickerSymbol":"SMR","unit":"SMR","subunit":"glow","decimals":6},"features":["test"]}`
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(api.InfoResponse)
		require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))

		// ignore computed values
		require.EqualValues(t, response, decoded)
	}

	// Test omitempty
	{
		response := &api.InfoResBaseToken{
			Name:         "IOTA",
			TickerSymbol: "IOTA",
			Unit:         "MIOTA",
			// No Subunit
		}

		jsonResponse, err := testAPI.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"name\":\"IOTA\",\"tickerSymbol\":\"IOTA\",\"unit\":\"MIOTA\",\"decimals\":0}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(api.InfoResBaseToken)
		require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}

func Test_IssuanceBlockHeaderResponse(t *testing.T) {
	testAPI := testAPI()

	response := &api.IssuanceBlockHeaderResponse{
		StrongParents: iotago.BlockIDs{
			iotago.BlockID{0x9},
		},
		WeakParents: iotago.BlockIDs{
			iotago.BlockID{0x8},
		},
		ShallowLikeParents: iotago.BlockIDs{
			iotago.BlockID{0x7},
		},
		LatestParentBlockIssuingTime: time.Unix(1690879505, 0).UTC(),
		LatestFinalizedSlot:          14,
		LatestCommitment: &iotago.Commitment{
			ProtocolVersion:      testAPI.Version(),
			Slot:                 18,
			PreviousCommitmentID: iotago.CommitmentID{0x1},
			RootsID:              iotago.Identifier{0x2},
			CumulativeWeight:     89,
			ReferenceManaCost:    123,
		},
	}

	jsonResponse, err := testAPI.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"strongParents\":[\"0x090000000000000000000000000000000000000000000000000000000000000000000000\"],\"weakParents\":[\"0x080000000000000000000000000000000000000000000000000000000000000000000000\"],\"shallowLikeParents\":[\"0x070000000000000000000000000000000000000000000000000000000000000000000000\"],\"latestParentBlockIssuingTime\":\"1690879505000000000\",\"latestFinalizedSlot\":14,\"latestCommitment\":{\"protocolVersion\":3,\"slot\":18,\"previousCommitmentId\":\"0x010000000000000000000000000000000000000000000000000000000000000000000000\",\"rootsId\":\"0x0200000000000000000000000000000000000000000000000000000000000000\",\"cumulativeWeight\":\"89\",\"referenceManaCost\":\"123\"}}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(api.IssuanceBlockHeaderResponse)
	require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_BlockCreatedResponse(t *testing.T) {
	testAPI := testAPI()

	response := &api.BlockCreatedResponse{
		BlockID: iotago.BlockID{0x1},
	}

	jsonResponse, err := testAPI.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"blockId\":\"0x010000000000000000000000000000000000000000000000000000000000000000000000\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(api.BlockCreatedResponse)
	require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_BlockMetadataResponse(t *testing.T) {
	testAPI := testAPI()

	{
		response := &api.BlockMetadataResponse{
			BlockID:            iotago.BlockID{0x9},
			BlockState:         api.BlockStateFailed,
			BlockFailureReason: api.BlockFailureParentNotFound,
			TransactionMetadata: &api.TransactionMetadataResponse{
				TransactionID:            iotago.TransactionID{0x1},
				TransactionState:         api.TransactionStateFailed,
				TransactionFailureReason: api.TxFailureFailedToClaimDelegationReward,
			},
		}

		jsonResponse, err := testAPI.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"blockId\":\"0x090000000000000000000000000000000000000000000000000000000000000000000000\",\"blockState\":\"failed\",\"blockFailureReason\":3,\"transactionMetadata\":{\"transactionId\":\"0x010000000000000000000000000000000000000000000000000000000000000000000000\",\"transactionState\":\"failed\",\"transactionFailureReason\":20}}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(api.BlockMetadataResponse)
		require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}

	// Test omitempty
	{
		response := &api.BlockMetadataResponse{
			BlockID:    iotago.BlockID{0x9},
			BlockState: api.BlockStateConfirmed,
		}

		jsonResponse, err := testAPI.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"blockId\":\"0x090000000000000000000000000000000000000000000000000000000000000000000000\",\"blockState\":\"confirmed\"}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(api.BlockMetadataResponse)
		require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}

func Test_OutputMetadataResponse(t *testing.T) {
	testAPI := testAPI()

	{
		response := &api.OutputMetadata{
			OutputID: iotago.OutputID{0x01},
			BlockID:  iotago.BlockID{0x02},
			Included: &api.OutputInclusionMetadata{
				Slot:          3,
				TransactionID: iotago.TransactionID{0x4},
				CommitmentID:  iotago.CommitmentID{0x5},
			},
			Spent: &api.OutputConsumptionMetadata{
				Slot:          6,
				TransactionID: iotago.TransactionID{0x7},
				CommitmentID:  iotago.CommitmentID{0x8},
			},
			LatestCommitmentID: iotago.CommitmentID{0x9},
		}

		jsonResponse, err := testAPI.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"outputId\":\"0x0100000000000000000000000000000000000000000000000000000000000000000000000000\",\"blockId\":\"0x020000000000000000000000000000000000000000000000000000000000000000000000\",\"included\":{\"slot\":3,\"transactionId\":\"0x040000000000000000000000000000000000000000000000000000000000000000000000\",\"commitmentId\":\"0x050000000000000000000000000000000000000000000000000000000000000000000000\"},\"spent\":{\"slot\":6,\"transactionId\":\"0x070000000000000000000000000000000000000000000000000000000000000000000000\",\"commitmentId\":\"0x080000000000000000000000000000000000000000000000000000000000000000000000\"},\"latestCommitmentId\":\"0x090000000000000000000000000000000000000000000000000000000000000000000000\"}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(api.OutputMetadata)
		require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}

	// Test omitempty
	{
		response := &api.OutputMetadata{
			OutputID: iotago.OutputID{0x01},
			BlockID:  iotago.BlockID{0x02},
			Included: &api.OutputInclusionMetadata{
				Slot:          3,
				TransactionID: iotago.TransactionID{0x4},
				// CommitmentID is omitted
			},
			// Spent is omitted
			LatestCommitmentID: iotago.CommitmentID{0x9},
		}

		jsonResponse, err := testAPI.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"outputId\":\"0x0100000000000000000000000000000000000000000000000000000000000000000000000000\",\"blockId\":\"0x020000000000000000000000000000000000000000000000000000000000000000000000\",\"included\":{\"slot\":3,\"transactionId\":\"0x040000000000000000000000000000000000000000000000000000000000000000000000\"},\"latestCommitmentId\":\"0x090000000000000000000000000000000000000000000000000000000000000000000000\"}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(api.OutputMetadata)
		require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}

func Test_UTXOChangesResponse(t *testing.T) {
	testAPI := testAPI()

	commitmentID := iotago.NewCommitmentID(42, iotago.Identifier{})

	response := &api.UTXOChangesResponse{
		CommitmentID: commitmentID,
		CreatedOutputs: iotago.OutputIDs{
			iotago.OutputID{0x1},
		},
		ConsumedOutputs: iotago.OutputIDs{
			iotago.OutputID{0x2},
		},
	}

	jsonResponse, err := testAPI.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"commitmentId\":\"0x00000000000000000000000000000000000000000000000000000000000000002a000000\",\"createdOutputs\":[\"0x0100000000000000000000000000000000000000000000000000000000000000000000000000\"],\"consumedOutputs\":[\"0x0200000000000000000000000000000000000000000000000000000000000000000000000000\"]}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(api.UTXOChangesResponse)
	require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_UTXOChangesFullResponse(t *testing.T) {
	testAPI := testAPI()

	commitmentID := iotago.NewCommitmentID(42, iotago.Identifier{})

	response := &api.UTXOChangesFullResponse{
		CommitmentID: commitmentID,
		CreatedOutputs: []*api.OutputWithID{
			{
				OutputID: iotago.OutputID{0x1},
				Output: &iotago.BasicOutput{
					Amount: 123,
					Mana:   456,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: &iotago.Ed25519Address{0x01},
						},
					},
					Features: iotago.BasicOutputFeatures{},
				},
			},
		},
		ConsumedOutputs: []*api.OutputWithID{
			{
				OutputID: iotago.OutputID{0x2},
				Output: &iotago.BasicOutput{
					Amount: 456,
					Mana:   123,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: &iotago.Ed25519Address{0x02},
						},
					},
					Features: iotago.BasicOutputFeatures{},
				},
			},
		},
	}

	jsonResponse, err := testAPI.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"commitmentId\":\"0x00000000000000000000000000000000000000000000000000000000000000002a000000\",\"createdOutputs\":[{\"outputId\":\"0x0100000000000000000000000000000000000000000000000000000000000000000000000000\",\"output\":{\"type\":0,\"amount\":\"123\",\"mana\":\"456\",\"unlockConditions\":[{\"type\":0,\"address\":{\"type\":0,\"pubKeyHash\":\"0x0100000000000000000000000000000000000000000000000000000000000000\"}}]}}],\"consumedOutputs\":[{\"outputId\":\"0x0200000000000000000000000000000000000000000000000000000000000000000000000000\",\"output\":{\"type\":0,\"amount\":\"456\",\"mana\":\"123\",\"unlockConditions\":[{\"type\":0,\"address\":{\"type\":0,\"pubKeyHash\":\"0x0200000000000000000000000000000000000000000000000000000000000000\"}}]}}]}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(api.UTXOChangesFullResponse)
	require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_CongestionResponse(t *testing.T) {
	testAPI := testAPI()

	response := &api.CongestionResponse{
		Slot:                 12,
		Ready:                true,
		ReferenceManaCost:    100,
		BlockIssuanceCredits: 80,
	}

	jsonResponse, err := testAPI.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"slot\":12,\"ready\":true,\"referenceManaCost\":\"100\",\"blockIssuanceCredits\":\"80\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(api.CongestionResponse)
	require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_AccountStakingListResponse(t *testing.T) {
	testAPI := testAPI()

	response := &api.ValidatorsResponse{
		Validators: []*api.ValidatorResponse{
			{
				AddressBech32:                  iotago.AccountID{0xFF}.ToAddress().Bech32(iotago.PrefixTestnet),
				StakingEndEpoch:                0,
				PoolStake:                      123,
				ValidatorStake:                 456,
				FixedCost:                      69,
				Active:                         true,
				LatestSupportedProtocolVersion: 9,
			},
		},
		Cursor:   "0,1",
		PageSize: 50,
	}

	jsonResponse, err := testAPI.JSONEncode(response)
	require.NoError(t, err)
	expected := "{\"stakers\":[{\"address\":\"rms1prlsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqcyz9fx\",\"stakingEndEpoch\":0,\"poolStake\":\"123\",\"validatorStake\":\"456\",\"fixedCost\":\"69\",\"active\":true,\"latestSupportedProtocolVersion\":9,\"latestSupportedProtocolHash\":\"0x0000000000000000000000000000000000000000000000000000000000000000\"}],\"pageSize\":50,\"cursor\":\"0,1\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(api.ValidatorsResponse)
	require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_ManaRewardsResponse(t *testing.T) {
	testAPI := testAPI()

	response := &api.ManaRewardsResponse{
		StartEpoch:                      123,
		EndEpoch:                        133,
		Rewards:                         456,
		LatestCommittedEpochPoolRewards: 555,
	}

	jsonResponse, err := testAPI.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"startEpoch\":123,\"endEpoch\":133,\"rewards\":\"456\",\"latestCommittedEpochPoolRewards\":\"555\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(api.ManaRewardsResponse)
	require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_CommitteeResponse(t *testing.T) {
	testAPI := testAPI()

	response := &api.CommitteeResponse{
		Committee: []*api.CommitteeMemberResponse{
			{
				AddressBech32:  iotago.AccountID{0xFF}.ToAddress().Bech32(iotago.PrefixTestnet),
				PoolStake:      456,
				ValidatorStake: 123,
				FixedCost:      789,
			},
		},
		TotalStake:          456,
		TotalValidatorStake: 123,
		Epoch:               872,
	}

	jsonResponse, err := testAPI.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"committee\":[{\"address\":\"rms1prlsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqcyz9fx\",\"poolStake\":\"456\",\"validatorStake\":\"123\",\"fixedCost\":\"789\"}],\"totalStake\":\"456\",\"totalValidatorStake\":\"123\",\"epoch\":872}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(api.CommitteeResponse)
	require.NoError(t, testAPI.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}
