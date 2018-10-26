package samples

import (
	. "github.com/iotaledger/iota.go/trinary"
	"strings"
)

var Seed = "HZVEINVKVIKGFRAWRTRXWD9JLIYLCQNCXZRBLDETPIQGKZJRYKZXLTV9JNUVBIAHAGUZVIQWIAWDZ9ACW"

func DefaultHashes() Hashes {
	return Hashes{strings.Repeat("A", 81), strings.Repeat("B", 81)}
}
