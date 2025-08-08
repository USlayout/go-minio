package auth

import (
    "errors"
    "fmt"
    "net/http"
    "strings"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

// JWT秘密鍵（本番環境では環境変数から取得すべき）
var jwtSecret = []byte("your-super-secret-key")

// ユーザー情報構造体
type User struct {
    UserID   string `json:"userID"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
}

// JWTクレーム構造体
type Claims struct {
    UserID   string `json:"userID"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

// ダミーユーザーデータベース（本番環境ではデータベースを使用）
var users = map[string]User{
    "user123": {
        UserID:   "user123",
        Username: "testuser",
        Email:    "test@example.com",
        Role:     "user",
    },
    "admin": {
        UserID:   "admin",
        Username: "admin",
        Email:    "admin@example.com",
        Role:     "admin",
    },
}

// パスワードデータベース（本番環境ではハッシュ化したパスワードを使用）
var passwords = map[string]string{
    "user123": "password123",
    "admin":   "adminpass",
}

// JWT トークンを生成
func GenerateToken(user User) (string, error) {
    // トークンの有効期限を24時間に設定
    expirationTime := time.Now().Add(24 * time.Hour)
    
    // クレームを作成
    claims := &Claims{
        UserID:   user.UserID,
        Username: user.Username,
        Email:    user.Email,
        Role:     user.Role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "minio-cloud-storage",
        },
    }
    
    // トークンを作成
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    
    // 署名付きトークン文字列を生成
    tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        return "", err
    }
    
    return tokenString, nil
}

// JWT トークンを検証
func ValidateToken(tokenString string) (*Claims, error) {
    // トークンを解析
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // 署名方法を確認
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return jwtSecret, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    // クレームを取得
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("invalid token")
}

// ユーザー認証
func AuthenticateUser(userID, password string) (*User, error) {
    // ユーザーの存在確認
    user, exists := users[userID]
    if !exists {
        return nil, errors.New("user not found")
    }
    
    // パスワード確認
    if passwords[userID] != password {
        return nil, errors.New("invalid password")
    }
    
    return &user, nil
}

// HTTP リクエストからJWTトークンを取得
func GetTokenFromRequest(r *http.Request) string {
    // Authorization ヘッダーから取得
    authHeader := r.Header.Get("Authorization")
    if authHeader != "" {
        // "Bearer <token>" 形式から抽出
        if strings.HasPrefix(authHeader, "Bearer ") {
            return strings.TrimPrefix(authHeader, "Bearer ")
        }
    }
    
    // クエリパラメータから取得
    return r.URL.Query().Get("token")
}

// JWT認証ミドルウェア
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // CORS設定
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        // OPTIONSリクエストの処理
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        // トークンを取得
        tokenString := GetTokenFromRequest(r)
        if tokenString == "" {
            http.Error(w, "Missing authorization token", http.StatusUnauthorized)
            return
        }
        
        // トークンを検証
        claims, err := ValidateToken(tokenString)
        if err != nil {
            http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
            return
        }
        
        // リクエストコンテキストにユーザー情報を追加
        r.Header.Set("X-User-ID", claims.UserID)
        r.Header.Set("X-User-Role", claims.Role)
        
        // 次のハンドラーを実行
        next.ServeHTTP(w, r)
    }
}

// 管理者権限チェックミドルウェア
func AdminOnlyMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
        role := r.Header.Get("X-User-Role")
        if role != "admin" {
            http.Error(w, "Admin access required", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// リフレッシュトークン生成
func GenerateRefreshToken(userID string) (string, error) {
    expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7日間有効
    
    claims := &jwt.RegisteredClaims{
        Subject:   userID,
        ExpiresAt: jwt.NewNumericDate(expirationTime),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        Issuer:    "minio-cloud-storage",
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

// リフレッシュトークンの検証とアクセストークンの再生成
func RefreshAccessToken(refreshToken string) (string, error) {
    token, err := jwt.ParseWithClaims(refreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    
    if err != nil {
        return "", err
    }
    
    if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
        userID := claims.Subject
        user, exists := users[userID]
        if !exists {
            return "", errors.New("user not found")
        }
        
        return GenerateToken(user)
    }
    
    return "", errors.New("invalid refresh token")
}