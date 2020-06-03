/*
Ported from https://github.com/gohornet/hornet codebase.
Original authors: muXxer <mux3r@web.de>
                  Alexander Sporn <github@alexsporn.de>
                  Thoralf-M <46689931+Thoralf-M@users.noreply.github.com>


Package tool provides miscellneous functionality to deal with IOTA types.
*/
package tool

import (
	"bufio"
	"os"

	"github.com/iotaledger/iota.go/merkle"
)

// StoreMerkleTreeFile stores the MerkleTree structure in a binary output file.
func StoreMerkleTreeFile(filePath string, merkleTree *merkle.MerkleTree) error {

	outputFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// write into the file with an 8kB buffer.
	fileBufWriter := bufio.NewWriterSize(outputFile, 8192)

	if _, err := merkleTree.WriteTo(fileBufWriter); err != nil {
		return err
	}

	if err := fileBufWriter.Flush(); err != nil {
		return err
	}

	return nil
}

// LoadMerkleTreeFile loads a binary file persisted with StoreMerkleTreeFile
// into a MerkleTree structure.
func LoadMerkleTreeFile(filePath string) (*merkle.MerkleTree, error) {

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := bufio.NewReader(file)

	merkleTree := &merkle.MerkleTree{}
	if _, err := merkleTree.ReadFrom(buf); err != nil {
		return nil, err
	}

	return merkleTree, nil
}
