package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/utils"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseMiddleware(t *testing.T) {
	router := gin.Default()
	router.Use(ResponseMiddleware)
	router.GET("/hello", func(c *gin.Context) {
		c.Set("data", "hello")
	})
	router.GET("/error", func(c *gin.Context) {
		err := utils.NewHttpError(http.StatusUnauthorized, 101, "Login first")
		c.Set("error", err)
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
	require.Equal(t, `{"data":"hello"}`, w.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/error", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Equal(t, `{"code":101,"msg":"Login first"}`, w.Body.String())
}
