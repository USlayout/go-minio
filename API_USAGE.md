# MinIO Cloud Storage API 使用例（JWT認証付き）

## 認証

### 1. ログイン
```bash
# ユーザーログイン
curl -X POST -H "Content-Type: application/json" \
  -d '{"userID":"user123","password":"password123"}' \
  http://app.nitmcr.f5.si:8080/auth/login
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
  http://app.nitmcr.f5.si:8080/auth/refresh
```

### 3. ユーザー情報取得
```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  http://app.nitmcr.f5.si:8080/auth/me
```

## ファイル操作（認証が必要）

### 1. ファイルアップロード
```bash
# 基本アップロード（Authorizationヘッダー使用）
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "file=@test.txt" -F "path=docs" \
  http://app.nitmcr.f5.si:8080/upload

# クエリパラメータでトークン指定
curl -F "file=@test.txt" -F "path=docs" \
  "http://app.nitmcr.f5.si:8080/upload?token=YOUR_ACCESS_TOKEN"

# ルートディレクトリにアップロード
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "file=@test.txt" \
  http://app.nitmcr.f5.si:8080/upload
```

### 2. フォルダ作成
```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -X POST -d "path=docs/reports" \
  http://app.nitmcr.f5.si:8080/mkdir
```

### 3. ファイル一覧取得
```bash
# ユーザーのルートディレクトリ一覧
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "http://app.nitmcr.f5.si:8080/list"

# 特定フォルダの一覧
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "http://app.nitmcr.f5.si:8080/list?path=docs"
```

### 4. ファイルダウンロード
```bash
# ファイルダウンロード
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "http://app.nitmcr.f5.si:8080/download?path=docs&filename=test.txt" -O

# ルートディレクトリのファイル
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "http://app.nitmcr.f5.si:8080/download?filename=test.txt" -O
```

### 5. ファイル削除
```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -X DELETE "http://app.nitmcr.f5.si:8080/delete?path=docs&filename=test.txt"
```

## 管理者専用エンドポイント

### ユーザー管理
```bash
# ユーザー一覧取得（管理者のみ）
curl -H "Authorization: Bearer ADMIN_ACCESS_TOKEN" \
  "http://app.nitmcr.f5.si:8080/admin/users"
```

## PowerShell例

```powershell
# ログイン
$loginData = @{
    userID = "user123"
    password = "password123"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://app.nitmcr.f5.si:8080/auth/login" -Method Post -Body $loginData -ContentType "application/json"
$accessToken = $response.accessToken

# ファイルアップロード
$headers = @{ "Authorization" = "Bearer $accessToken" }
$form = @{
    file = Get-Item "test.txt"
    path = "docs/reports"
}
Invoke-RestMethod -Uri "http://app.nitmcr.f5.si:8080/upload" -Method Post -Form $form -Headers $headers

# フォルダ作成
$body = @{ path = "docs/new_folder" }
Invoke-RestMethod -Uri "http://app.nitmcr.f5.si:8080/mkdir" -Method Post -Body $body -Headers $headers

# ファイル一覧
$files = Invoke-RestMethod -Uri "http://app.nitmcr.f5.si:8080/list?path=docs" -Headers $headers
$files | ConvertTo-Json

# ファイルダウンロード
Invoke-WebRequest -Uri "http://app.nitmcr.f5.si:8080/download?path=docs&filename=test.txt" -Headers $headers -OutFile "downloaded_test.txt"

# ファイル削除
Invoke-RestMethod -Uri "http://app.nitmcr.f5.si:8080/delete?path=docs&filename=test.txt" -Method Delete -Headers $headers
```

## JavaScript/Fetch API例

```javascript
// ログイン
async function login(userID, password) {
    const response = await fetch('http://app.nitmcr.f5.si:8080/auth/login', {
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
    
    const response = await fetch('http://app.nitmcr.f5.si:8080/upload', {
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

// ファイル一覧取得
async function listFiles(path = '') {
    const token = localStorage.getItem('accessToken');
    const url = path ? 
        `http://app.nitmcr.f5.si:8080/list?path=${encodeURIComponent(path)}` :
        'http://app.nitmcr.f5.si:8080/list';
    
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

// 使用例
login('user123', 'password123')
    .then(data => {
        console.log('Logged in:', data.user);
        return listFiles();
    })
    .then(files => {
        console.log('Files:', files);
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
