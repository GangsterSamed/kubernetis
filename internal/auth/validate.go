package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/polzovatel/todo-learning/internal/models"
)

func (s *JWTSigner) ValidateToken(token string) (*models.Claims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{s.jwtMethodName()}),
		jwt.WithIssuedAt(),
		jwt.WithExpirationRequired(),
	)

	var claims models.Claims
	_, err := parser.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		switch s.alg {
		case "HS256":
			return s.hsSectet, nil
		case "RS256":
			return s.rsaPub, nil
		default:
			return nil, fmt.Errorf("unsupported alg %s", s.alg)
		}
	})
	if err != nil {
		return nil, err
	}

	return &claims, nil
}
