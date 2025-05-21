package model

type GitHubClientIDResp struct {
	ClientID string `json:"clientId"`
}

type GitHubCallbackResp struct {
	Token string `json:"token"`
}

type UserInfoResp struct {
	Username string   `json:"username"`
	Policy   []string `json:"policy"`
}
