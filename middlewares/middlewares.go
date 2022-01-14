package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/nightsilvertech/utl/jsonwebtoken"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

func LoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			var resp interface{}
			req, _ := json.Marshal(request)
			defer func(begin time.Time) {
				level.Info(logger).Log(
					"err", err,
					"took", time.Since(begin),
					"request", string(req),
				)
			}(time.Now())
			resp, err = next(ctx, request)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	}
}

func JwtTestMiddleware(name, phone, secret string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			var tokenMetadata []string
			md, ok := metadata.FromIncomingContext(ctx)
			if ok {
				tokenMetadata = md.Get("authorization")
			}
			if len(tokenMetadata) == 0 {
				return response, errors.New("no token given")
			}
			token := strings.Split(tokenMetadata[0], " ")
			if len(token) != 2 || token[0] != "Bearer" {
				return response, errors.New("invalid bearer token")
			}

			jwtData, err := jsonwebtoken.ExtractKeys(
				token[jsonwebtoken.BearerTokenIndex],
				secret,
				[]string{"name", "phone"},
			)
			if err != nil {
				return response, errors.New(fmt.Sprintf("jwt middleware extract keys %v", err))
			}

			nameJwt := jwtData["name"]
			phoneJwt := jwtData["phone"]

			if nameJwt == name && phoneJwt == phone {
				resp, err := next(ctx, request)
				if err != nil {
					return nil, err
				}
				return resp, nil
			} else {
				return nil, errors.New("data inside jwt not allowed")
			}
		}
	}
}

func CircuitBreakerMiddleware(command string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			var resp interface{}
			var logicErr error
			err = hystrix.Do(command, func() (err error) {
				resp, logicErr = next(ctx, request)
				return logicErr
			}, func(err error) error {
				return err
			})
			if logicErr != nil {
				return nil, logicErr
			}
			if err != nil {
				errMsg := fmt.Sprintf(
					"service %s is busy or unavailable, please try again later",
					command,
				)
				return nil, status.Error(
					codes.Unavailable,
					errors.New(errMsg).Error(),
				)
			}
			return resp, nil
		}
	}
}
