package curl_test

import (
	"testing"

	"github.com/iotaledger/iota.go/legacy/curl"
	"github.com/iotaledger/iota.go/legacy/trinary"
)

func TestMustHashTrytes(t *testing.T) {
	type args struct {
		t      trinary.Trytes
		rounds []curl.CurlRounds
	}
	tests := []struct {
		name string
		args args
		want trinary.Trytes
	}{
		{
			name: "normal trytes",
			args: args{
				t:      "A",
				rounds: []curl.CurlRounds{curl.CurlP81},
			},
			want: "TJVKPMTAMIZVBVHIVQUPTKEMPROEKV9SB9COEDQYRHYPTYSKQIAN9PQKMZHCPO9TS9BHCORFKW9CQXZEE",
		},
		{
			name: "normal trytes #2",
			args: args{
				t:      "B",
				rounds: nil,
			},
			want: "QFZXTJUJNLAOSZKXXMMGJJLFACVLRQMRBKOJLMTZXPLPVDSWWWXLBX9CDZWHMDMSDMDQKXQGEWPC9BJHN",
		},
		{
			name: "normal trytes #3",
			args: args{
				t:      "ABCDEFGHIJ",
				rounds: []curl.CurlRounds{curl.CurlP81},
			},
			want: "JKSGOZW9WFTALAYESGNJYRGCKIMZSVBMFIIHYBFCUCSLWDI9EEPTZBLGWNPJOMW9HZWNOFGBR9RNHKCYI",
		},
		{
			name: "empty trytes - P81",
			args: args{
				t:      "",
				rounds: []curl.CurlRounds{curl.CurlP81},
			},
			want: "999999999999999999999999999999999999999999999999999999999999999999999999999999999",
		}, {
			name: "P27",
			args: args{
				t:      "TWENTYSEVEN",
				rounds: []curl.CurlRounds{curl.CurlP27},
			},
			want: "RQPYXJPRXEEPLYLAHWTTFRXXUZTV9SZPEVOQ9FZATCXJOZLZ9A9BFXTUBSHGXN9OOA9GWIPGAAWEDVNPN",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := curl.MustHashTrytes(tt.args.t, tt.args.rounds...); got != tt.want {
				t.Errorf("MustHashTrytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
