package network

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"
    "github.com/USlayout/go-minio/storage"
    "github.com/USlayout/go-minio/auth"
)

func StartServer(addr string) error {
    // 認証エンドポイント
    http.HandleFunc("/auth/login", handleLogin)
    http.HandleFunc("/auth/refresh", handleRefresh)
    http.HandleFunc("/auth/me", auth.JWTMiddleware(handleMe))
    
    // 保護されたエンドポイント（JWT認証が必要）
    http.HandleFunc("/upload", auth.JWTMiddleware(handleUpload))
    http.HandleFunc("/download", auth.JWTMiddleware(handleDownload))
    http.HandleFunc("/delete", auth.JWTMiddleware(handleDelete))
    http.HandleFunc("/mkdir", auth.JWTMiddleware(handleMakeDir))
    http.HandleFunc("/list", auth.JWTMiddleware(handleList))
    
    // 管理者専用エンドポイント
    http.HandleFunc("/admin/users", auth.AdminOnlyMiddleware(handleAdminUsers))
    
    // CORS対応
    http.HandleFunc("/", corsMiddleware)

    fmt.Println("MinIO Cloud Storage Server running on", addr)
    fmt.Println("Available endpoints:")
    fmt.Println("  POST /auth/login    - ユーザーログイン")
    fmt.Println("  POST /auth/refresh  - トークンリフレッシュ")
    fmt.Println("  GET  /auth/me       - ユーザー情報取得")
    fmt.Println("  POST /upload        - ファイルアップロード (要認証)")
    fmt.Println("  GET  /download      - ファイルダウンロード (要認証)")
    fmt.Println("  DELETE /delete      - ファイル削除 (要認証)")
    fmt.Println("  POST /mkdir         - フォルダ作成 (要認証)")
    fmt.Println("  GET  /list          - ファイル/フォルダ一覧 (要認証)")
    fmt.Println("  GET  /admin/users   - ユーザー管理 (管理者のみ)")
    
    return http.ListenAndServe(addr, nil)
}

func corsMiddleware(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }
    
    http.Error(w, "Not Found", http.StatusNotFound)
}

// ログインハンドラー
func handleLogin(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/json")
    
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }
    
    var loginReq struct {
        UserID   string `json:"userID"`
        Password string `json:"password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // ユーザー認証
    user, err := auth.AuthenticateUser(loginReq.UserID, loginReq.Password)
    if err != nil {
        http.Error(w, "Authentication failed: "+err.Error(), http.StatusUnauthorized)
        return
    }
    
    // JWTトークン生成
    accessToken, err := auth.GenerateToken(*user)
    if err != nil {
        http.Error(w, "Token generation failed", http.StatusInternalServerError)
        return
    }
    
    // リフレッシュトークン生成
    refreshToken, err := auth.GenerateRefreshToken(user.UserID)
    if err != nil {
        http.Error(w, "Refresh token generation failed", http.StatusInternalServerError)
        return
    }
    
    response := map[string]interface{}{
        "accessToken":  accessToken,
        "refreshToken": refreshToken,
        "user":         user,
        "expiresIn":    86400, // 24時間（秒）
    }
    
    json.NewEncoder(w).Encode(response)
}

// トークンリフレッシュハンドラー
func handleRefresh(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/json")
    
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }
    
    var refreshReq struct {
        RefreshToken string `json:"refreshToken"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&refreshReq); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // 新しいアクセストークンを生成
    newAccessToken, err := auth.RefreshAccessToken(refreshReq.RefreshToken)
    if err != nil {
        http.Error(w, "Token refresh failed: "+err.Error(), http.StatusUnauthorized)
        return
    }
    
    response := map[string]interface{}{
        "accessToken": newAccessToken,
        "expiresIn":   86400, // 24時間（秒）
    }
    
    json.NewEncoder(w).Encode(response)
}

// ユーザー情報取得ハンドラー
func handleMe(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/json")
    
    if r.Method != http.MethodGet {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }
    
    userID := r.Header.Get("X-User-ID")
    role := r.Header.Get("X-User-Role")
    
    response := map[string]interface{}{
        "userID": userID,
        "role":   role,
    }
    
    json.NewEncoder(w).Encode(response)
}

// 管理者専用ユーザー管理ハンドラー
func handleAdminUsers(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/json")
    
    if r.Method != http.MethodGet {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }
    
    // 簡単なユーザー一覧を返す（実際のアプリケーションではデータベースから取得）
    users := []map[string]interface{}{
        {"userID": "user123", "username": "testuser", "role": "user"},
        {"userID": "admin", "username": "admin", "role": "admin"},
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "users": users,
    })
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
    // CORS設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    // 認証済みユーザーIDを取得
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        http.Error(w, "User ID not found in token", http.StatusBadRequest)
        return
    }
    
    virtualPath := r.FormValue("path")
    if virtualPath == "" {
        virtualPath = "" // ルートディレクトリ
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // オブジェクトキーを構築: <ユーザーID>/<仮想ディレクトリパス>/<ファイル名>
    objectKey := buildObjectKey(userID, virtualPath, header.Filename)

    err = storage.SaveFile(objectKey, file, header.Size)
    if err != nil {
        http.Error(w, "Upload failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Uploaded: %s\n", objectKey)
}

// オブジェクトキーを構築する関数
func buildObjectKey(userID, virtualPath, filename string) string {
    if virtualPath == "" {
        return fmt.Sprintf("%s/%s", userID, filename)
    }
    // パスの正規化
    virtualPath = strings.Trim(virtualPath, "/")
    return fmt.Sprintf("%s/%s/%s", userID, virtualPath, filename)
}

// フォルダ作成ハンドラー（ダミーオブジェクト使用）
func handleMakeDir(w http.ResponseWriter, r *http.Request) {
    // CORS設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    // 認証済みユーザーIDを取得
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        http.Error(w, "User ID not found in token", http.StatusBadRequest)
        return
    }
    
    folderPath := r.FormValue("path")
    if folderPath == "" {
        http.Error(w, "Missing path parameter", http.StatusBadRequest)
        return
    }

    // .keepオブジェクトを作成して空フォルダを表現
    objectKey := buildObjectKey(userID, folderPath, ".keep")
    
    // 空の内容で.keepファイルを作成
    emptyContent := strings.NewReader("")
    err := storage.SaveFile(objectKey, emptyContent, 0)
    if err != nil {
        http.Error(w, "Failed to create folder: "+err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Folder created: %s/%s\n", userID, folderPath)
}

// 複数ファイルアップロードハンドラー
func handleMultipleUpload(w http.ResponseWriter, r *http.Request) {
    // CORS設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    // ユーザーIDと仮想パスを取得
    userID := r.FormValue("userID")
    if userID == "" {
        userID = "default"
    }
    
    virtualPath := r.FormValue("path")

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

        // オブジェクトキーを構築（フォルダ構造を維持）
        objectKey := buildObjectKey(userID, virtualPath, fileHeader.Filename)
        
        err = storage.SaveFile(objectKey, file, fileHeader.Size)
        file.Close()

        if err != nil {
            errors = append(errors, fmt.Sprintf("Failed to save %s: %v", objectKey, err))
        } else {
            uploadedFiles = append(uploadedFiles, objectKey)
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
    
    // 認証済みユーザーIDを取得
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        http.Error(w, "User ID not found in token", http.StatusBadRequest)
        return
    }
    
    path := r.URL.Query().Get("path")
    filename := r.URL.Query().Get("filename")
    
    if filename == "" {
        http.Error(w, "Missing filename parameter", http.StatusBadRequest)
        return
    }

    // オブジェクトキーを構築
    objectKey := buildObjectKey(userID, path, filename)

    reader, err := storage.GetFile(objectKey)
    if err != nil {
        http.Error(w, "Download failed: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer reader.Close()

    w.Header().Set("Content-Disposition", "attachment; filename="+filename)
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

    // 認証済みユーザーIDを取得
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        http.Error(w, "User ID not found in token", http.StatusBadRequest)
        return
    }
    
    path := r.URL.Query().Get("path")

    // 階層構造でファイル/フォルダ一覧を取得
    items, err := storage.ListUserFiles(userID, path)
    if err != nil {
        http.Error(w, "Failed to list files: "+err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(items)
}

// ファイル削除ハンドラー
func handleDelete(w http.ResponseWriter, r *http.Request) {
    // CORS設定
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    if r.Method != http.MethodDelete {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    // 認証済みユーザーIDを取得
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        http.Error(w, "User ID not found in token", http.StatusBadRequest)
        return
    }
    
    path := r.URL.Query().Get("path")
    filename := r.URL.Query().Get("filename")
    
    if filename == "" {
        http.Error(w, "Missing filename parameter", http.StatusBadRequest)
        return
    }

    // オブジェクトキーを構築
    objectKey := buildObjectKey(userID, path, filename)

    err := storage.DeleteFile(objectKey)
    if err != nil {
        http.Error(w, "Delete failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Deleted: %s\n", objectKey)
}