package curl

// uint256 is a simple 256-bit uint modelled as an array of uint64.
type uint256 [4]uint64

// bit returns the value of the i-th bit of z.
// If i ≥ 256 the bit a position i % 256 is considered.
func (z *uint256) bit(i uint) uint {
	return uint((z[(i/64)%4] >> (i % 64)) & 1)
}

// setBit sets the i-th bit in z to 1 and returns z.
// If i ≥ 256 the bit a position i % 256 is considered.
func (z *uint256) setBit(i uint) *uint256 {
	z[(i/64)%4] |= uint64(1) << (i % 64)
	return z
}

// shrInto sets z = z | x >> s and returns z.
func (z *uint256) shrInto(x *uint256, s uint) *uint256 {
	offset, r := s/64, s%64
	if r == 0 { // no shifting is needed
		for i := offset; i < 4; i++ {
			z[(i-offset)%4] |= x[i] // the modulus is a hint to the compiler that no bound checks are needed
		}
		return z
	}
	l := 64 - r
	l &= 63 // hint to the compiler that shifts by l don't need guard code

	switch offset {
	case 0:
		z[0] |= x[0]>>r | x[1]<<l
		z[1] |= x[1]>>r | x[2]<<l
		z[2] |= x[2]>>r | x[3]<<l
		z[3] |= x[3] >> r
	case 1:
		z[0] |= x[1]>>r | x[2]<<l
		z[1] |= x[2]>>r | x[3]<<l
		z[2] |= x[3] >> r
	case 2:
		z[0] |= x[2]>>r | x[3]<<l
		z[1] |= x[3] >> r
	case 3:
		z[0] |= x[3] >> r
	}
	return z
}

// shlInto sets z = z | x << s and returns z.
func (z *uint256) shlInto(x *uint256, s uint) *uint256 {
	offset, l := s/64, s%64
	if l == 0 { // no shifting is needed
		for i := offset; i < 4; i++ {
			z[i] |= x[(i-offset)%4] // the modulus is a hint to the compiler that no bound checks are needed
		}
		return z
	}
	r := 64 - l
	r &= 63 // hint to the compiler that shifts by r don't need guard code

	switch offset {
	case 0:
		z[3] |= x[3]<<l | x[2]>>r
		z[2] |= x[2]<<l | x[1]>>r
		z[1] |= x[1]<<l | x[0]>>r
		z[0] |= x[0] << l
	case 1:
		z[3] |= x[2]<<l | x[1]>>r
		z[2] |= x[1]<<l | x[0]>>r
		z[1] |= x[0] << l
	case 2:
		z[3] |= x[1]<<l | x[0]>>r
		z[2] |= x[0] << l
	case 3:
		z[3] |= x[0] << l
	}
	return z
}

// norm243 clears the bits higher than 243 in z and returns z.
func (z *uint256) norm243() *uint256 {
	z[3] &= 1<<(64-(256-243)) - 1
	return z
}
