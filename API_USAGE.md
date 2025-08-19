# MinIO Cloud Storage API 使用例（JWT認証付き）

## 認証

### 1. ログイン
```bash
# ユーザーログイン
curl -X POST -H "Content-Type: application/json" \
  -d '{"userID":"user123","password":"password123"}' \
  https://app.nitmcr.f5.si/auth/login
```

レスポンス例：
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "userID": "user123",
    "username": "testuser",
    "email": "test@example.com",
    "role": "user"
  },
  "expiresIn": 86400
}
```

### 2. トークンリフレッシュ
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"refreshToken":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}' \
  https://app.nitmcr.f5.si/auth/refresh
```

### 3. ユーザー情報取得
```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  https://app.nitmcr.f5.si/auth/me
```

## ファイル操作（認証が必要）

### 1. ファイルアップロード
```bash
# 基本アップロード（Authorizationヘッダー使用）
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "file=@test.txt" -F "path=docs" \
  https://app.nitmcr.f5.si/upload

# クエリパラメータでトークン指定
curl -F "file=@test.txt" -F "path=docs" \
  "https://app.nitmcr.f5.si/upload?token=YOUR_ACCESS_TOKEN"

# ルートディレクトリにアップロード
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "file=@test.txt" \
  https://app.nitmcr.f5.si/upload
```

### 1-2. 複数ファイルアップロード
```bash
# 複数ファイルの一括アップロード
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "files=@file1.txt" -F "files=@file2.txt" -F "files=@file3.pdf" \
  -F "path=documents" \
  https://app.nitmcr.f5.si/upload-multiple
```

### 1-3. フォルダアップロード
```bash
# フォルダ内のファイルを一括アップロード（フォルダ構造を保持）
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "files=@folder/subfolder/file1.txt" \
  -F "files=@folder/file2.txt" \
  -F "path=projects" \
  https://app.nitmcr.f5.si/upload-folder
```

### 2. フォルダ作成
```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -X POST -d "path=docs/reports" \
  https://app.nitmcr.f5.si/mkdir
```

### 3. ファイル一覧取得
```bash
# ユーザーのルートディレクトリ一覧
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "https://app.nitmcr.f5.si/list"

# 特定フォルダの一覧
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "https://app.nitmcr.f5.si/list?path=docs"
```

### 3-2. ファイル詳細一覧取得
```bash
# 詳細情報付きでファイル一覧取得（ファイルサイズ、更新日時、統計情報含む）
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "https://app.nitmcr.f5.si/list-details?path=docs"
```

### 3-3. フォルダ構造一覧取得
```bash
# ユーザーのフォルダ構造を詳細情報付きで取得
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "https://app.nitmcr.f5.si/list-folders"
```

### 4. ファイルダウンロード
```bash
# ファイルダウンロード
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "https://app.nitmcr.f5.si/download?path=docs&filename=test.txt" -O

# ルートディレクトリのファイル
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "https://app.nitmcr.f5.si/download?filename=test.txt" -O
```

### 5. ファイル削除
```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -X DELETE "https://app.nitmcr.f5.si/delete?path=docs&filename=test.txt"
```

## ファイル情報取得（認証が必要）

### 1. ファイル詳細情報取得
```bash
# ファイルの詳細情報（サイズ、更新日時、コンテンツタイプ）を取得
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "https://app.nitmcr.f5.si/info?path=docs&filename=report.pdf"
```

### 2. ファイルサイズ取得
```bash
# ファイルサイズのみを取得（人間が読みやすい形式も含む）
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "https://app.nitmcr.f5.si/size?path=docs&filename=report.pdf"
```

### 3. ファイルメタデータ取得
```bash
# ファイルの完全なメタデータ（ETag、バージョンID、ストレージクラスなど）を取得
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "https://app.nitmcr.f5.si/metadata?path=docs&filename=report.pdf"
```

## 管理者専用エンドポイント

### ユーザー管理
```bash
# ユーザー一覧取得（管理者のみ）
curl -H "Authorization: Bearer ADMIN_ACCESS_TOKEN" \
  "https://app.nitmcr.f5.si/admin/users"
```

## PowerShell例

```powershell
# ログイン
$loginData = @{
    userID = "user123"
    password = "password123"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "https://app.nitmcr.f5.si/auth/login" -Method Post -Body $loginData -ContentType "application/json"
$accessToken = $response.accessToken

# ファイルアップロード
$headers = @{ "Authorization" = "Bearer $accessToken" }
$form = @{
    file = Get-Item "test.txt"
    path = "docs/reports"
}
Invoke-RestMethod -Uri "https://app.nitmcr.f5.si/upload" -Method Post -Form $form -Headers $headers

# 複数ファイルアップロード
$files = @("file1.txt", "file2.txt", "file3.pdf")
$form = @{
    path = "documents"
}
foreach ($file in $files) {
    $form["files"] = Get-Item $file
}
Invoke-RestMethod -Uri "https://app.nitmcr.f5.si/upload-multiple" -Method Post -Form $form -Headers $headers

# フォルダ作成
$body = @{ path = "docs/new_folder" }
Invoke-RestMethod -Uri "https://app.nitmcr.f5.si/mkdir" -Method Post -Body $body -Headers $headers

# ファイル一覧
$files = Invoke-RestMethod -Uri "https://app.nitmcr.f5.si/list?path=docs" -Headers $headers
$files | ConvertTo-Json

# 詳細ファイル一覧
$detailedFiles = Invoke-RestMethod -Uri "https://app.nitmcr.f5.si/list-details?path=docs" -Headers $headers
$detailedFiles | ConvertTo-Json

# ファイル情報取得
$fileInfo = Invoke-RestMethod -Uri "https://app.nitmcr.f5.si/info?path=docs&filename=report.pdf" -Headers $headers
$fileInfo | ConvertTo-Json

# ファイルサイズ取得
$fileSize = Invoke-RestMethod -Uri "https://app.nitmcr.f5.si/size?path=docs&filename=report.pdf" -Headers $headers
Write-Output "File size: $($fileSize.sizeHuman)"

# ファイルダウンロード
Invoke-WebRequest -Uri "https://app.nitmcr.f5.si/download?path=docs&filename=test.txt" -Headers $headers -OutFile "downloaded_test.txt"

# ファイル削除
Invoke-RestMethod -Uri "https://app.nitmcr.f5.si/delete?path=docs&filename=test.txt" -Method Delete -Headers $headers
```

## JavaScript/Fetch API例

```javascript
// ログイン
async function login(userID, password) {
    const response = await fetch('https://app.nitmcr.f5.si/auth/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ userID, password })
    });
    
    if (!response.ok) {
        throw new Error('Login failed');
    }
    
    const data = await response.json();
    localStorage.setItem('accessToken', data.accessToken);
    localStorage.setItem('refreshToken', data.refreshToken);
    return data;
}

// ファイルアップロード
async function uploadFile(file, path = '') {
    const token = localStorage.getItem('accessToken');
    const formData = new FormData();
    formData.append('file', file);
    if (path) formData.append('path', path);
    
    const response = await fetch('https://app.nitmcr.f5.si/upload', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`
        },
        body: formData
    });
    
    if (!response.ok) {
        throw new Error('Upload failed');
    }
    
    return response.text();
}

// 複数ファイルアップロード
async function uploadMultipleFiles(files, path = '') {
    const token = localStorage.getItem('accessToken');
    const formData = new FormData();
    
    files.forEach(file => {
        formData.append('files', file);
    });
    
    if (path) formData.append('path', path);
    
    const response = await fetch('https://app.nitmcr.f5.si/upload-multiple', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`
        },
        body: formData
    });
    
    if (!response.ok) {
        throw new Error('Multiple upload failed');
    }
    
    return response.json();
}

// ファイル一覧取得
async function listFiles(path = '') {
    const token = localStorage.getItem('accessToken');
    const url = path ? 
        `https://app.nitmcr.f5.si/list?path=${encodeURIComponent(path)}` :
        'https://app.nitmcr.f5.si/list';
    
    const response = await fetch(url, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    if (!response.ok) {
        throw new Error('Failed to list files');
    }
    
    return response.json();
}

// 詳細ファイル一覧取得
async function listFilesWithDetails(path = '') {
    const token = localStorage.getItem('accessToken');
    const url = path ? 
        `https://app.nitmcr.f5.si/list-details?path=${encodeURIComponent(path)}` :
        'https://app.nitmcr.f5.si/list-details';
    
    const response = await fetch(url, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    if (!response.ok) {
        throw new Error('Failed to list files with details');
    }
    
    return response.json();
}

// ファイル情報取得
async function getFileInfo(filename, path = '') {
    const token = localStorage.getItem('accessToken');
    const url = path ? 
        `https://app.nitmcr.f5.si/info?path=${encodeURIComponent(path)}&filename=${encodeURIComponent(filename)}` :
        `https://app.nitmcr.f5.si/info?filename=${encodeURIComponent(filename)}`;
    
    const response = await fetch(url, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    if (!response.ok) {
        throw new Error('Failed to get file info');
    }
    
    return response.json();
}

// ファイルサイズ取得
async function getFileSize(filename, path = '') {
    const token = localStorage.getItem('accessToken');
    const url = path ? 
        `https://app.nitmcr.f5.si/size?path=${encodeURIComponent(path)}&filename=${encodeURIComponent(filename)}` :
        `https://app.nitmcr.f5.si/size?filename=${encodeURIComponent(filename)}`;
    
    const response = await fetch(url, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    if (!response.ok) {
        throw new Error('Failed to get file size');
    }
    
    return response.json();
}

// 使用例
login('user123', 'password123')
    .then(data => {
        console.log('Logged in:', data.user);
        return listFilesWithDetails();
    })
    .then(files => {
        console.log('Files with details:', files);
        console.log('Total files:', files.statistics.totalFiles);
        console.log('Total size:', files.statistics.totalSizeHuman);
        return getFileInfo('report.pdf', 'docs');
    })
    .then(fileInfo => {
        console.log('File info:', fileInfo);
    })
    .catch(error => {
        console.error('Error:', error);
    });
```

## テストユーザー

### 一般ユーザー
- **ユーザーID**: `user123`
- **パスワード**: `password123`
- **権限**: ファイル操作のみ

### 管理者ユーザー
- **ユーザーID**: `admin`
- **パスワード**: `adminpass`
- **権限**: 全ての操作 + 管理者機能

## セキュリティ機能

- **JWT トークンベース認証**: アクセストークン（24時間有効）
- **リフレッシュトークン**: 7日間有効
- **ユーザー分離**: 各ユーザーは自分のファイルのみアクセス可能
- **権限ベースアクセス制御**: 管理者専用エンドポイント
- **CORS対応**: クロスオリジンリクエスト対応

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

### 詳細ファイル一覧レスポンス
```json
{
  "path": "docs",
  "userID": "user123",
  "files": [
    {
      "name": "report.pdf",
      "size": 2097152,
      "lastModified": "2025-08-19T10:30:00Z",
      "contentType": "application/pdf"
    },
    {
      "name": "summary.txt",
      "size": 1024,
      "lastModified": "2025-08-19T09:15:00Z",
      "contentType": "text/plain"
    }
  ],
  "folders": [
    {
      "name": "reports",
      "type": "folder",
      "itemCount": 5,
      "lastModified": "2025-08-19T11:00:00Z"
    }
  ],
  "statistics": {
    "totalFiles": 2,
    "totalFolders": 1,
    "totalSize": 2098176,
    "totalSizeHuman": "2.0 MB",
    "latestModified": "2025-08-19T11:00:00Z"
  }
}
```

### ファイル情報レスポンス
```json
{
  "name": "user123/docs/report.pdf",
  "size": 2097152,
  "lastModified": "2025-08-19T10:30:00Z",
  "contentType": "application/pdf"
}
```

### ファイルサイズレスポンス
```json
{
  "filename": "report.pdf",
  "path": "docs",
  "size": 2097152,
  "sizeHuman": "2.0 MB"
}
```

### 複数ファイルアップロードレスポンス
```json
{
  "uploaded": [
    "user123/documents/file1.txt",
    "user123/documents/file2.txt"
  ],
  "errors": [
    "Failed to save user123/documents/file3.pdf: file too large"
  ],
  "total": 3,
  "success": 2,
  "failed": 1
}
```
