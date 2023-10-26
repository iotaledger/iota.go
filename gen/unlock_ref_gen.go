//go:build ignore

package gen

//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 unlock_ref.tmpl ../unlock_reference.gen.go ReferenceUnlock r "" "SourceAddressType=ChainAddress,UnlockType=UnlockReference"
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 unlock_ref.tmpl ../unlock_account.gen.go AccountUnlock r "chainable" "SourceAddressType=*AccountAddress,UnlockType=UnlockAccount"
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 unlock_ref.tmpl ../unlock_anchor.gen.go AnchorUnlock r "chainable" "SourceAddressType=*AnchorAddress,UnlockType=UnlockAnchor"
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 unlock_ref.tmpl ../unlock_nft.gen.go NFTUnlock r "chainable" "SourceAddressType=*NFTAddress,UnlockType=UnlockNFT"
