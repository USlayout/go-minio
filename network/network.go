package network

package network

import (
    "fmt"
    "net/http"
    "github.com/USlayout/go-minio/storage"
)

func StartServer(addr string) error {
    http.HandleFunc("/upload", handleUpload)
    http.HandleFunc("/download", handleDownload)
    http.HandleFunc("/list", handleList)      // 新規追加
    http.HandleFunc("/delete", handleDelete)  // 新規追加
    
    // CORS対応
    http.HandleFunc("/", corsMiddleware)

    fmt.Println("Server running on", addr)
    return http.ListenAndServe(addr, nil)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    err = storage.SaveFile(header.Filename, file, header.Size)
    if err != nil {
        http.Error(w, "Upload failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Uploaded: %s\n", header.Filename)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    if name == "" {
        http.Error(w, "Missing name param", http.StatusBadRequest)
        return
    }

    reader, err := storage.GetFile(name)
    if err != nil {
        http.Error(w, "Download failed: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer reader.Close() // リソースリークを防ぐ

    w.Header().Set("Content-Disposition", "attachment; filename="+name)
    _, err = io.Copy(w, reader) // ServeContentの代わりにio.Copyを使用
    if err != nil {
        log.Printf("Error copying file: %v", err)
    }
}

func corsMiddleware(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }
    
    http.Error(w, "Not Found", http.StatusNotFound)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
    // CORS設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    err = storage.SaveFile(header.Filename, file, header.Size)
    if err != nil {
        http.Error(w, "Upload failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Uploaded: %s\n", header.Filename)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
    // CORS設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    name := r.URL.Query().Get("name")
    if name == "" {
        http.Error(w, "Missing name param", http.StatusBadRequest)
        return
    }

    reader, err := storage.GetFile(name)
    if err != nil {
        http.Error(w, "Download failed: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer reader.Close()

    w.Header().Set("Content-Disposition", "attachment; filename="+name)
    _, err = io.Copy(w, reader)
    if err != nil {
        log.Printf("Error copying file: %v", err)
    }
}

// ファイル一覧を返すハンドラー
func handleList(w http.ResponseWriter, r *http.Request) {
    // CORS設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/json")
    
    if r.Method != http.MethodGet {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    files, err := storage.ListFiles()
    if err != nil {
        http.Error(w, "Failed to list files: "+err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(files)
}

// ファイル削除ハンドラー
func handleDelete(w http.ResponseWriter, r *http.Request) {
    // CORS設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    if r.Method != http.MethodDelete {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    name := r.URL.Query().Get("name")
    if name == "" {
        http.Error(w, "Missing name param", http.StatusBadRequest)
        return
    }

    err := storage.DeleteFile(name)
    if err != nil {
        http.Error(w, "Delete failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Deleted: %s\n", name)
}