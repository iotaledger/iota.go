//go:build ignore

package gen

//go:generate go run ./cmd slot_identifier.tmpl ../block_id.gen.go BlockID b "ids"
//go:generate go run ./cmd slot_identifier.tmpl ../commitment_id.gen.go CommitmentID c ""
//go:generate go run ./cmd slot_identifier.tmpl ../transaction_id.gen.go TransactionID t "ids"
