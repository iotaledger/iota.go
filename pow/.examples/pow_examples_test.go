package pow_examples_test

// i req: trunkTx, The trunk transaction hash.
// i req: branchTx, the branch transaction hash.
// i req: trytes, The transactions Trytes slice.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: pow, The Proof-of-Work implementation function.
// o: []Trytes, The transaction Trytes with computed nonce fields, ready for broadcasting.
// o: error, Returned for invalid transaction Trytes and internal errors.
func ExampleDoPoW() {}

// i req: name, The name of the Proof-of-Work implementation to get.
// o: ProofOfWorkFunc, The Proof-of-Work implementation.
// o: error, Returned if the Proof-of-Work implementation is not known.
func ExampleGetProofOfWorkImpl() {}

// o: []string, The names of all available Proof-of-Work implementations.
func ExampleGetProofOfWorkImplementations() {}

// o: string, The name of the Proof-of-Work function.
// o: ProofOfWorkFunc, The actual Proof-of-Work implementation.
func ExampleGetFastestProofOfWorkImpl() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleGoProofOfwork() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleSyncGoProofOfwork() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleAVXProofOfwork() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleSyncAVXProofOfwork() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleCProofOfwork() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleSyncCProofOfwork() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleC128ProofOfwork() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleSyncC128ProofOfwork() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleCARM64ProofOfwork() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleSyncCARM64ProofOfwork() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleSSEProofOfwork() {}

// i req: trytes, The Trytes of the transaction.
// i req: mwm, The minimum weight magnitude to fulfill.
// i req: ...int, The amount of goroutines to spawn for executing the Proof-of-Work.
// o: Trytes, The computed nonce.
// o: error, Returned for internal errors.
func ExampleSyncSSEProofOfwork() {}
