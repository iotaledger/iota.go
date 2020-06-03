package merkle

import (
	"bufio"
	"encoding/binary"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
)

// WriteTo writes the binary representation of the Merkle tree to a buffer.
func (mt *MerkleTree) WriteTo(buf *bufio.Writer) (n int64, err error) {

	/*
	 4 bytes uint32 depth
	 4 bytes uint32 lengthLayers
	*/

	var bytesWritten int64 = 0

	if err := binary.Write(buf, binary.LittleEndian, uint32(mt.Depth)); err != nil {
		return bytesWritten, err
	}
	bytesWritten += int64(binary.Size(uint32(mt.Depth)))

	if err := binary.Write(buf, binary.LittleEndian, uint32(len(mt.Layers))); err != nil {
		return bytesWritten, err
	}
	bytesWritten += int64(binary.Size(uint32(len(mt.Layers))))

	for layer := 0; layer < len(mt.Layers); layer++ {
		merkleTreeLayer := mt.Layers[layer]

		bytesWrittenLayer, err := merkleTreeLayer.WriteTo(buf)
		if err != nil {
			return bytesWritten, err
		}
		bytesWritten += bytesWrittenLayer
	}

	return bytesWritten, err
}

// ReadFrom parses the binary encoded representation of the Merkle tree from a buffer.
func (mt *MerkleTree) ReadFrom(buf *bufio.Reader) (n int64, err error) {

	/*
	 4 bytes uint32 depth
	 4 bytes uint32 lengthLayers
	*/

	var bytesRead int64 = 0

	var depth uint32
	if err := binary.Read(buf, binary.LittleEndian, &depth); err != nil {
		return bytesRead, err
	}
	mt.Depth = int(depth)
	bytesRead += int64(binary.Size(&depth))

	var lengthLayers uint32
	if err := binary.Read(buf, binary.LittleEndian, &lengthLayers); err != nil {
		return bytesRead, err
	}
	bytesRead += int64(binary.Size(&lengthLayers))

	mt.Layers = make(map[int]*MerkleTreeLayer)

	for i := 0; i < int(lengthLayers); i++ {
		mtl := &MerkleTreeLayer{}

		bytesReadLayer, err := mtl.ReadFrom(buf)
		if err != nil {
			return bytesRead, err
		}
		bytesRead += bytesReadLayer

		mt.Layers[mtl.Level] = mtl
	}

	mt.Root = mt.Layers[0].Hashes[0]
	return bytesRead, nil
}

// WriteTo writes the binary representation of the Merkle tree layer to a buffer.
func (mtl *MerkleTreeLayer) WriteTo(buf *bufio.Writer) (int64, error) {

	var bytesWritten int64 = 0

	if err := binary.Write(buf, binary.LittleEndian, uint32(mtl.Level)); err != nil {
		return bytesWritten, err
	}
	bytesWritten += int64(binary.Size(uint32(mtl.Level)))

	if err := binary.Write(buf, binary.LittleEndian, uint32(len(mtl.Hashes))); err != nil {
		return bytesWritten, err
	}
	bytesWritten += int64(binary.Size(uint32(len(mtl.Hashes))))

	for _, hash := range mtl.Hashes {
		nodesData := trinary.MustTrytesToBytes(hash)[:49]
		if err := binary.Write(buf, binary.LittleEndian, nodesData); err != nil {
			return bytesWritten, err
		}
		bytesWritten += int64(binary.Size(nodesData))
	}

	return bytesWritten, nil
}

// ReadFrom parses the binary encoded representation of the merkle tree layer from a buffer.
func (mtl *MerkleTreeLayer) ReadFrom(buf *bufio.Reader) (n int64, err error) {

	var read int64 = 0
	var level uint32

	if err := binary.Read(buf, binary.LittleEndian, &level); err != nil {
		return read, err
	}
	read += int64(binary.Size(&level))
	mtl.Level = int(level)

	var hashesCount uint32
	if err := binary.Read(buf, binary.LittleEndian, &hashesCount); err != nil {
		return read, err
	}
	read += int64(binary.Size(&hashesCount))

	hashBuf := make([]byte, 49)
	for i := 0; i < int(hashesCount); i++ {
		if err := binary.Read(buf, binary.LittleEndian, hashBuf); err != nil {
			return read, err
		}
		read += int64(binary.Size(hashBuf))
		mtl.Hashes = append(mtl.Hashes, trinary.MustBytesToTrytes(hashBuf, consts.HashTrytesSize))
	}

	return read, nil
}
