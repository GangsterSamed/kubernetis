package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/polzovatel/todo-learning/config"
	"github.com/polzovatel/todo-learning/internal/models"
	"time"
)

type JWTSigner struct {
	alg        string
	hsSectet   []byte
	rsaPriv    *rsa.PrivateKey
	rsaPub     *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTSigner(cfg *config.Config) (*JWTSigner, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	s := &JWTSigner{
		alg:        cfg.JWTAlg,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}

	switch s.alg {
	case "HS256":
		if cfg.JWTSecret == "" {
			return nil, errors.New("JWTSecret is empty")
		}
		s.hsSectet = []byte(cfg.JWTSecret)

	case "RS256":
		if cfg.JWTPrivatePEM == "" || cfg.JWTPublicPEM == "" {
			return nil, errors.New("RS256 requires JWT_PRIVATE_PEM and JWT_PUBLIC_PEM")
		}
		var err error
		s.rsaPriv, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.JWTPrivatePEM))
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		s.rsaPub, err = jwt.ParseRSAPublicKeyFromPEM([]byte(cfg.JWTPublicPEM))
		if err != nil {
			return nil, fmt.Errorf("parse public key: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown signing algorithm: %s", s.alg)
	}

	return s, nil
}

func (s *JWTSigner) GenerateAccessToken(userID, email, role string) (string, error) {
	now := time.Now()
	claims := &models.Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   "access_token",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	return s.sign(claims)
}

func (s *JWTSigner) GenerateRefreshToken(userID, email, role string) (string, error) {
	now := time.Now()
	claims := &models.Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   "refresh_token",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	return s.sign(claims)
}

func (s *JWTSigner) sign(claims *models.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.GetSigningMethod(s.jwtMethodName()), claims)
	switch s.alg {
	case "HS256":
		return token.SignedString(s.hsSectet)
	case "RS256":
		return token.SignedString(s.rsaPriv)
	default:
		return "", fmt.Errorf("unknown signing algorithm: %s", s.alg)
	}
}

func (s *JWTSigner) jwtMethodName() string {
	switch s.alg {
	case "HS256":
		return jwt.SigningMethodHS256.Name
	case "RS256":
		return jwt.SigningMethodRS256.Name
	default:
		return ""
	}
}
