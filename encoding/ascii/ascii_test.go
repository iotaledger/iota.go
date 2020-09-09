package ascii_test

import (
	"fmt"
	"testing"

	"github.com/iotaledger/iota.go/encoding/ascii"
	"github.com/iotaledger/iota.go/trinary"
)

func TestEncodeToTrytes(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    trinary.Trytes
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				src: "IOTA",
			},
			want:    "SBYBCCKB",
			wantErr: false,
		},
		{
			name: "invalid input",
			args: args{
				src: "Γιώτα",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ascii.EncodeToTrytes(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeToTrytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EncodeToTrytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleEncodeToTrytes() {
	encodedAscii, err := ascii.EncodeToTrytes("IOTA")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(encodedAscii)
	// Output: SBYBCCKB
}

func TestDecodeTrytes(t *testing.T) {
	type args struct {
		src trinary.Trytes
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				src: "SBYBCCKB",
			},
			want:    "IOTA",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ascii.DecodeTrytes(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeTrytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DecodeTrytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleDecodeTrytes() {
	decodedTrytes, err := ascii.DecodeTrytes("SBYBCCKB")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(decodedTrytes)
	// Output: IOTA
}
