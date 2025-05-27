package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// 常量定義
const (
	zohoEndpointAuth         = "https://accounts.zoho.com/oauth/v2/auth"
	zohoEndpointToken        = "https://accounts.zoho.com/oauth/v2/token"
	zohoEndpointKeys         = "https://accounts.zoho.com/oauth/v2/keys"
	zohoEndpointAccounts     = "https://mail.zoho.com/api/accounts"
	zohoEndpointRefreshToken = zohoEndpointToken
)

// Token 結構體
type ZohoToken struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Raw          map[string]interface{}
	ExpiredAt    time.Time `json:"expired_at"`
}

// 創建新 Token
func NewZohoToken(clientID, clientSecret string, data map[string]interface{}) *ZohoToken {
	expiresIn := int64(0)
	if v, ok := data["expires_in"].(float64); ok {
		expiresIn = int64(v)
	}
	return &ZohoToken{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Raw:          data,
		ExpiredAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
}

func (t *ZohoToken) GetAccessToken() (string, error) {
	if t.IsExpired() {
		if err := t.RefreshToken(); err != nil {
			return "", err
		}
	}
	if v, ok := t.Raw["access_token"].(string); ok {
		return v, nil
	}
	return "", errors.New("access_token not found")
}

func (t *ZohoToken) GetRefreshToken() string {
	if v, ok := t.Raw["refresh_token"].(string); ok {
		return v
	}
	return ""
}

func (t *ZohoToken) GetApiDomain() string {
	if v, ok := t.Raw["api_domain"].(string); ok {
		return v
	}
	return ""
}

func (t *ZohoToken) GetTokenType() string {
	if v, ok := t.Raw["token_type"].(string); ok {
		return v
	}
	return ""
}

func (t *ZohoToken) GetExpiresIn() int64 {
	if v, ok := t.Raw["expires_in"].(float64); ok {
		return int64(v)
	}
	return 0
}

func (t *ZohoToken) GetExpiredAt() time.Time {
	return t.ExpiredAt
}

func (t *ZohoToken) IsExpired() bool {
	return time.Now().After(t.ExpiredAt)
}

func (t *ZohoToken) GetAuthHeaders() map[string]string {
	accessToken, _ := t.GetAccessToken()
	return map[string]string{
		"Authorization": "Zoho-oauthtoken " + accessToken,
	}
}

// 刷新 token
func (t *ZohoToken) RefreshToken() error {
	data := url.Values{}
	data.Set("refresh_token", t.GetRefreshToken())
	data.Set("client_id", t.ClientID)
	data.Set("client_secret", t.ClientSecret)
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm(zohoEndpointRefreshToken, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	for k, v := range result {
		t.Raw[k] = v
	}
	t.ExpiredAt = time.Now().Add(time.Duration(t.GetExpiresIn()) * time.Second)
	return nil
}

// ZohoConfig 結構體，封裝所有 Zoho 相關設定
type ZohoConfig struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
	AllowDomains []string
	RedirectURI  string
	DBPath       string // Add DBPath field
}

// 從環境變數載入 ZohoConfig
func LoadZohoConfigFromEnv() ZohoConfig {
	var scopes []string
	if v := os.Getenv("ZOHO_SCOPES"); strings.TrimSpace(v) != "" {
		scopes = filterNonEmpty(strings.Split(v, ","))
	} else {
		scopes = []string{}
	}
	var allowDomains []string
	if v := os.Getenv("ZOHO_ALLOW_DOMAINS"); strings.TrimSpace(v) != "" {
		allowDomains = filterNonEmpty(strings.Split(v, ","))
	} else {
		allowDomains = []string{}
	}
	return ZohoConfig{
		ClientID:     os.Getenv("ZOHO_CLIENT_ID"),
		ClientSecret: os.Getenv("ZOHO_CLIENT_SECRET"),
		Scopes:       scopes,
		AllowDomains: allowDomains,
		RedirectURI:  os.Getenv("ZOHO_REDIRECT_URI"),
		DBPath:       os.Getenv("ZOHO_DB_PATH"),
	}
}

// 工具: 過濾空字串
func filterNonEmpty(arr []string) []string {
	var out []string
	for _, v := range arr {
		if s := strings.TrimSpace(v); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// OAuth 主體
type ZohoOAuth struct {
	Config ZohoConfig
}

// 創建 OAuth 實例
func NewZohoOAuth(config ZohoConfig) *ZohoOAuth {
	return &ZohoOAuth{
		Config: config,
	}
}

// 取得授權 URL
func (z *ZohoOAuth) GetAuthRedirectURL() string {
	params := url.Values{}
	params.Set("client_id", z.Config.ClientID)
	params.Set("response_type", "code")
	scopes := z.Config.Scopes
	if len(scopes) == 0 {
		scopes = []string{"ZohoMail.accounts.READ"}
	}
	params.Set("scope", strings.Join(scopes, ","))
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")
	params.Set("redirect_uri", z.Config.RedirectURI)
	return fmt.Sprintf("%s?%s", zohoEndpointAuth, params.Encode())
}

// 使用請求信息獲取授權URL
func (z *ZohoOAuth) GetAuthRedirectURLCustom(redirect_uri string) string {
	z.Config.RedirectURI = redirect_uri
	return z.GetAuthRedirectURL()
}

// 用 code 換取 access token
func (z *ZohoOAuth) GetAccessToken(code string) (*ZohoToken, error) {
	data := url.Values{}
	data.Set("client_id", z.Config.ClientID)
	data.Set("grant_type", "authorization_code")
	data.Set("client_secret", z.Config.ClientSecret)
	data.Set("redirect_uri", z.Config.RedirectURI)
	data.Set("code", code)

	resp, err := http.PostForm(zohoEndpointToken, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return NewZohoToken(z.Config.ClientID, z.Config.ClientSecret, result), nil
}

// 使用請求信息獲取訪問令牌
func (z *ZohoOAuth) GetAccessTokenCustom(code string, redirect_url string) (*ZohoToken, error) {
	z.Config.RedirectURI = redirect_url
	return z.GetAccessToken(code)
}

// 用 code 取得 email
func (z *ZohoOAuth) GetEmail(code string) (string, error) {
	token, err := z.GetAccessToken(code)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("GET", zohoEndpointAccounts, nil)
	if err != nil {
		return "", err
	}
	for k, v := range token.GetAuthHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	// 解析 data.0.primaryEmailAddress
	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		return "", errors.New("no account data")
	}
	account, ok := data[0].(map[string]interface{})
	if !ok {
		return "", errors.New("invalid account data")
	}
	email, ok := account["primaryEmailAddress"].(string)
	if !ok {
		return "", errors.New("primaryEmailAddress not found")
	}
	return email, nil
}

// Helper method to extract email from token
func (z *ZohoOAuth) getEmailFromToken(token *ZohoToken) (string, error) {
	req, err := http.NewRequest("GET", zohoEndpointAccounts, nil)
	if err != nil {
		return "", err
	}
	for k, v := range token.GetAuthHeaders() {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		return "", errors.New("no account data")
	}
	account, ok := data[0].(map[string]interface{})
	if !ok {
		return "", errors.New("invalid account data")
	}
	email, ok := account["primaryEmailAddress"].(string)
	if !ok {
		return "", errors.New("primaryEmailAddress not found")
	}
	return email, nil
}

// 使用請求信息獲取Email
func (z *ZohoOAuth) GetEmailCustom(code string, redirect_url string) (string, error) {
	token, err := z.GetAccessTokenCustom(code, redirect_url)
	if err != nil {
		return "", err
	}
	return z.getEmailFromToken(token)
}

// 提取 email domain
func extractDomain(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// 檢查 email 的 domain 是否允許
func (z *ZohoOAuth) IsEmailDomainAllowed(email string) bool {
	domain := extractDomain(email)
	for _, allow := range z.Config.AllowDomains {
		if strings.EqualFold(strings.TrimSpace(allow), domain) {
			return true
		}
	}
	return false
}
