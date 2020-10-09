package bech32

var gen = []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

// For more details on the checksum calculation, please refer to BIP 173.
func bech32CreateChecksum(hrp string, blocks []byte) []byte {
	values := append(bech32HrpExpand(hrp), blocks...)
	polymod := bech32Polymod(append(values, []byte{0, 0, 0, 0, 0, 0}...)) ^ 1
	res := make([]byte, 6)
	for i := range res {
		res[i] = byte((polymod >> (5 * (5 - i))) & 31)
	}
	return res
}

// For more details on the polymod calculation, please refer to BIP 173.
func bech32Polymod(values []byte) int {
	chk := 1
	for _, v := range values {
		b := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ int(v)
		for i := range gen {
			if (b>>i)&1 != 0 {
				chk ^= gen[i]
			}
		}
	}
	return chk
}

// For more details on String expansion, please refer to BIP 173.
func bech32HrpExpand(s string) []byte {
	res := make([]byte, 0, 2*len(s)+1)
	for _, x := range []byte(s) {
		res = append(res, x>>5)
	}
	res = append(res, 0)
	for _, x := range []byte(s) {
		res = append(res, x&31)
	}
	return res
}

// For more details on the checksum verification, please refer to BIP 173.
func bech32VerifyChecksum(hrp string, data []byte) bool {
	return bech32Polymod(append(bech32HrpExpand(hrp), data...)) == 1
}
