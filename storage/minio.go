package storage

import (
	"context"
	"errors"
	"fmt"
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

type FolderInfo struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"` // "folder"
	ItemCount    int       `json:"itemCount"`
	LastModified time.Time `json:"lastModified"`
}

func InitMinIO() error {
	var err error
	client, err = minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "789632145", ""),
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

// ユーザー別階層構造でファイル/フォルダ一覧を取得（詳細情報付き）
func ListUserFiles(userID, path string) (map[string]interface{}, error) {
	ctx := context.Background()

	// プレフィックスを構築
	prefix := userID + "/"
	if path != "" {
		prefix += strings.Trim(path, "/") + "/"
	}

	objectCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false, // 指定階層のみ
	})

	files := []FileInfo{}
	folders := map[string]bool{}

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		// プレフィックスを除去して相対パスを取得
		relativePath := strings.TrimPrefix(object.Key, prefix)

		// .keepファイルは除外（フォルダ作成用ダミー）
		if strings.HasSuffix(relativePath, ".keep") {
			folderName := strings.TrimSuffix(relativePath, "/.keep")
			if folderName != "" {
				folders[folderName] = true
			}
			continue
		}

		// サブディレクトリの場合
		if strings.Contains(relativePath, "/") {
			folderName := strings.Split(relativePath, "/")[0]
			folders[folderName] = true
		} else if relativePath != "" {
			// ファイルの場合 - 詳細情報を含める
			fileInfo := FileInfo{
				Name:         relativePath,
				Size:         object.Size,
				LastModified: object.LastModified,
				ContentType:  object.ContentType,
			}
			files = append(files, fileInfo)
		}
	}

	// 結果をまとめる
	result := map[string]interface{}{
		"path":    path,
		"userID":  userID,
		"files":   files,
		"folders": getFolderList(folders),
	}

	return result, nil
}

// ユーザー別階層構造でファイル/フォルダ一覧を取得（より詳細な情報付き）
func ListUserFilesWithDetails(userID, path string) (map[string]interface{}, error) {
	ctx := context.Background()

	// プレフィックスを構築
	prefix := userID + "/"
	if path != "" {
		prefix += strings.Trim(path, "/") + "/"
	}

	objectCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false, // 指定階層のみ
	})

	files := []FileInfo{}
	folderMap := map[string]*FolderInfo{}
	var totalSize int64
	var totalFiles int
	var latestModified time.Time

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		// プレフィックスを除去して相対パスを取得
		relativePath := strings.TrimPrefix(object.Key, prefix)

		// .keepファイルは除外（フォルダ作成用ダミー）
		if strings.HasSuffix(relativePath, ".keep") {
			folderName := strings.TrimSuffix(relativePath, "/.keep")
			if folderName != "" {
				if folderMap[folderName] == nil {
					folderMap[folderName] = &FolderInfo{
						Name:         folderName,
						Type:         "folder",
						ItemCount:    0,
						LastModified: object.LastModified,
					}
				}
			}
			continue
		}

		// サブディレクトリの場合
		if strings.Contains(relativePath, "/") {
			folderName := strings.Split(relativePath, "/")[0]
			if folderMap[folderName] == nil {
				folderMap[folderName] = &FolderInfo{
					Name:         folderName,
					Type:         "folder",
					ItemCount:    0,
					LastModified: object.LastModified,
				}
			}
			folderMap[folderName].ItemCount++
			if object.LastModified.After(folderMap[folderName].LastModified) {
				folderMap[folderName].LastModified = object.LastModified
			}
		} else if relativePath != "" {
			// ファイルの場合 - 詳細情報を含める
			fileInfo := FileInfo{
				Name:         relativePath,
				Size:         object.Size,
				LastModified: object.LastModified,
				ContentType:  object.ContentType,
			}
			files = append(files, fileInfo)

			// 統計情報を更新
			totalSize += object.Size
			totalFiles++
			if object.LastModified.After(latestModified) {
				latestModified = object.LastModified
			}
		}
	}

	// フォルダリストを作成
	folders := make([]FolderInfo, 0, len(folderMap))
	for _, folder := range folderMap {
		folders = append(folders, *folder)
	}

	// 結果をまとめる
	result := map[string]interface{}{
		"path":    path,
		"userID":  userID,
		"files":   files,
		"folders": folders,
		"statistics": map[string]interface{}{
			"totalFiles":     totalFiles,
			"totalFolders":   len(folders),
			"totalSize":      totalSize,
			"totalSizeHuman": formatSizeBytes(totalSize),
			"latestModified": latestModified,
		},
	}

	return result, nil
}

// フォルダマップからスライスに変換
func getFolderList(folders map[string]bool) []string {
	var result []string
	for folder := range folders {
		result = append(result, folder)
	}
	return result
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

// ファイル詳細情報を取得
func GetFileInfo(filename string) (*FileInfo, error) {
	ctx := context.Background()

	// オブジェクトの統計情報を取得
	objInfo, err := client.StatObject(ctx, bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	// ファイル名から拡張子に基づいてContent-Typeを推測
	contentType := objInfo.ContentType
	if contentType == "" {
		contentType = "application/octet-stream" // デフォルト
	}

	fileInfo := &FileInfo{
		Name:         objInfo.Key,
		Size:         objInfo.Size,
		LastModified: objInfo.LastModified,
		ContentType:  contentType,
	}

	return fileInfo, nil
}

// ファイルサイズのみを取得
func GetFileSize(filename string) (int64, error) {
	ctx := context.Background()

	objInfo, err := client.StatObject(ctx, bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}

	return objInfo.Size, nil
}

// ファイルメタデータを取得（詳細な情報）
func GetFileMetadata(filename string) (map[string]interface{}, error) {
	ctx := context.Background()

	objInfo, err := client.StatObject(ctx, bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	metadata := map[string]interface{}{
		"name":           objInfo.Key,
		"size":           objInfo.Size,
		"lastModified":   objInfo.LastModified,
		"contentType":    objInfo.ContentType,
		"etag":           objInfo.ETag,
		"versionId":      objInfo.VersionID,
		"isDeleteMarker": objInfo.IsDeleteMarker,
		"metadata":       objInfo.UserMetadata,
		"expires":        objInfo.Expires,
		"storageClass":   objInfo.StorageClass,
	}

	return metadata, nil
}

// ファイルサイズを人間が読みやすい形式にフォーマット
func formatSizeBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
