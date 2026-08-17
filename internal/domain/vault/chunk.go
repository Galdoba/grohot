package vault

import (
	"fmt"

	"github.com/Galdoba/grohot/internal/domain/note"
)

// ChunkAll applies the configured chunker to all segments of all notes
// and returns the combined list of chunks.
func (v *Vault) ChunkAll() ([]note.Chunk, error) {
	if v.chunker == nil {
		return nil, fmt.Errorf("chunker is not set")
	}
	var all []note.Chunk
	for _, n := range v.notes {
		chunks, err := v.chunkNote(n)
		if err != nil {
			return nil, err
		}
		all = append(all, chunks...)
	}
	return all, nil
}

// chunkNote chunks all segments of a single note.
func (v *Vault) chunkNote(n *note.Note) ([]note.Chunk, error) {
	var chunks []note.Chunk
	for _, seg := range n.Segments() {
		segChunks, err := v.chunker.Chunk(seg)
		if err != nil {
			return nil, fmt.Errorf("chunk segment %s: %w", seg.ID, err)
		}
		chunks = append(chunks, segChunks...)
	}
	return chunks, nil
}