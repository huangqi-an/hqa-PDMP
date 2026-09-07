package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func performAuthRequest(secret string, authorizationHeader string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Auth(secret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"userID": c.GetString("userID"),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authorizationHeader != "" {
		req.Header.Set("Authorization", authorizationHeader)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	return recorder
}

func signToken(secret string, subject string, tokenType string) string {
	claims := AccessClaims{
		Email: "test@example.com",
		Type:  tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))

	return signed
}

func TestAuthRejectsMissingAuthorizationHeader(t *testing.T) {
	recorder := performAuthRequest("test-secret", "")

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestAuthRejectsInvalidToken(t *testing.T) {
	recorder := performAuthRequest("test-secret", "Bearer invalid-token")

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestAuthAcceptsValidAccessToken(t *testing.T) {
	token := signToken("test-secret", "user-123", "access")

	recorder := performAuthRequest("test-secret", "Bearer "+token)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	body := recorder.Body.String()
	if body == "" {
		t.Fatal("response body should not be empty")
	}
}

func TestAuthRejectsRefreshToken(t *testing.T) {
	token := signToken("test-secret", "user-123", "refresh")

	recorder := performAuthRequest("test-secret", "Bearer "+token)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestAuthRejectsTokenSignedWithDifferentSecret(t *testing.T) {
	token := signToken("wrong-secret", "user-123", "access")

	recorder := performAuthRequest("test-secret", "Bearer "+token)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
