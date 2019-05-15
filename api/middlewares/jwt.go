package middlewares

import (
	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/conf"
	jwtGo "gopkg.in/dgrijalva/jwt-go.v3"
	"net/http"
	"time"
)

const (
	identityKey = "ID"
	roleKey     = "ROLE"
)

var (
	JWTMiddleware *jwt.GinJWTMiddleware
)

type loginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type AccountMatchResult struct {
	Id   string
	Role string
}

type accountMatcher interface {
	AccountMatch(name, password string) (bool, AccountMatchResult)
}

type JwtMiddleware struct {
	middleware *jwt.GinJWTMiddleware
	matcher    accountMatcher
}

func NewJwtMiddleware(matcher accountMatcher) *jwt.GinJWTMiddleware {
	authenticator := func(c *gin.Context) (interface{}, error) {
		var request loginRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			return nil, jwt.ErrMissingLoginValues
		}

		if request.Name == "" || request.Password == "" {
			return nil, jwt.ErrMissingLoginValues
		}

		match, result := matcher.AccountMatch(request.Name, request.Password)
		if !match {
			return nil, jwt.ErrFailedAuthentication
		}
		return result, nil
	}

	middleware := &jwt.GinJWTMiddleware{
		Realm:             conf.GetConfString(conf.JWTRealmConfKey),
		SigningAlgorithm:  "HS256",
		Key:               conf.GetConfByteSlice(conf.JWTKeyConfKey),
		Timeout:           7 * 24 * time.Hour,
		IdentityHandler:   identityHandler,
		Authenticator:     authenticator,
		PayloadFunc:       payloadFunc,
		SendCookie:        false,
		SecureCookie:      false,
		SendAuthorization: false,
		TokenLookup:       "cookie:jwt, header:Authorization",
	}
	if err := middleware.MiddlewareInit(); err != nil {
		panic(err)
	}
	return middleware
}

func payloadFunc(data interface{}) jwt.MapClaims {
	if result, ok := data.(AccountMatchResult); ok {
		return jwt.MapClaims{
			identityKey: result.Id,
			roleKey:     result.Role,
		}
	}
	return jwt.MapClaims{}
}

func identityHandler(claims jwtGo.MapClaims) interface{} {
	return claims[identityKey]
}

func RequireRoleMiddleware(roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		role, ok := claims[roleKey]
		if !ok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatus(http.StatusForbidden)
	}
}
