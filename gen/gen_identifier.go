//go:build ignore

package gen

//go:generate go run ./cmd identifier.tmpl ../identifier.gen.go Identifier i ""
//go:generate go run ./cmd identifier.tmpl ../identifier_account.gen.go AccountID a "output"
