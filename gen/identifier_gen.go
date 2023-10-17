//go:build ignore

package gen

//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@86973b2edb3b identifier.tmpl ../identifier.gen.go Identifier i ""
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@86973b2edb3b identifier.tmpl ../identifier_account.gen.go AccountID a ""
