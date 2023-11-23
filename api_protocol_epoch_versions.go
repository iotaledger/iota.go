package iotago

import (
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/stringify"
)

type ProtocolEpochVersions struct {
	versionsPerEpoch []ProtocolEpochVersion
	knownVersions    map[Version]EpochIndex
}

func NewProtocolEpochVersions() *ProtocolEpochVersions {
	return &ProtocolEpochVersions{
		versionsPerEpoch: make([]ProtocolEpochVersion, 0),
		knownVersions:    make(map[Version]EpochIndex),
	}
}

func (p *ProtocolEpochVersions) VersionForEpoch(epoch EpochIndex) Version {
	for i := len(p.versionsPerEpoch) - 1; i >= 0; i-- {
		if p.versionsPerEpoch[i].StartEpoch <= epoch {
			return p.versionsPerEpoch[i].Version
		}
	}

	// This means that the protocol versions are not properly configured.
	panic(ierrors.Errorf("could not find a protocol version for epoch %d", epoch))
}

func (p *ProtocolEpochVersions) Add(version Version, epoch EpochIndex) {
	if _, exists := p.knownVersions[version]; exists {
		return
	}

	p.versionsPerEpoch = append(p.versionsPerEpoch, ProtocolEpochVersion{
		Version:    version,
		StartEpoch: epoch,
	})

	p.knownVersions[version] = epoch

	sort.Slice(p.versionsPerEpoch, func(i, j int) bool {
		return p.versionsPerEpoch[i].Version < p.versionsPerEpoch[j].Version
	})
}

func (p *ProtocolEpochVersions) Slice() []ProtocolEpochVersion {
	return lo.CopySlice(p.versionsPerEpoch)
}

func (p *ProtocolEpochVersions) EpochForVersion(version Version) (EpochIndex, bool) {
	startEpoch, exists := p.knownVersions[version]
	return startEpoch, exists
}

func (p *ProtocolEpochVersions) Bytes() []byte {
	versionsBytes := make([]byte, 0)
	for _, protocolEpochVersion := range p.versionsPerEpoch {
		versionsBytes = append(versionsBytes, protocolEpochVersion.Bytes()...)
	}

	return versionsBytes
}

func (p *ProtocolEpochVersions) String() string {
	builder := stringify.NewStructBuilder("ProtocolEpochVersions")

	for i, protocolEpochVersion := range p.versionsPerEpoch {
		builder.AddField(stringify.NewStructField(fmt.Sprintf("entry%d", i), protocolEpochVersion.String()))
	}

	return builder.String()
}

type ProtocolEpochVersion struct {
	Version    Version
	StartEpoch EpochIndex
}

func (p *ProtocolEpochVersion) Bytes() []byte {
	return append(lo.PanicOnErr(p.Version.Bytes()), lo.PanicOnErr(p.StartEpoch.Bytes())...)
}

func (p *ProtocolEpochVersion) String() string {
	return fmt.Sprintf("Version %d: Epoch %d", p.Version, p.StartEpoch)
}
