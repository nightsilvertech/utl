package jsonwebtoken

import (
	"context"
	"errors"
	"fmt"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	jwtlib "github.com/golang-jwt/jwt"
	"google.golang.org/grpc/metadata"
	"strings"
	"time"
)

const (
	BearerSeparator   = ` `
	BearerPrefix      = `Bearer`
	BearerValidLength = 2
	BearerTokenIndex  = 1
)

var (
	ErrorInvalidTokenGiven     = errors.New("invalid bearer token")
	ErrorNoTokenGiven          = errors.New("no bearer token given")
	errorUnexpectSigningMethod = errors.New("unexpected signing method")
	errorClaimNotOK            = errors.New("claim not ok or token invalid")
	errorClaimKeyNotFound      = errors.New("fail claim key or key not recognized")
	errorClaimCastingFailed    = errors.New("claim key found but cast error")
)

func BearerTokenMetadataToContext() grpctransport.ServerRequestFunc {
	return func(ctx context.Context, md metadata.MD) context.Context {
		jwt, ok := md["authorization"]
		if !ok {
			return ctx
		}
		if ok {
			ctx = context.WithValue(ctx, "authorization", jwt[0])
		}
		return ctx
	}
}

func ContextToBearerTokenMetadata() grpctransport.ClientRequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		requestID, ok := ctx.Value("authorization").(string)
		if ok {
			(*md)["authorization"] = []string{requestID}
		}
		return ctx
	}
}

func Parser(stringToken, secret string) (*jwtlib.Token, error) {
	token, err := jwtlib.Parse(stringToken, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, errorUnexpectSigningMethod
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func Claim(token *jwtlib.Token, key string) (res string, err error) {
	if claims, ok := token.Claims.(jwtlib.MapClaims); ok && token.Valid {
		val, claimOk := claims[key]
		if claimOk {
			casted, castOk := val.(string)
			if castOk {
				return casted, nil
			}
			return res, errorClaimCastingFailed
		}
		return res, errorClaimKeyNotFound
	}
	return res, errorClaimNotOK
}

func ExtractKeys(stringToken, secret string, keys []string) (map[string]string, error) {
	claims := make(map[string]string)
	token, err := Parser(stringToken, secret)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		claimed, err := Claim(token, key)
		if err != nil {
			return nil, err
		}
		claims[key] = claimed
	}
	return claims, nil
}

func GenerateJWT(data map[string]string, expired time.Duration, secret string) (res string, err error) {
	claims := jwtlib.MapClaims{}
	claims["exp"] = time.Now().Add(expired).Unix()
	for k, v := range data {
		claims[k] = v
	}
	jwtToken := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		return res, err
	}

	refreshToken := jwtlib.New(jwtlib.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwtlib.MapClaims)
	rtClaims["exp"] = time.Now().Add(expired * 2).Unix()
	for k, v := range data {
		rtClaims[k] = v
	}
	refresh, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return res, err
	}
	res = fmt.Sprintf("%s:%s", token, refresh)
	return res, nil
}

func bearer(ctx context.Context) (string, error) {
	var tokenMetadata []string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		tokenMetadata = md.Get("authorization")
	}
	if len(tokenMetadata) == 0 {
		return "", errors.New("no token given")
	}
	token := strings.Split(tokenMetadata[0], " ")
	if len(token) != 2 || token[0] != "Bearer" {
		return "", errors.New("invalid bearer token")
	}
	return token[1], nil
}

func ExtractKeysFromCtx(ctx context.Context, secret string, keys []string) (res map[string]string, err error) {
	bearer, err := bearer(ctx)
	if err != nil {
		return res, err
	}
	token, err := Parser(bearer, secret)
	if err != nil {
		return res, err
	}
	claims := make(map[string]string)
	for _, key := range keys {
		claimed, err := Claim(token, key)
		if err != nil {
			return res, err
		}
		claims[key] = claimed
	}
	return claims, nil
}
