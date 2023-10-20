//go:build ignore

package gen

//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 address.tmpl ../address_ed25519.gen.go Ed25519Address addr "" "AddressType=AddressEd25519,Description=is the Blake2b-256 hash of an Ed25519 public key"
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 address.tmpl ../address_account.gen.go AccountAddress addr "chainid" "IdentifierType=AccountID,AddressType=AddressAccount,Description=is the Blake2b-256 hash of the OutputID which created it"
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 address.tmpl ../address_nft.gen.go NFTAddress addr "chainid" "IdentifierType=NFTID,AddressType=AddressNFT,Description=is the Blake2b-256 hash of the OutputID which created it"
