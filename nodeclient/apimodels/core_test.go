package apimodels_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
)

func testAPI() iotago.API {
	params := iotago.NewV3ProtocolParameters(
		iotago.WithTimeProviderOptions(time.Unix(1690879505, 0).UTC().Unix(), 10, 13),
	)

	return iotago.V3API(params)
}

func Test_InfoResponse(t *testing.T) {
	api := testAPI()
	{
		response := &apimodels.InfoResponse{
			Name:    "test",
			Version: "2.0.0",
			Status: &apimodels.InfoResNodeStatus{
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
			Metrics: &apimodels.InfoResNodeMetrics{
				BlocksPerSecond:          1.1,
				ConfirmedBlocksPerSecond: 2.2,
				ConfirmationRate:         3.3,
			},
			ProtocolParameters: []*apimodels.InfoResProtocolParameters{
				{
					StartEpoch: 0,
					Parameters: api.ProtocolParameters(),
				},
			},
			BaseToken: &apimodels.InfoResBaseToken{
				Name:            "Shimmer",
				TickerSymbol:    "SMR",
				Unit:            "SMR",
				Subunit:         "glow",
				Decimals:        6,
				UseMetricPrefix: false,
			},
			Features: []string{"test"},
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"name\":\"test\",\"version\":\"2.0.0\",\"status\":{\"isHealthy\":false,\"acceptedTangleTime\":\"1690879505000000000\",\"relativeAcceptedTangleTime\":\"1690879505000000000\",\"confirmedTangleTime\":\"1690879505000000000\",\"relativeConfirmedTangleTime\":\"1690879505000000000\",\"latestCommitmentId\":\"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"latestFinalizedSlot\":\"1\",\"latestAcceptedBlockSlot\":\"2\",\"latestConfirmedBlockSlot\":\"3\",\"pruningEpoch\":\"4\"},\"metrics\":{\"blocksPerSecond\":\"1.1E+00\",\"confirmedBlocksPerSecond\":\"2.2E+00\",\"confirmationRate\":\"3.3E+00\"},\"protocolParameters\":[{\"startEpoch\":\"0\",\"parameters\":{\"type\":0,\"version\":3,\"networkName\":\"testnet\",\"bech32Hrp\":\"rms\",\"rentStructure\":{\"vByteCost\":100,\"vByteFactorData\":1,\"vByteFactorKey\":10,\"vByteFactorBlockIssuerKey\":100,\"vByteFactorStakingFeature\":100,\"vByteFactorDelegation\":100},\"workScoreStructure\":{\"workScoreDataKibibyte\":0,\"workScoreBlock\":1,\"workScoreMissingParent\":0,\"workScoreInput\":0,\"workScoreContextInput\":0,\"workScoreOutput\":0,\"workScoreNativeToken\":0,\"workScoreStaking\":0,\"workScoreBlockIssuer\":0,\"workScoreAllotment\":0,\"workScoreSignatureEd25519\":0,\"workScoreMinStrongParentsThreshold\":0},\"tokenSupply\":\"1813620509061365\",\"genesisUnixTimestamp\":\"1690879505\",\"slotDurationInSeconds\":10,\"slotsPerEpochExponent\":13,\"manaStructure\":{\"manaBitsCount\":63,\"manaGenerationRate\":1,\"manaGenerationRateExponent\":17,\"manaDecayFactors\":[4291249941,4287535805,4283824883,4280117173,4276412671,4272711377,4269013285,4265318395,4261626702,4257938205,4254252900,4250570785,4246891856,4243216112,4239543550,4235874166,4232207957,4228544922,4224885058,4221228361,4217574829,4213924459,4210277249,4206633195,4202992295,4199354547,4195719947,4192088493,4188460182,4184835011,4181212978,4177594080,4173978314,4170365677,4166756168,4163149782,4159546518,4155946372,4152349343,4148755427,4145164621,4141576923,4137992331,4134410840,4130832450,4127257157,4123684959,4120115852,4116549834,4112986903,4109427055,4105870289,4102316601,4098765988,4095218449,4091673981,4088132580,4084594244,4081058971,4077526757,4073997601,4070471499,4066948449,4063428449,4059911495,4056397585,4052886716,4049378886,4045874092,4042372332,4038873602,4035377901,4031885225,4028395572,4024908939,4021425325,4017944725,4014467138,4010992560,4007520990,4004052425,4000586862,3997124298,3993664731,3990208159,3986754578,3983303986,3979856381,3976411760,3972970120,3969531459,3966095774,3962663063,3959233323,3955806551,3952382745,3948961903,3945544021,3942129098,3938717130,3935308116,3931902052,3928498936,3925098765,3921701537,3918307250,3914915900,3911527486,3908142004,3904759453,3901379829,3898003131,3894629355,3891258499,3887890560,3884525537,3881163426,3877804224,3874447931,3871094542,3867744056,3864396469,3861051780,3857709986,3854371084,3851035072,3847701948,3844371708,3841044351,3837719873,3834398273,3831079548,3827763695,3824450713,3821140597,3817833347,3814528959,3811227431,3807928760,3804632945,3801339982,3798049869,3794762604,3791478184,3788196607,3784917870,3781641970,3778368907,3775098676,3771831275,3768566702,3765304955,3762046031,3758789928,3755536643,3752286174,3749038518,3745793673,3742551636,3739312405,3736075978,3732842352,3729611525,3726383494,3723158258,3719935812,3716716156,3713499286,3710285201,3707073897,3703865373,3700659626,3697456653,3694256453,3691059023,3687864360,3684672462,3681483326,3678296951,3675113334,3671932472,3668754363,3665579005,3662406395,3659236531,3656069411,3652905032,3649743392,3646584488,3643428318,3640274880,3637124172,3633976190,3630830933,3627688398,3624548583,3621411486,3618277104,3615145434,3612016476,3608890225,3605766680,3602645839,3599527699,3596412257,3593299512,3590189461,3587082102,3583977433,3580875450,3577776153,3574679537,3571585602,3568494345,3565405764,3562319855,3559236618,3556156049,3553078146,3550002907,3546930330,3543860413,3540793152,3537728546,3534666593,3531607290,3528550634,3525496624,3522445258,3519396533,3516350446,3513306995,3510266179,3507227995,3504192440,3501159513,3498129210,3495101531,3492076472,3489054031,3486034206,3483016995,3480002395,3476990404,3473981020,3470974241,3467970065,3464968488,3461969510,3458973127,3455979337,3452988139,3449999530,3447013507,3444030069,3441049213,3438070937,3435095238,3432122115,3429151566,3426183587,3423218178,3420255335,3417295056,3414337339,3411382183,3408429584,3405479541,3402532051,3399587112,3396644722,3393704878,3390767579,3387832823,3384900606,3381970927,3379043784,3376119175,3373197097,3370277548,3367360525,3364446028,3361534053,3358624598,3355717662,3352813241,3349911335,3347011940,3344115054,3341220676,3338328803,3335439433,3332552563,3329668193,3326786318,3323906939,3321030051,3318155653,3315283743,3312414319,3309547378,3306682918,3303820938,3300961435,3298104407,3295249852,3292397767,3289548151,3286701001,3283856315,3281014092,3278174328,3275337023,3272502173,3269669777,3266839832,3264012336,3261187288,3258364685,3255544525,3252726806,3249911526,3247098682,3244288273,3241480296,3238674749,3235871631,3233070939,3230272671,3227476825,3224683399,3221892391,3219103798,3216317619,3213533851,3210752492,3207973541,3205196995,3202422853,3199651111,3196881768,3194114823,3191350272,3188588114,3185828346,3183070967,3180315975,3177563367,3174813142,3172065297,3169319830,3166576739,3163836023,3161097679,3158361705,3155628099,3152896859,3150167982,3147441468,3144717314,3141995517,3139276076,3136558989,3133844253,3131131867],\"manaDecayFactorsExponent\":32,\"manaDecayFactorEpochsSum\":2420916375,\"manaDecayFactorEpochsSumExponent\":21},\"stakingUnbondingPeriod\":\"10\",\"validationBlocksPerSlot\":10,\"punishmentEpochs\":\"10\",\"livenessThreshold\":\"3\",\"minCommittableAge\":\"10\",\"maxCommittableAge\":\"20\",\"epochNearingThreshold\":\"24\",\"congestionControlParameters\":{\"rmcMin\":\"1\",\"increase\":\"0\",\"decrease\":\"0\",\"increaseThreshold\":800000,\"decreaseThreshold\":500000,\"schedulerRate\":100000,\"minMana\":\"1\",\"maxBufferSize\":1000,\"maxValidationBufferSize\":100},\"versionSignaling\":{\"windowSize\":7,\"windowTargetRatio\":5,\"activationOffset\":7},\"rewardsParameters\":{\"validatorBlocksPerSlot\":10,\"profitMarginExponent\":8,\"bootstrappinDuration\":\"1154\",\"rewardsManaShareCoefficient\":\"2\",\"decayBalancingConstantExponent\":8,\"decayBalancingConstant\":\"1\",\"poolCoefficientExponent\":31}}}],\"baseToken\":{\"name\":\"Shimmer\",\"tickerSymbol\":\"SMR\",\"unit\":\"SMR\",\"subunit\":\"glow\",\"decimals\":6,\"useMetricPrefix\":false},\"features\":[\"test\"]}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.InfoResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))

		// ignore computed values
		require.EqualValues(t, response, decoded)
	}

	// Test omitempty
	{
		response := &apimodels.InfoResBaseToken{
			Name:            "IOTA",
			TickerSymbol:    "IOTA",
			Unit:            "MIOTA",
			UseMetricPrefix: true,
			// No Subunit
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"name\":\"IOTA\",\"tickerSymbol\":\"IOTA\",\"unit\":\"MIOTA\",\"decimals\":0,\"useMetricPrefix\":true}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.InfoResBaseToken)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}

func Test_IssuanceBlockHeaderResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.IssuanceBlockHeaderResponse{
		StrongParents: iotago.BlockIDs{
			iotago.BlockID{0x9},
		},
		WeakParents: iotago.BlockIDs{
			iotago.BlockID{0x8},
		},
		ShallowLikeParents: iotago.BlockIDs{
			iotago.BlockID{0x7},
		},
		LatestFinalizedSlot: 14,
		Commitment: &iotago.Commitment{
			ProtocolVersion:      api.Version(),
			Index:                18,
			PreviousCommitmentID: iotago.CommitmentID{0x1},
			RootsID:              iotago.Identifier{0x2},
			CumulativeWeight:     89,
			ReferenceManaCost:    123,
		},
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"strongParents\":[\"0x09000000000000000000000000000000000000000000000000000000000000000000000000000000\"],\"weakParents\":[\"0x08000000000000000000000000000000000000000000000000000000000000000000000000000000\"],\"shallowLikeParents\":[\"0x07000000000000000000000000000000000000000000000000000000000000000000000000000000\"],\"latestFinalizedSlot\":\"14\",\"commitment\":{\"protocolVersion\":3,\"index\":\"18\",\"previousCommitmentId\":\"0x01000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"rootsId\":\"0x0200000000000000000000000000000000000000000000000000000000000000\",\"cumulativeWeight\":\"89\",\"referenceManaCost\":\"123\"}}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.IssuanceBlockHeaderResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_BlockCreatedResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.BlockCreatedResponse{
		BlockID: iotago.BlockID{0x1},
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"blockId\":\"0x01000000000000000000000000000000000000000000000000000000000000000000000000000000\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.BlockCreatedResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_BlockMetadataResponse(t *testing.T) {
	api := testAPI()

	{
		response := &apimodels.BlockMetadataResponse{
			BlockID:            iotago.BlockID{0x9},
			BlockState:         apimodels.BlockStateFailed.String(),
			BlockFailureReason: apimodels.BlockFailureParentNotFound,
			TxState:            apimodels.TransactionStateFailed.String(),
			TxFailureReason:    apimodels.TxFailureFailedToClaimDelegationReward,
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"blockId\":\"0x09000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"blockState\":\"failed\",\"blockFailureReason\":3,\"txState\":\"failed\",\"txFailureReason\":21}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.BlockMetadataResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}

	// Test omitempty
	{
		response := &apimodels.BlockMetadataResponse{
			BlockID:    iotago.BlockID{0x9},
			BlockState: apimodels.BlockStateConfirmed.String(),
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"blockId\":\"0x09000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"blockState\":\"confirmed\"}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.BlockMetadataResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}

func Test_OutputMetadataResponse(t *testing.T) {
	api := testAPI()

	{
		response := &apimodels.OutputMetadataResponse{
			BlockID:              iotago.BlockID{0x8},
			TransactionID:        iotago.TransactionID{0x9},
			OutputIndex:          3,
			IsSpent:              true,
			CommitmentIDSpent:    iotago.CommitmentID{0x6},
			TransactionIDSpent:   iotago.TransactionID{0x1},
			IncludedCommitmentID: iotago.CommitmentID{0x3},
			LatestCommitmentID:   iotago.CommitmentID{0x2},
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"blockId\":\"0x08000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"transactionId\":\"0x0900000000000000000000000000000000000000000000000000000000000000\",\"outputIndex\":3,\"isSpent\":true,\"commitmentIdSpent\":\"0x06000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"transactionIdSpent\":\"0x0100000000000000000000000000000000000000000000000000000000000000\",\"includedCommitmentId\":\"0x03000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"latestCommitmentId\":\"0x02000000000000000000000000000000000000000000000000000000000000000000000000000000\"}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.OutputMetadataResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}

	// Test omitempty
	{
		response := &apimodels.OutputMetadataResponse{
			BlockID:            iotago.BlockID{0x8},
			TransactionID:      iotago.TransactionID{0x9},
			OutputIndex:        3,
			IsSpent:            false,
			LatestCommitmentID: iotago.CommitmentID{0x2},
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"blockId\":\"0x08000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"transactionId\":\"0x0900000000000000000000000000000000000000000000000000000000000000\",\"outputIndex\":3,\"isSpent\":false,\"latestCommitmentId\":\"0x02000000000000000000000000000000000000000000000000000000000000000000000000000000\"}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.OutputMetadataResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}

func Test_UTXOChangesResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.UTXOChangesResponse{
		Index: 42,
		CreatedOutputs: iotago.OutputIDs{
			iotago.OutputID{0x1},
		},
		ConsumedOutputs: iotago.OutputIDs{
			iotago.OutputID{0x2},
		},
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"index\":\"42\",\"createdOutputs\":[\"0x01000000000000000000000000000000000000000000000000000000000000000000\"],\"consumedOutputs\":[\"0x02000000000000000000000000000000000000000000000000000000000000000000\"]}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.UTXOChangesResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_CongestionResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.CongestionResponse{
		SlotIndex:            12,
		Ready:                true,
		ReferenceManaCost:    100,
		BlockIssuanceCredits: 80,
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"slotIndex\":\"12\",\"ready\":true,\"referenceManaCost\":\"100\",\"blockIssuanceCredits\":\"80\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.CongestionResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_AccountStakingListResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.ValidatorsResponse{
		Validators: []*apimodels.ValidatorResponse{
			{
				AccountID:                      iotago.AccountID{0xFF},
				StakingEpochEnd:                0,
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

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)
	expected := "{\"stakers\":[{\"accountId\":\"0xff00000000000000000000000000000000000000000000000000000000000000\",\"stakingEpochEnd\":\"0\",\"poolStake\":\"123\",\"validatorStake\":\"456\",\"fixedCost\":\"69\",\"active\":true,\"latestSupportedProtocolVersion\":9,\"latestSupportedProtocolHash\":\"0x0000000000000000000000000000000000000000000000000000000000000000\"}],\"pageSize\":50,\"cursor\":\"0,1\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.ValidatorsResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_ManaRewardsResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.ManaRewardsResponse{
		EpochStart: 123,
		EpochEnd:   133,
		Rewards:    456,
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"epochIndexStart\":\"123\",\"epochIndexEnd\":\"133\",\"rewards\":\"456\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.ManaRewardsResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}

func Test_CommitteeResponse(t *testing.T) {
	api := testAPI()

	response := &apimodels.CommitteeResponse{
		Committee: []*apimodels.CommitteeMemberResponse{
			{
				AccountID:      iotago.AccountID{0xFF},
				PoolStake:      456,
				ValidatorStake: 123,
				FixedCost:      789,
			},
		},
		TotalStake:          456,
		TotalValidatorStake: 123,
		EpochIndex:          872,
	}

	jsonResponse, err := api.JSONEncode(response)
	require.NoError(t, err)

	expected := "{\"committee\":[{\"accountId\":\"0xff00000000000000000000000000000000000000000000000000000000000000\",\"poolStake\":\"456\",\"validatorStake\":\"123\",\"fixedCost\":\"789\"}],\"totalStake\":\"456\",\"totalValidatorStake\":\"123\",\"epochIndex\":\"872\"}"
	require.Equal(t, expected, string(jsonResponse))

	decoded := new(apimodels.CommitteeResponse)
	require.NoError(t, api.JSONDecode(jsonResponse, decoded))
	require.EqualValues(t, response, decoded)
}
