// Package jwt 提供 HS256 token 的签发与校验。
package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken 表示 token 缺失、被篡改、算法不符、过期或签发者不匹配。
var ErrInvalidToken = errors.New("invalid token")

// Claims 是项目自定义的 JWT 载荷：user_id 同时写入 uid 与 sub，便于透传与检索。
type Claims struct {
	UserID uint64 `json:"uid"`
	jwtlib.RegisteredClaims
}

// Manager 负责 HS256 的签发与校验，now 可注入以便测试过期逻辑。
type Manager struct {
	secret []byte
	ttl    time.Duration
	issuer string
	now    func() time.Time
}

// NewManager 构造 JWT 管理器。
func NewManager(secret string, ttl time.Duration, issuer string) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl, issuer: issuer, now: time.Now}
}

// Generate 为指定用户签发 HS256 token。
func (m *Manager) Generate(userID uint64) (string, error) {
	now := m.now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   strconv.FormatUint(userID, 10),
			Issuer:    m.issuer,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.ttl)),
		},
	}

	signed, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// Parse 校验签名、算法、签发者与有效期，返回载荷。
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwtlib.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwtlib.Token) (any, error) {
			if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.secret, nil
		},
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
		jwtlib.WithIssuer(m.issuer),
		jwtlib.WithTimeFunc(m.now),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
