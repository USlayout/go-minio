package network

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "github.com/USlayout/go-minio/storage"
)

func StartServer(addr string) error {
    http.HandleFunc("/upload", handleUpload)
    http.HandleFunc("/upload-multiple", handleMultipleUpload)  // 複数ファイルアップロード
    http.HandleFunc("/upload-folder", handleFolderUpload)     // フォルダアップロード
    http.HandleFunc("/download", handleDownload)
    http.HandleFunc("/list", handleList)      // 新規追加
    http.HandleFunc("/list-folders", handleListFolders)  // フォルダ構造付き一覧
    http.HandleFunc("/delete", handleDelete)  // 新規追加
    
    // CORS対応
    http.HandleFunc("/", corsMiddleware)

    fmt.Println("Server running on", addr)
    return http.ListenAndServe(addr, nil)
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

// 複数ファイルアップロードハンドラー
func handleMultipleUpload(w http.ResponseWriter, r *http.Request) {
    // CORS設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    // フォーム解析（最大32MBまで）
    err := r.ParseMultipartForm(32 << 20)
    if err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }

    files := r.MultipartForm.File["files"]
    if len(files) == 0 {
        http.Error(w, "No files provided", http.StatusBadRequest)
        return
    }

    var uploadedFiles []string
    var errors []string

    for _, fileHeader := range files {
        file, err := fileHeader.Open()
        if err != nil {
            errors = append(errors, fmt.Sprintf("Failed to open %s: %v", fileHeader.Filename, err))
            continue
        }

        err = storage.SaveFile(fileHeader.Filename, file, fileHeader.Size)
        file.Close()

        if err != nil {
            errors = append(errors, fmt.Sprintf("Failed to save %s: %v", fileHeader.Filename, err))
        } else {
            uploadedFiles = append(uploadedFiles, fileHeader.Filename)
        }
    }

    // 結果を返す
    result := map[string]interface{}{
        "uploaded": uploadedFiles,
        "errors":   errors,
        "total":    len(files),
        "success":  len(uploadedFiles),
        "failed":   len(errors),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

// フォルダアップロードハンドラー（パス情報付き）
func handleFolderUpload(w http.ResponseWriter, r *http.Request) {
    // CORS設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    // フォーム解析（最大100MBまで）
    err := r.ParseMultipartForm(100 << 20)
    if err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }

    files := r.MultipartForm.File["files"]
    if len(files) == 0 {
        http.Error(w, "No files provided", http.StatusBadRequest)
        return
    }

    var uploadedFiles []string
    var errors []string

    for _, fileHeader := range files {
        file, err := fileHeader.Open()
        if err != nil {
            errors = append(errors, fmt.Sprintf("Failed to open %s: %v", fileHeader.Filename, err))
            continue
        }

        // ファイルパスをそのまま使用（フォルダ構造を維持）
        err = storage.SaveFile(fileHeader.Filename, file, fileHeader.Size)
        file.Close()

        if err != nil {
            errors = append(errors, fmt.Sprintf("Failed to save %s: %v", fileHeader.Filename, err))
        } else {
            uploadedFiles = append(uploadedFiles, fileHeader.Filename)
        }
    }

    // 結果を返す
    result := map[string]interface{}{
        "uploaded": uploadedFiles,
        "errors":   errors,
        "total":    len(files),
        "success":  len(uploadedFiles),
        "failed":   len(errors),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

// フォルダ構造付きファイル一覧ハンドラー
func handleListFolders(w http.ResponseWriter, r *http.Request) {
    // CORS設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/json")
    
    if r.Method != http.MethodGet {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    folders, err := storage.ListFilesWithFolders()
    if err != nil {
        http.Error(w, "Failed to list folders: "+err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(folders)
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