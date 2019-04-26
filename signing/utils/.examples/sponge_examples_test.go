package sponge_examples_test

// o: SpongeFunction, The SpongeFunction interface using CurlP27.
func ExampleNewCurlP27() {}

// o: SpongeFunction, The SpongeFunction interface using CurlP81.
func ExampleNewCurlP81() {}

// o: SpongeFunction, The SpongeFunction interface using Kerl.
func ExampleNewKerl() {}

// i req: spongeFunc, The sponge function to use. If none is given, the defaultSpongeFuncCreator will be used.
// i: defaultSpongeFuncCreator, The optional sponge function creator to use. If no spongeFunc and no defaultSpongeFuncCreator is given, Kerl will be used.
// o: SpongeFunction, The SpongeFunction interface.
func ExampleGetSpongeFunc() {}
