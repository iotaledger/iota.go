//go:build ignore

package gen

//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 identifier.tmpl ../identifier.gen.go Identifier i "" ""
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 identifier.tmpl ../identifier_account.gen.go AccountID a "chainid" "AddressType=AccountAddress"
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 identifier.tmpl ../identifier_anchor.gen.go AnchorID a "chainid" "AddressType=AnchorAddress"
