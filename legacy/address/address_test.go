package address_test

import (
	"os"
	"testing"

	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/address"
	"github.com/iotaledger/iota.go/legacy/trinary"
	"github.com/stretchr/testify/assert"
)

const seed = "ZLNM9UHJWKTTDEZOTH9CXDEIFUJQCIACDPJIXPOWBDW9LTBHC9AQRIXTIHYLIIURLZCXNSTGNIVC9ISVB"

var addresses = []trinary.Trytes{
	"CLAAFXEY9AHHCSZCXNKDRZEJHIAFVKYORWNOZAGFPAZYNTSLCXUAG9WBSXBRXYEDPVPLXYVDCBCEKRUBD",
	"CDWOADSZWJMDCLYKEDMPIBTYIFAUUAGM9ZQYDKARBUKFXW9LDRQLNG9MI9DGXSOSPDDFFWWJCB9PTGXPW",
	"VVFGHNRFUQEQILXZYUIHWQFUVEEBQCXCUUENADOKRLTVGULYBNMITSYHVRWMAPKPERRLLTC9ELIWSMMMD",
	"IEKMSNDJVIQMLEFUMVQUUFOMI9RWVMJUXKYABPVOYWVMOOVVABYJUKMHZSXDSACYNTEKBCXRJGRJCKXRY",
	"QELVIIRYZZFJSRKMJSDAEOQJRSAWCGMZOGMTBDNJPIOXQTUGMVPCYLWHGHREKDRRVABPULZI9BOWZQPF9",
}

var addressesWithChecksum []trinary.Trytes

var addrWithChecksum = "9OUUHSXJCDAMBWCRUJNYDP9UONYTZPSWYAEPUWUDNTBHCINBJY9QBLERA9OKCBJSUUIADQSIVVFNKTPRYKOAXNZYRX"
var checksums = []trinary.Trytes{"OHHJEQVCY", "QUGONROEZ", "DEJHGSUFY", "GEBW9UBHZ", "L9ACZQYCA"}

func TestMain(m *testing.M) {
	addressesWithChecksum = make([]trinary.Trytes, len(addresses))
	for i := range addresses {
		addressesWithChecksum[i] = addresses[i] + checksums[i]
	}
	os.Exit(m.Run())
}

func TestChecksum(t *testing.T) {

	for i := 0; i < len(addresses); i++ {
		check, err := address.Checksum(addresses[i])
		assert.NoError(t, err)
		assert.Equal(t, checksums[i], check)
	}

	_, err := address.Checksum("BALALAIKA")
	assert.Error(t, err)
}

func TestValidAddress(t *testing.T) {
	assert.NoError(t, address.ValidAddress(addresses[0]))
	assert.NoError(t, address.ValidAddress(addrWithChecksum))
	assert.Error(t, address.ValidAddress("BLABLA"))
}

func TestValidChecksum(t *testing.T) {
	assert.NoError(t, address.ValidChecksum(addresses[0], checksums[0]))
	assert.Error(t, address.ValidChecksum("BLABLA", "DDFDF"))
}

func TestGenerateAddress(t *testing.T) {

	for i := 0; i < len(addresses); i++ {
		addr, err := address.GenerateAddress(seed, uint64(i), legacy.SecurityLevelMedium)
		assert.NoError(t, err)
		assert.Equal(t, addresses[i], addr)
	}

	for i := 0; i < len(addresses); i++ {
		addr, err := address.GenerateAddress(seed, uint64(i), legacy.SecurityLevelMedium, true)
		assert.NoError(t, err)
		assert.Equal(t, addresses[i]+checksums[i], addr)
	}
}

func TestGenerateAddresses(t *testing.T) {
	addrs, err := address.GenerateAddresses(seed, 0, 5, legacy.SecurityLevelMedium, false)
	assert.NoError(t, err)
	assert.Equal(t, addresses, addrs)

	addrs, err = address.GenerateAddresses(seed, 0, 5, legacy.SecurityLevelMedium, true)
	assert.NoError(t, err)
	assert.Equal(t, addressesWithChecksum, addrs)
}
