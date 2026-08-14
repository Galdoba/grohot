package chunker

import (
    "github.com/Galdoba/grohot/internal/domain/link"
    "github.com/Galdoba/grohot/internal/domain/note"
)

type SegmentChunker struct{}

func NewSegmentChunker() *SegmentChunker {
    return &SegmentChunker{}
}

func (c *SegmentChunker) Chunk(seg *note.Segment) ([]note.Chunk, error) {
    text := seg.Text()          // метод нужно добавить (см. ниже)
    if text == "" {
        return nil, nil
    }
    input := note.ChunkInput{
        Text:        text,
        Index:       0,
        TotalChunks: 1,
    }
    chunk, err := note.NewChunk(seg, input)
    if err != nil {
        return nil, err
    }
    // Копируем связи сегмента в метаданные чанка
    links := make([]link.Link, 0, len(seg.OutgoingLinks)+len(seg.IncomingLinks))
    links = append(links, seg.OutgoingLinks...)
    links = append(links, seg.IncomingLinks...)
    chunk.Metadata.Links = links
    return []note.Chunk{*chunk}, nil
}