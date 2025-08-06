package storage

import (
    "context"
    "errors"
    "io"
    "log"
    "strings"
    "time"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

var (
    client     *minio.Client
    bucketName = "files"
    modTime    = time.Now()
)

type FileInfo struct {
    Name         string    `json:"name"`
    Size         int64     `json:"size"`
    LastModified time.Time `json:"lastModified"`
    ContentType  string    `json:"contentType"`
}

func InitMinIO() error {
    var err error
    client, err = minio.New("localhost:9000", &minio.Options{
        Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
        Secure: false,
    })
    if err != nil {
        return err
    }

    exists, err := client.BucketExists(context.Background(), bucketName)
    if err != nil {
        return err
    }
    if !exists {
        err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
        if err != nil {
            return err
        }
    }

    log.Println("Connected to MinIO")
    return nil
}

func ListFiles() ([]FileInfo, error) {
    var files []FileInfo
    
    ctx := context.Background()
    objectCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
        Recursive: true,
    })

    for object := range objectCh {
        if object.Err != nil {
            return nil, object.Err
        }

        files = append(files, FileInfo{
            Name:         object.Key,
            Size:         object.Size,
            LastModified: object.LastModified,
            ContentType:  object.ContentType,
        })
    }

    return files, nil
}

// フォルダ構造付きでファイル一覧を取得
func ListFilesWithFolders() (map[string]interface{}, error) {
    var allFiles []FileInfo
    
    ctx := context.Background()
    objectCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
        Recursive: true,
    })

    for object := range objectCh {
        if object.Err != nil {
            return nil, object.Err
        }

        allFiles = append(allFiles, FileInfo{
            Name:         object.Key,
            Size:         object.Size,
            LastModified: object.LastModified,
            ContentType:  object.ContentType,
        })
    }

    // フォルダ構造を構築
    return buildFolderStructure(allFiles), nil
}

// ファイルリストからフォルダ構造を構築
func buildFolderStructure(files []FileInfo) map[string]interface{} {
    root := make(map[string]interface{})
    
    for _, file := range files {
        parts := strings.Split(file.Name, "/")
        current := root
        
        // フォルダ部分を処理
        for _, part := range parts[:len(parts)-1] {
            if current[part] == nil {
                current[part] = make(map[string]interface{})
            }
            if folder, ok := current[part].(map[string]interface{}); ok {
                current = folder
            }
        }
        
        // ファイル部分を処理
        fileName := parts[len(parts)-1]
        current[fileName] = file
    }
    
    return root
}

// DeleteFile ファイルを削除
func DeleteFile(filename string) error {
    err := client.RemoveObject(context.Background(), bucketName, filename, minio.RemoveObjectOptions{})
    if err != nil {
        return err
    }
    modTime = time.Now()
    return nil
}

func SaveFile(filename string, data io.Reader, size int64) error {
    _, err := client.PutObject(context.Background(), bucketName, filename, data, size, minio.PutObjectOptions{})
    if err == nil {
        modTime = time.Now()
    }
    return err
}

// パス付きでファイルを保存（フォルダ構造対応）
func SaveFileWithPath(path, filename string, data io.Reader, size int64) error {
    // パスを正規化（スラッシュで統一）
    fullPath := normalizePath(path, filename)
    return SaveFile(fullPath, data, size)
}

// パスを正規化する関数
func normalizePath(path, filename string) string {
    if path == "" {
        return filename
    }
    
    // バックスラッシュをスラッシュに変換
    path = strings.ReplaceAll(path, "\\", "/")
    
    // 先頭のスラッシュを削除
    path = strings.TrimPrefix(path, "/")
    
    // 末尾のスラッシュを削除
    path = strings.TrimSuffix(path, "/")
    
    if path == "" {
        return filename
    }
    
    return path + "/" + filename
}

func GetFile(filename string) (io.ReadSeekCloser, error) {
    obj, err := client.GetObject(context.Background(), bucketName, filename, minio.GetObjectOptions{})
    if err != nil {
        return nil, err
    }
    stat, err := obj.Stat()
    if err != nil {
        obj.Close() // リソースリークを防ぐ
        return nil, errors.New("file not found")
    }
    modTime = stat.LastModified
    return obj, nil
}

func LastModified() time.Time {
    return modTime
}
