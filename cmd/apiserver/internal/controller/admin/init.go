package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func getUser(ctx *gin.Context) (username string, policy []string, err error) {
	usernameRaw, exists := ctx.Get("username")
	if !exists {
		return "", nil, fmt.Errorf("no login info")
	}

	userPolicyRaw, exists := ctx.Get("user_policy")
	if !exists {
		return "", nil, fmt.Errorf("user policy not found in context")
	}

	// username and userPolicy must be of type string
	username, ok := usernameRaw.(string)
	if !ok {
		return "", nil, fmt.Errorf("invalid username")
	}
	userPolicy, ok := userPolicyRaw.([]string)
	if !ok {
		return "", nil, fmt.Errorf("invalid user policy")
	}
	return username, userPolicy, nil
}

func jwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		token := c.GetHeader("Authorization")
		if token == "" {
			if c.Request.ParseForm() == nil {
				token = c.Request.Form.Get("auth_token")
			}

			if token == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "unauthorized",
				})
				return
			}
		}

		// Verify JWT token here
		_, githubClientSecret := config.GetWebGitHubOAuth()
		jwtSecret := []byte(githubClientSecret)
		claims := &struct {
			Username string   `json:"username"`
			Policy   []string `json:"policy"`
			jwt.StandardClaims
		}{}
		// remove "Bearer " prefix
		if !strings.HasPrefix(token, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "no Bearer token",
			})
			return
		}
		token = token[len("Bearer "):]
		tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err != nil {
			logger.Errorf("JWT parse error: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}
		if !tkn.Valid {
			logger.Warn("JWT token is invalid")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		// store user info in context
		c.Set("username", claims.Username)
		c.Set("user_policy", claims.Policy)

		c.Next()
	}
}

func Regist(e gin.IRouter) {
	g := e.Group("/admin")
	w := e.Group("/admin").Use(jwtMiddleware())

	registSession(g, w)
	registGitFile(w)
	registToolset(w)
	registWorkflow(w)
}
