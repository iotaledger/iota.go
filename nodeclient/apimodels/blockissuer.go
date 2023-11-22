package apimodels

// BlockIssuerInfo is the response to the BlockIssuerRouteInfo endpoint.
type BlockIssuerInfo struct {
	// The account address of the block issuer.
	BlockIssuerAddress string `serix:""`
	// The number of trailing zeroes required for the proof of work to be valid.
	PowTargetTrailingZeros uint8 `serix:""`
}
