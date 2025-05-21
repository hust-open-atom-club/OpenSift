package admin

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/HUSTSecLab/criticality_score/cmd/apiserver/internal/model"
	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var (
	// Temporary whitelist of allowed GitHub usernames
	allowedUsers = []string{"jxlpzqc"} // TODO: Replace with database
)

// SessionController handles session-related operations
// @Summary Get github client id
// @Description Get github client id
// @Tags admin
// @Produce json
// @Success 200 {object} model.GitHubClientIDResp
// @Router /admin/session/github/clientid [get]
func getClientID(ctx *gin.Context) {
	githubClientID, _ := config.GetWebGitHubOAuth()
	ctx.JSON(http.StatusOK, model.GitHubClientIDResp{
		ClientID: githubClientID,
	})
}

// githubCallback godoc
// @Summary GitHub OAuth callback
// @Description Handles the GitHub OAuth callback and returns JWT token if user is authorized
// @Tags admin
// @Produce json
// @Param code query string true "GitHub OAuth Code"
// @Success 200 {object} model.GitHubCallbackResp
// @Failure 401 {object} map[string]string
// @Router /admin/session/github/callback [get]
func githubCallback(ctx *gin.Context) {
	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No code provided"})
		return
	}

	// Exchange code for access token
	accessToken, err := getGithubAccessToken(code)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get access token"})
		return
	}

	// Get GitHub user info
	username, err := getGithubUsername(accessToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user info"})
		return
	}

	// Check if user is allowed
	if !isUserAllowed(username) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authorized"})
		return
	}

	// Generate JWT token
	token, err := generateJWT(username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, model.GitHubCallbackResp{
		Token: token,
	})
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func getGithubAccessToken(code string) (string, error) {
	githubClientID, githubClientSecret := config.GetWebGitHubOAuth()
	reqURL := "https://github.com/login/oauth/access_token"
	data := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
	}{
		ClientID:     githubClientID,
		ClientSecret: githubClientSecret,
		Code:         code,
	}

	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", reqURL, strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

func getGithubUsername(accessToken string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Login, nil
}

func isUserAllowed(username string) bool {
	for _, user := range allowedUsers {
		if user == username {
			return true
		}
	}
	return false
}

func generateJWT(username string) (string, error) {
	_, githubClientSecret := config.GetWebGitHubOAuth()
	jwtSecret := []byte(githubClientSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"policy":   []string{"gitfile"},
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtSecret)
}

// getUserInfo godoc
// @Summary Get user information
// @Description Returns the authenticated user's username and policy
// @Tags admin
// @Produce json
// @Success 200 {object} model.UserInfoResp
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/session/userinfo [get]
func getUserInfo(ctx *gin.Context) {
	username, policy, err := getUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.UserInfoResp{
		Username: username,
		Policy:   policy,
	})
}

func registSession(e gin.IRoutes, w gin.IRoutes) {
	e.GET("/session/github/callback", githubCallback)
	e.GET("/session/github/clientid", getClientID)
	w.GET("/session/userinfo", getUserInfo)
}
