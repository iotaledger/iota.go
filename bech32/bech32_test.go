package bech32

import (
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/iotaledger/iota.go/v2/bech32/internal/base32"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	var tests = []*struct {
		hrp    string
		src    []byte
		expS   string
		expErr error
	}{
		{
			hrp:  "A",
			src:  []byte{},
			expS: "A12UEL5L",
		},
		{
			hrp:  "a",
			src:  []byte{},
			expS: "a12uel5l",
		},
		{
			hrp:  "an83characterlonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio",
			src:  []byte{},
			expS: "an83characterlonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1tt5tgs",
		},
		{
			hrp:  "abcdef",
			src:  decodeHex("00443214c74254b635cf84653a56d7c675be77df"),
			expS: "abcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw",
		},
		{
			hrp:  "1",
			src:  make([]byte, 51),
			expS: "11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqc8247j",
		},
		{
			hrp:  "split",
			src:  decodeHex("c5f38b70305f519bf66d85fb6cf03058f3dde463ecd7918f2dc743918f2d"),
			expS: "split1checkupstagehandshakeupstreamerranterredcaperred2y9e3w",
		},
		{
			hrp:  "?",
			src:  []byte{},
			expS: "?1ezyfcl",
		},
		{
			hrp:  "test",
			src:  decodeHex("daf4d8dd12769fcc87f7ad04a6495b"),
			expS: "test1mt6d3hgjw60ueplh45z2vj2mnsrk9a",
		},
		{
			hrp:  "test",
			src:  decodeHex("ff"),
			expS: "test1lu0zy72x",
		},
		{
			hrp:  "bc",
			src:  []byte{},
			expS: "bc1gmk9yu",
		},
		{
			hrp:  "tb",
			src:  decodeHex("00c318a1e0a628b34025e8c9019ab6d09b64c2b3c66a693d0dc63194b024819310"),
			expS: "tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sl5k7",
		},
		{
			hrp:  "ca",
			src:  decodeHex("030102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			expS: "ca1qvqsyqcyq5rqwzqfpg9scrgwpugpzysnzs23v9ccrydpk8qarc0jqxuzx4s",
		},
		{hrp: "\x20", src: []byte{}, expErr: ErrInvalidCharacter},
		{hrp: "\x7F", src: []byte{}, expErr: ErrInvalidCharacter},
		{hrp: "\x80", src: []byte{}, expErr: ErrInvalidCharacter},
		{hrp: "an84characterslonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio", src: []byte{}, expErr: ErrInvalidLength},
		{hrp: "", src: decodeHex("ff"), expErr: ErrInvalidLength},
		{hrp: "bC", src: []byte{}, expErr: ErrMixedCase},
		{hrp: "Cb", src: []byte{}, expErr: ErrMixedCase},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.hrp, tt.src), func(t *testing.T) {
			s, err := Encode(tt.hrp, tt.src)
			if assert.Truef(t, errors.Is(err, tt.expErr), "unexpected error: %v", err) {
				assert.Equal(t, tt.expS, s)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	var tests = []*struct {
		s       string
		expHRP  string
		expData []byte
		expErr  error
	}{
		{
			s:       "A12UEL5L",
			expHRP:  "a",
			expData: []byte{},
		},
		{
			s:       "a12uel5l",
			expHRP:  "a",
			expData: []byte{},
		},
		{
			s:       "an83characterlonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1tt5tgs",
			expHRP:  "an83characterlonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio",
			expData: []byte{},
		},
		{
			s:       "abcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw",
			expHRP:  "abcdef",
			expData: decodeHex("00443214c74254b635cf84653a56d7c675be77df"),
		},
		{
			s:       "11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqc8247j",
			expHRP:  "1",
			expData: make([]byte, 51),
		},
		{
			s:       "split1checkupstagehandshakeupstreamerranterredcaperred2y9e3w",
			expHRP:  "split",
			expData: decodeHex("c5f38b70305f519bf66d85fb6cf03058f3dde463ecd7918f2dc743918f2d"),
		},
		{
			s:       "?1ezyfcl",
			expHRP:  "?",
			expData: []byte{},
		},
		{
			s:       "test1mt6d3hgjw60ueplh45z2vj2mnsrk9a",
			expHRP:  "test",
			expData: decodeHex("daf4d8dd12769fcc87f7ad04a6495b"),
		},
		{
			s:       "test1lu0zy72x",
			expHRP:  "test",
			expData: decodeHex("ff"),
		},
		{
			s:       "bc1gmk9yu",
			expHRP:  "bc",
			expData: []byte{},
		},
		{
			s:       "tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sl5k7",
			expHRP:  "tb",
			expData: decodeHex("00c318a1e0a628b34025e8c9019ab6d09b64c2b3c66a693d0dc63194b024819310"),
		},
		{
			s:       "ca1qvqsyqcyq5rqwzqfpg9scrgwpugpzysnzs23v9ccrydpk8qarc0jqxuzx4s",
			expHRP:  "ca",
			expData: decodeHex("030102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
		},
		{s: "\x201nwldj5", expErr: ErrInvalidCharacter},
		{s: "\x7F1axkwrx", expErr: ErrInvalidCharacter},
		{s: "\x801eym55h", expErr: ErrInvalidCharacter},
		{s: "an84characterslonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1569pvx", expErr: ErrInvalidLength},
		{s: "pzry9x0s0muk", expErr: ErrMissingSeparator},
		{s: "1pzry9x0s0muk", expErr: ErrInvalidSeparator},
		{s: "x1b4n0q5v", expErr: ErrInvalidCharacter},
		{s: "li1dgmt3", expErr: ErrInvalidChecksum},
		{s: "A1G7SGD8", expErr: ErrInvalidChecksum},
		{s: "10a06t8", expErr: ErrInvalidSeparator},
		{s: "1qzzfhee", expErr: ErrInvalidSeparator},
		{s: "tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sL5k7", expErr: ErrMixedCase},
		{s: "tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3pjxtptv", expErr: base32.ErrNonZeroPadding},
		{s: "bC1gmk9yu", expErr: ErrMixedCase},
		{s: "Cb1gmk9yu", expErr: ErrMixedCase},
		{s: "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4", expErr: base32.ErrInvalidLength},
		{s: "test1ls7uz56", expErr: base32.ErrInvalidLength},
		{s: "test1lllxt840c", expErr: base32.ErrInvalidLength},
		{s: "test1llllllkmgrnu", expErr: base32.ErrInvalidLength},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			hrp, data, err := Decode(tt.s)
			if assert.Truef(t, errors.Is(err, tt.expErr), "unexpected error: %v", err) {
				assert.Equal(t, tt.expHRP, hrp)
				assert.Equal(t, tt.expData, data)
			}
		})
	}
}

func decodeHex(s string) []byte {
	dst, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return dst
}
