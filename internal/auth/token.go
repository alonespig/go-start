package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 是签发到 token 里的身份信息，可按业务扩展（如租户、权限位）。
type Claims struct {
	UserID int64    `json:"uid"`
	Roles  []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// TokenService 负责 JWT 的签发与校验，供登录接口和鉴权中间件复用。
type TokenService struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

func NewTokenService(secret, issuer string, ttl time.Duration) (*TokenService, error) {
	if len(secret) < 16 {
		return nil, errors.New("jwt_secret 长度至少 16 个字符")
	}
	if ttl <= 0 {
		return nil, errors.New("jwt_expire 必须大于 0")
	}
	if issuer == "" {
		issuer = "go-starter"
	}
	return &TokenService{secret: []byte(secret), issuer: issuer, ttl: ttl}, nil
}

// Issue 为用户签发 token；登录接口拿到用户后调用。
func (s *TokenService) Issue(userID int64, roles []string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

// Parse 校验签名、算法、签发方与过期时间，返回 claims。
func (s *TokenService) Parse(token string) (*Claims, error) {
	claims := new(Claims)
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return s.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
	)
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("token 无效")
	}
	return claims, nil
}
