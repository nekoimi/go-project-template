package jwtutil

import "github.com/golang-jwt/jwt/v5"

// ValidateToken 验证 JWT token 并返回用户 ID
func ValidateToken(tokenStr, secret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrTokenInvalidClaims
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", jwt.ErrTokenInvalidClaims
	}

	return userID, nil
}
