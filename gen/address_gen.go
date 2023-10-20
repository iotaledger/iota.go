//go:build ignore

package gen

//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 address.tmpl ../address_ed25519.gen.go Ed25519Address addr "ed25519unlock,frompubkey" "AddressType=AddressEd25519,Description=is the Blake2b-256 hash of an Ed25519 public key"
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 address.tmpl ../address_account.gen.go AccountAddress addr "chainid" "IdentifierType=AccountID,AddressType=AddressAccount,Description=is the Blake2b-256 hash of the OutputID which created it"
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 address.tmpl ../address_nft.gen.go NFTAddress addr "chainid" "IdentifierType=NFTID,AddressType=AddressNFT,Description=is the Blake2b-256 hash of the OutputID which created it"
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 address.tmpl ../address_implicit_account_creation.gen.go ImplicitAccountCreationAddress addr "storagescore,ed25519unlock,frompubkey" "AddressType=AddressImplicitAccountCreation,Description=is an address that is used to create implicit accounts by sending basic outputs to it"
