package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Galdoba/grohot/internal/domain/vault"
	"github.com/Galdoba/grohot/internal/infrastructure/chunker"
	"github.com/Galdoba/grohot/internal/infrastructure/embedder"
	"github.com/Galdoba/grohot/internal/infrastructure/obsidian/hierarchy"
	"github.com/Galdoba/grohot/internal/infrastructure/obsidian/parser"
	"github.com/Galdoba/grohot/internal/infrastructure/storage"
)

func main() {
	home, _ := os.UserHomeDir()
	start := time.Now()

	// 1. Определяем корень хранилища
	path := filepath.Join(home, "Documents", "Obsidian", "nounamental_rhizome")
	root, err := vault.FindVaultRoot(path)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Создаём компоненты
	blockParser := parser.NewParser()
	blockBuilder := hierarchy.NewBuilder()
	chunker := chunker.NewSegmentChunker()
	cfg := vault.Dependencies{
		Parser:  blockParser,
		Builder: blockBuilder,
		Chunker: chunker,
	}
	v, err := vault.NewVault(root, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Сканируем и загружаем все заметки
	if err := v.ScanAndLoad(); err != nil {
		log.Fatal(err)
	}

	// 4. Строим горизонтальные связи
	if err := v.BuildLinks(); err != nil {
		log.Fatal(err)
	}

	// 5. Генерируем чанки
	chunks, err := v.ChunkAll()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("generated chunks: %d\n", len(chunks))

	// 6. Эмбеддинг
	emb := embedder.NewHashEmbedder(256)
	store := storage.NewJSONStore(filepath.Join(home, "chunks.json"))
	for i := range chunks {
		chunk := &chunks[i]
		embedding, err := emb.Embed(chunk.TextToEmbed())
		if err != nil {
			log.Printf("failed to embed chunk %s: %v", chunk.ID, err)
			continue
		}
		chunk.Embedding = embedding
		store.Insert(*chunk)
	}

	// 7. Сохраняем в файл
	if err := store.Save(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("elapsed: %v\n", time.Since(start))

	time.Sleep(time.Second * 3)

	// Загружаем хранилище
	store = storage.NewJSONStore("chunks.json")
	if err := store.Load(); err != nil {
		log.Fatal(err)
	}

	// Получаем вектор запроса (используем тот же эмбеддер)
	emb = embedder.NewHashEmbedder(256)
	queryVec, _ := emb.Embed("What weapons should I use in boarding action during space combat?")

	// Ищем топ-5
	results, err := store.Search(queryVec, 10)
	if err != nil {
		log.Fatal(err)
	}
	for i, res := range results {
		fmt.Printf("%d. score=%.4f  %s\n", i, res.Score, res.Chunk.Metadata.FilePath)
	}
}