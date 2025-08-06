package main

import (
    "log"
    "github.com/USlayout/go-minio/network"
    "github.com/USlayout/go-minio/storage"
)


func main() {
    // MinIOクライアント初期化
    err := storage.InitMinIO()
    if err != nil {
        log.Fatalf("MinIO init error: %v", err)
    }

    // HTTPサーバ起動
    err = network.StartServer(":8080")
    if err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
