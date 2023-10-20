//go:build ignore

package gen

//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 index.tmpl ../slot_index.gen.go Slot i "" "Description=is the index of a"
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 index.tmpl ../epoch_index.gen.go Epoch i "" "Description=is the index of an"
