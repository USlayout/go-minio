package storage

import (
    "context"
    "errors"
    "io"
    "log"
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
