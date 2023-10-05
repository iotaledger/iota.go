//go:build ignore

package gen

//go:generate go run ./cmd slot_identifier.tmpl ../block_id.gen.go BlockID b "ids"
