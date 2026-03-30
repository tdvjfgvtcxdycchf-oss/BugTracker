package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type JWTClaims struct {
	Sub   string `json:"sub"`   // user id
	Email string `json:"email"` // user email
	Ver   int    `json:"ver"`   // jwt version
	Iat   int64  `json:"iat"`
	Exp   int64  `json:"exp"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func base64URLEncode(input []byte) string {
	return base64.RawURLEncoding.EncodeToString(input)
}

func base64URLDecode(input string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(input)
}

func GenerateJWT(secret string, userID int, email string, version int, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("JWT secret is empty")
	}

	now := time.Now().Unix()
	claims := JWTClaims{
		Sub:   strconv.Itoa(userID),
		Email: email,
		Ver:   version,
		Iat:   now,
		Exp:   now + int64(ttl.Seconds()),
	}

	header := jwtHeader{Alg: "HS256", Typ: "JWT"}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal jwt claims: %w", err)
	}

	encodedHeader := base64URLEncode(headerBytes)
	encodedClaims := base64URLEncode(claimsBytes)

	unsigned := encodedHeader + "." + encodedClaims
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(unsigned))
	signature := mac.Sum(nil)

	encodedSignature := base64URLEncode(signature)
	return unsigned + "." + encodedSignature, nil
}

func VerifyJWT(secret string, token string) (JWTClaims, error) {
	if secret == "" {
		return JWTClaims{}, errors.New("JWT secret is empty")
	}

	parts := splitToken(token)
	if len(parts) != 3 {
		return JWTClaims{}, errors.New("invalid token format")
	}

	encodedHeader, encodedClaims, encodedSignature := parts[0], parts[1], parts[2]
	unsigned := encodedHeader + "." + encodedClaims

	signatureBytes, err := base64URLDecode(encodedSignature)
	if err != nil {
		return JWTClaims{}, errors.New("invalid token signature encoding")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(unsigned))
	expected := mac.Sum(nil)

	if !hmac.Equal(signatureBytes, expected) {
		return JWTClaims{}, errors.New("invalid token signature")
	}

	claimsBytes, err := base64URLDecode(encodedClaims)
	if err != nil {
		return JWTClaims{}, errors.New("invalid token claims encoding")
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return JWTClaims{}, errors.New("invalid token claims json")
	}

	if claims.Exp == 0 || claims.Exp < time.Now().Unix() {
		return JWTClaims{}, errors.New("token expired")
	}
	if claims.Sub == "" {
		return JWTClaims{}, errors.New("token sub is empty")
	}
	return claims, nil
}

func splitToken(token string) []string {
	res := make([]string, 0, 3)
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			res = append(res, token[start:i])
			start = i + 1
			if len(res) == 2 {
				break
			}
		}
	}
	if len(res) < 2 {
		return []string{token}
	}
	res = append(res, token[start:])
	return res
}
