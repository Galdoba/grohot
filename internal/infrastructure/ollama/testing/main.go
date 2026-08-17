package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Galdoba/grohot/internal/infrastructure/ollama"
)

func main() {
	client := ollama.NewClient(
		ollama.WithDebug(),
		ollama.WithOutput(os.Stdout),
		ollama.WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	// Проверка наличия бинаря
	if err := client.EnsureInstalled(); err != nil {
		fmt.Println("Ollama не установлен:", err)
		return
	}

	// Список моделей
	out, err := client.List(ctx)
	if err != nil {
		fmt.Println("list error:", err)
		return
	}
	fmt.Println("Models:\n", string(out))

	// Pull модели (будет виден прогресс в stdout)
	_, err = client.Pull(ctx, "bge-m3")
	if err != nil {
		fmt.Println("pull error:", err)
		return
	}

	// Show
	info, err := client.Show(ctx, "bge-m3")
	if err != nil {
		fmt.Println("show error:", err)
		return
	}
	fmt.Println("Model info:\n", string(info))

	// Stop all running models
	_, err = client.Stop(ctx, "")
	if err != nil {
		fmt.Println("stop error:", err)
		return
	}
}
