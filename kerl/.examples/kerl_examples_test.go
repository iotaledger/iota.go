package ker_examples_test

// o: SpongeFunction, The SpongeFunction interface using Kerl.
func ExampleNewKerl() {}

// i req: length, The length of the Trits to squeeze out. Must be a multiple of 243.
// o: Trits, The squeezed out Trits.
// o: error, Returned for invalid lengths and internal errors.
func ExampleSqueeze() {}

// i req: length, The length of the trits to squeeze out.
// o: Trits, The Trits representation of the hash.
func ExampleMustSqueeze() {}

// i req: length, The length of the trits to squeeze out.
// o: Trytes, The Trytes representation of the hash.
// o: error, Returned for invalid lengths.
func ExampleSqueezeTrytes() {}

// i req: length, The length of the trits to squeeze out.
// o: Trytes, The Trytes representation of the hash.
func ExampleMustSqueezeTrytes() {}

// i req: in, The Trits slice to absorb. Must be a multiple of 243 in length.
// o: error, Returned for invalid slice lengths and internal errors.
func ExampleAbsorb() {}

// i req: inn, The Trytes to absorb.
// o: error, Returned for internal errors.
func ExampleAbsorbTrytes() {}

// i req: inn, The Trytes to absorb.
func ExampleMustAbsorbTrytes() {}

// i req: trits, The Trits to convert to []byte.
// o: []byte, The converted bytes.
// o: error, Returned for invalid lengths and internal errors.
func ExampleKerlTritsToBytes() {}

// i req: b, The bytes to convert.
// o: Trits, The converted Trits.
// o: error, Returned for invalid lengths and internal errors.
func ExampleKerlBytesToTrits() {}
