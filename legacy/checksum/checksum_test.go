package checksum_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/checksum"
	"github.com/iotaledger/iota.go/legacy/trinary"
)

var addrs = []trinary.Trytes{
	"GEXLJVJNKFPRGZSTOEVTODEEUJDQCFWOSLVBVVMTRVESTCCTPKILEADWUGMMZVUG9YTJSKNYQUNCSCBDY",
	"JSQITWMZVFUMJONOPSG9TG9SAWXODGVRHCZLYLGNYBUOWIBLDILU9FYGONNPSEQVLLJWB9D9IYCLTXJND",
	"STDNQP9USJOEZDFIMDMIRUVHDFDFUGJSTCDZGJFBBFOQTDPZAMLBYPWFXPDHWDLBUAKWLGTFQZWEFYKEB",
}

var checksums = []trinary.Trytes{"WDUCWPRQW", "JEADZTWUW", "DBKEDYNQC"}

var addrsWithChecksums []trinary.Trytes

func TestMain(m *testing.M) {
	addrsWithChecksums = make([]trinary.Trytes, 3)
	for i := range addrsWithChecksums {
		addrsWithChecksums[i] = addrs[i] + checksums[i]
	}
	os.Exit(m.Run())
}

func TestAddChecksum(t *testing.T) {
	type args struct {
		input          trinary.Trytes
		isAddress      bool
		checksumLength uint64
	}
	tests := []struct {
		name    string
		args    args
		want    trinary.Trytes
		wantErr bool
	}{
		{
			name: "ok - address #1",
			args: args{
				input:          addrs[0],
				isAddress:      true,
				checksumLength: legacy.AddressChecksumTrytesSize,
			},
			want:    addrs[0] + checksums[0],
			wantErr: false,
		},
		{
			name: "ok - address #2",
			args: args{
				input:          addrs[1],
				isAddress:      true,
				checksumLength: legacy.AddressChecksumTrytesSize,
			},
			want:    addrs[1] + checksums[1],
			wantErr: false,
		},
		{
			name: "ok - address #3",
			args: args{
				input:          addrs[2],
				isAddress:      true,
				checksumLength: legacy.AddressChecksumTrytesSize,
			},
			want:    addrs[2] + checksums[2],
			wantErr: false,
		},
		{
			name: "err - invalid trytes",
			args: args{
				input:          "",
				isAddress:      true,
				checksumLength: legacy.AddressChecksumTrytesSize,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "err - invalid checksums length",
			args: args{
				input:          "",
				isAddress:      false,
				checksumLength: 2,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checksum.AddChecksum(tt.args.input, tt.args.isAddress, tt.args.checksumLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddChecksum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AddChecksum() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddChecksums(t *testing.T) {
	type args struct {
		inputs         []trinary.Trytes
		isAddress      bool
		checksumLength uint64
	}
	tests := []struct {
		name    string
		args    args
		want    []trinary.Trytes
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				inputs:         addrs,
				isAddress:      true,
				checksumLength: legacy.AddressChecksumTrytesSize,
			},
			want:    addrsWithChecksums,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checksum.AddChecksums(tt.args.inputs, tt.args.isAddress, tt.args.checksumLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddChecksums() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddChecksums() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveChecksum(t *testing.T) {
	type args struct {
		input trinary.Trytes
	}
	tests := []struct {
		name    string
		args    args
		want    trinary.Trytes
		wantErr bool
	}{
		{
			name: "ok - address #1",
			args: args{
				input: addrs[0] + checksums[0],
			},
			want:    addrs[0],
			wantErr: false,
		},
		{
			name: "ok - address #2",
			args: args{
				input: addrs[1] + checksums[1],
			},
			want:    addrs[1],
			wantErr: false,
		},
		{
			name: "ok - address #3",
			args: args{
				input: addrs[2] + checksums[2],
			},
			want:    addrs[2],
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checksum.RemoveChecksum(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveChecksum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RemoveChecksum() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveChecksums(t *testing.T) {
	type args struct {
		inputs []trinary.Trytes
	}
	tests := []struct {
		name    string
		args    args
		want    []trinary.Trytes
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				inputs: addrsWithChecksums,
			},
			want:    addrs,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checksum.RemoveChecksums(tt.args.inputs)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveChecksums() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveChecksums() got = %v, want %v", got, tt.want)
			}
		})
	}
}
