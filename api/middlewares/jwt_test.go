package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockAccountMatcher struct{}

func (mam mockAccountMatcher) AccountMatch(name, password string) (bool, AccountMatchResult) {
	if name == "admin" && password == "123" {
		return true, AccountMatchResult{"1", "admin"}
	}

	if name == "operator" && password == "1234" {
		return true, AccountMatchResult{"2", "operator"}
	}

	return false, AccountMatchResult{}
}


func TestNewJwtMiddleware_LoginHandler(t *testing.T) {
	middleware := NewJwtMiddleware(mockAccountMatcher{})
	router := gin.Default()

	router.POST("/login", middleware.LoginHandler)

	loginJson := `{"name":"admin","password":"123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginJson))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	loginJson = `{"name":"admin","password":"1234"}`
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginJson))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}


func TestNewJwtMiddleware_MiddlewareFunc(t *testing.T) {
	middleware := NewJwtMiddleware(mockAccountMatcher{})
	router := gin.Default()

	router.POST("/login", middleware.LoginHandler)
	router.GET("/hello", middleware.MiddlewareFunc(), func(c *gin.Context) {
		data, ok := c.Get("userID")
		require.True(t, ok)
		id, ok := data.(string)
		require.True(t, ok)
		require.Equal(t, "1", id)
		c.Writer.WriteString("hello")
	})

	// login to get token
	loginJson := `{"name":"admin","password":"123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginJson))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var result struct {
		Code  int    `json:"code"`
		Token string `json:"token"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	require.Equal(t, 200, result.Code)

	// hello endpoint request
	req = httptest.NewRequest(http.MethodGet, "/hello", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Add("authorization", "Bearer " + result.Token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "hello", w.Body.String())
}

func TestRequireRoleMiddleware(t *testing.T) {
	middleware := NewJwtMiddleware(mockAccountMatcher{})
	router := gin.Default()

	router.POST("/login", middleware.LoginHandler)
	router.GET("/hello", middleware.MiddlewareFunc(), RequireRoleMiddleware([]string{"admin"}),func(c *gin.Context) {
		c.Writer.WriteString("hello")
	})

	getToken := func(name, password string) string {
		loginJson := fmt.Sprintf(`{"name":"%s","password":"%s"}`, name, password)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginJson))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var result struct {
			Code  int    `json:"code"`
			Token string `json:"token"`
		}
		json.Unmarshal(w.Body.Bytes(), &result)
		return result.Token
	}

	adminToken := getToken("admin", "123")
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("authorization", "Bearer " + adminToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "hello", w.Body.String())

	operatorToken := getToken("operator", "1234")
	req = httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("authorization", "Bearer " + operatorToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}