# MinIO Cloud Storage API 使用例

## 1. ファイルアップロード
```bash
# 基本アップロード
curl -F "file=@test.txt" -F "userID=user123" -F "path=docs" http://app.nitmcr.f5.si:8080/upload

# ルートディレクトリにアップロード
curl -F "file=@test.txt" -F "userID=user123" http://app.nitmcr.f5.si:8080/upload
```

## 2. フォルダ作成
```bash
# フォルダ作成（.keepオブジェクトで空フォルダを表現）
curl -X POST -d "userID=user123&path=docs/reports" http://app.nitmcr.f5.si:8080/mkdir
```

## 3. ファイル一覧取得
```bash
# ユーザーのルートディレクトリ一覧
curl "http://app.nitmcr.f5.si:8080/list?userID=user123"

# 特定フォルダの一覧
curl "http://app.nitmcr.f5.si:8080/list?userID=user123&path=docs"
```

## 4. ファイルダウンロード
```bash
# ファイルダウンロード
curl "http://app.nitmcr.f5.si:8080/download?userID=user123&path=docs&filename=test.txt" -O

# ルートディレクトリのファイル
curl "http://app.nitmcr.f5.si:8080/download?userID=user123&filename=test.txt" -O
```

## 5. ファイル削除
```bash
# ファイル削除
curl -X DELETE "http://app.nitmcr.f5.si:8080/delete?userID=user123&path=docs&filename=test.txt"
```

## PowerShell例

```powershell
# ファイルアップロード
$form = @{
    file = Get-Item "test.txt"
    userID = "user123"
    path = "docs/reports"
}
Invoke-RestMethod -Uri "http://app.nitmcr.f5.si:8080/upload" -Method Post -Form $form

# フォルダ作成
$body = @{
    userID = "user123"
    path = "docs/new_folder"
}
Invoke-RestMethod -Uri "http://app.nitmcr.f5.si:8080/mkdir" -Method Post -Body $body

# ファイル一覧
$files = Invoke-RestMethod -Uri "http://app.nitmcr.f5.si:8080/list?userID=user123&path=docs"
$files | ConvertTo-Json

# ファイルダウンロード
Invoke-WebRequest -Uri "http://app.nitmcr.f5.si:8080/download?userID=user123&path=docs&filename=test.txt" -OutFile "downloaded_test.txt"

# ファイル削除
Invoke-RestMethod -Uri "http://app.nitmcr.f5.si:8080/delete?userID=user123&path=docs&filename=test.txt" -Method Delete
```

## オブジェクトキー命名規則

- ユーザーファイル: `user123/docs/report.pdf`
- 空フォルダ: `user123/docs/reports/.keep`
- ルートファイル: `user123/readme.txt`

## レスポンス例

### ファイル一覧レスポンス
```json
{
  "path": "docs",
  "userID": "user123",
  "files": ["report.pdf", "summary.txt"],
  "folders": ["reports", "images"]
}
```
