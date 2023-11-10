//go:build ignore

package gen

//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 feat_metadata.tmpl ../feat_metadata.gen.go MetadataFeature m "" "FeatureType=FeatureMetadata,"
//go:generate go run github.com/iotaledger/hive.go/codegen/features/cmd@13da292 feat_metadata.tmpl ../feat_metadata_state.gen.go StateMetadataFeature m "" "FeatureType=FeatureStateMetadata,"
