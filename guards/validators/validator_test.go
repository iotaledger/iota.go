package validators_test

import (
	"testing"

	"github.com/iotaledger/iota.go/guards/validators"
	"github.com/iotaledger/iota.go/legacy"
	"github.com/stretchr/testify/assert"
)

// TODO: convert to table driven tests, however, maybe the entire pkg gets deleted anyway

func TestValidateNonEmptyStrings(t *testing.T) {
	assert.NoError(t, validators.ValidateNonEmptyStrings(nil, "123", "321")())
	assert.Error(t, validators.ValidateNonEmptyStrings(legacy.ErrInvalidTrytes, []string{}...)())
}

func TestValidateHashes(t *testing.T) {
	assert.NoError(t, validators.ValidateHashes(
		"I9GLMICBJTETFVYUFIJRXTSANYQC9PZCCYREMLDYNJLYTR9LUEK9CAHKQZGLBGZRMVXBLP99EUHMZ9999",
		"YQSPGXZQUYXQUKDKBECDPLPBVJGHGMAHYWBKGCBOVBRBYOGWAENUMSMOQWFMP9KOWNTNOZUOGZOXVZZPABMVEOGLI9",
	)())

	assert.Error(t, validators.ValidateHashes(
		"I9GLMICBJTETFVYUFIJRXTSANYQC9PZCCYREMLDYNabcdLYTR9LUEK9CAHKQZGLBGZRMVXBLP99EUHMZ9999",
		"YQSPGXZQUYXQUKDKBECDPLPavcdMSMOQWFMP9KOWNTNOZUOGZOXVZZPABMVEOGLI9",
	)())
}

func TestValidateURIs(t *testing.T) {
	assert.NoError(t, validators.ValidateURIs(
		"tcp://example.com:14600",
		"udp://example.com",
	)())
}

func TestValidateSecurityLevel(t *testing.T) {
	assert.NoError(t, validators.ValidateSecurityLevel(legacy.SecurityLevelMedium)())
	assert.Error(t, validators.ValidateSecurityLevel(legacy.SecurityLevel(-1))())
	assert.Error(t, validators.ValidateSecurityLevel(legacy.SecurityLevel(4))())
	assert.Error(t, validators.ValidateSecurityLevel(legacy.SecurityLevel(0))())
	assert.Error(t, validators.ValidateSecurityLevel(legacy.SecurityLevel(-1))())
}

func TestValidateSeed(t *testing.T) {
	assert.NoError(t, validators.ValidateSeed("JEDDPBHQSSKN9TDZVDITVFHFZOGXKUHGUATPHLLVJCVOQFCAFRBJATLVZLPCHVUKTHATGANRCIETJRGBB")())
	assert.Error(t, validators.ValidateSeed("JED8PBHQS7KN9TDZVDIT2FHFZOGXKUHG3ATPHLLVJCVOQFCAFRBJATLVZLPCHVUKTHATGANRCIETJRGBB")())
}

func TestValidateStartEndOptions(t *testing.T) {
	assert.NoError(t, validators.ValidateStartEndOptions(0, nil)())
	var e uint64 = 100
	assert.NoError(t, validators.ValidateStartEndOptions(0, &e)())
	e = 500
	assert.NoError(t, validators.ValidateStartEndOptions(0, &e)())

	e = 2
	assert.Error(t, validators.ValidateStartEndOptions(5, &e)())
	e = 2000
	assert.Error(t, validators.ValidateStartEndOptions(0, &e)())
}
