package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrExpiredRefreshToken = errors.New("expired refresh token")
)

type TokenService interface {
	GenerateTokens(ctx context.Context, userID int64) (accessToken, refreshToken string, err error)
	ParseAccessToken(tokenStr string) (*Claims, error)
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)
}

type tokenService struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	redis      *redis.Client
}

func NewTokenService(secret string, accessTTL, refreshTTL time.Duration, rdb *redis.Client) TokenService {
	return &tokenService{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		redis:      rdb,
	}
}

func (s *tokenService) parseRefreshClaims(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidRefreshToken
	}

	if claims.ID == "" {
		return nil, ErrInvalidRefreshToken
	}

	return claims, nil
}

func (s *tokenService) GenerateTokens(ctx context.Context, userID int64) (string, string, error) {
	now := time.Now()

	// ----- access token -----
	accessClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.secret)
	if err != nil {
		return "", "", err
	}

	// ----- refresh token -----
	jti := uuid.NewString()
	refreshClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti, // ใช้แยกว่าเป็น refresh + ผูกกับ redis
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.secret)
	if err != nil {
		return "", "", err
	}

	// เก็บ jti -> userID ใน redis (ไว้เช็คตอน refresh / revoke)
	key := "refresh:" + jti
	if err := s.redis.Set(ctx, key, userID, s.refreshTTL).Err(); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *tokenService) ParseAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		// กัน alg แปลก ๆ
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func (s *tokenService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	// 1) parse + verify refresh JWT
	claims, err := s.parseRefreshClaims(refreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			return "", "", ErrInvalidRefreshToken
		}
		return "", "", err
	}

	// เช็คหมดอายุ (กันเหนียว)
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return "", "", ErrExpiredRefreshToken
	}

	// access token ที่คุณออกไม่มี ID (jti)
	// refresh token เท่านั้นที่มี ID และถูกเก็บใน redis
	if claims.ID == "" {
		return "", "", ErrInvalidRefreshToken
	}

	// 2) เช็คใน redis ว่า jti นี้ยัง valid มั้ย
	key := "refresh:" + claims.ID

	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return "", "", err
	}
	if exists == 0 {
		// ไม่มีใน redis = หมดอายุ / revoke / ใช้ไปแล้ว
		return "", "", ErrInvalidRefreshToken
	}

	// single-use: ลบ jti เดิมออก
	_ = s.redis.Del(ctx, key).Err()

	// 3) ออก access + refresh ใหม่ ด้วย userID เดิม
	newAccess, newRefresh, err := s.GenerateTokens(ctx, claims.UserID)
	if err != nil {
		return "", "", err
	}

	return newAccess, newRefresh, nil
}
