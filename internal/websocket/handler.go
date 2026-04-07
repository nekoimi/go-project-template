package websocket

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	wslib "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = wslib.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: restrict in production
	},
}

type WSHandler struct {
	manager   *Manager
	jwtSecret string
	logger    *zap.Logger
}

func NewWSHandler(manager *Manager, jwtSecret string, logger *zap.Logger) *WSHandler {
	return &WSHandler{
		manager:   manager,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// Upgrade handles WebSocket upgrade requests.
// GET /ws/v1/chat?token=<jwt>
func (h *WSHandler) Upgrade(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	userID, err := h.validateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	client := newClient(h.manager, conn, userID, h.logger)
	h.manager.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (h *WSHandler) validateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrTokenInvalidClaims
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return "", jwt.ErrTokenExpired
		}
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", jwt.ErrTokenInvalidClaims
	}

	return userID, nil
}
