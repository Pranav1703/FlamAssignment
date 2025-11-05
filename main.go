package main

import (
	"log"
	"os"
	"path/filepath"
	"queueCtl/cmd"
	"queueCtl/internal/config"
	"queueCtl/internal/storage"
)

func main() {
	cfg, err := config.LoadConfig()	
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	dbPath := filepath.Join(cfg.DataDir, "queue.db")

	store,err := storage.NewStore(dbPath)
	if err!=nil{
		log.Fatal("Failed to initialize storage:", err)
	}

	cmd.Execute(store,cfg)
}