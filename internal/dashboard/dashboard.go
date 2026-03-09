package dashboard

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/webdunesurfer/SearchInlet/internal/auth"
	"github.com/webdunesurfer/SearchInlet/internal/db"
	"gorm.io/gorm"
)

//go:embed templates/*.html
var dashboardFS embed.FS

type Dashboard struct {
	db           *gorm.DB
	tm           *auth.TokenManager
	ll           *auth.LoginLimiter
	templateDir  string
	adminUser    string
	adminPass    string
	sessionToken string
}

type DashboardData struct {
	Tokens       []db.Token
	UsageStats   []UsageStat
	ActiveTokens int
	TotalTokens  int
	ErrorMessage string
	SuccessToken string
}

type UsageStat struct {
	TokenName string
	TotalUses int64
	LastUsed  time.Time
}

func NewDashboard(db *gorm.DB, tm *auth.TokenManager, adminUser, adminPass string) *Dashboard {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	sessionToken := hex.EncodeToString(bytes)

	return &Dashboard{
		db:           db,
		tm:           tm,
		ll:           auth.NewLoginLimiter(db),
		templateDir:  "./templates",
		adminUser:    adminUser,
		adminPass:    adminPass,
		sessionToken: sessionToken,
	}
}

func (d *Dashboard) HandleHome(w http.ResponseWriter, r *http.Request) {
	if !d.authenticate(w, r) {
		d.loginForm(w, "")
		return
	}

	data := DashboardData{}
	
	// Check for recently created token cookie
	if cookie, err := r.Cookie("token_created"); err == nil {
		data.SuccessToken = cookie.Value
		// Expire the cookie immediately
		http.SetCookie(w, &http.Cookie{
			Name:   "token_created",
			Value:  "",
			MaxAge: -1,
			Path:   "/",
		})
	}

	tokens, _ := d.tm.GetAllTokens()
	data.Tokens = tokens
	data.TotalTokens = len(tokens)
	data.ActiveTokens = 0

	for _, t := range tokens {
		if t.Active {
			data.ActiveTokens++
		}
	}

	stats, _ := d.getUsageStats()
	data.UsageStats = stats

	tmpl := template.New("dashboard")
	tmpl, err := template.ParseFS(dashboardFS, "templates/dashboard.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func (d *Dashboard) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Extract IP (handling potential proxy headers if needed later)
	ip := strings.Split(r.RemoteAddr, ":")[0]

	if d.ll.IsBanned(ip) {
		http.Error(w, "Too many failed attempts. Please try again later.", http.StatusTooManyRequests)
		return
	}

	if r.Method == "POST" {
		user := r.FormValue("username")
		pass := r.FormValue("password")

		if user == d.adminUser && pass == d.adminPass {
			d.ll.LogAttempt(ip, true)
			http.SetCookie(w, &http.Cookie{
				Name:     "admin_session",
				Value:    d.sessionToken,
				MaxAge:   86400,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		d.ll.LogAttempt(ip, false)
		d.loginForm(w, "Invalid credentials")
		return
	}

	d.loginForm(w, "")
}

func (d *Dashboard) loginForm(w http.ResponseWriter, errorMsg string) {
	tmplStr := `
<!DOCTYPE html>
<html>
<head>
	<title>SearchInlet - Login</title>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; background: #f0f2f5; margin: 0; }
		.login-box { background: white; padding: 40px; border-radius: 12px; box-shadow: 0 8px 24px rgba(0,0,0,0.1); width: 100%; max-width: 400px; }
		input { width: 100%; padding: 12px; margin: 10px 0; border: 1px solid #ddd; border-radius: 6px; box-sizing: border-box; font-size: 16px; }
		button { width: 100%; padding: 12px; background: #007bff; color: white; border: none; border-radius: 6px; cursor: pointer; font-size: 16px; font-weight: 600; transition: background 0.2s; }
		button:hover { background: #0056b3; }
		h2 { text-align: center; margin-bottom: 8px; color: #1c1e21; }
		p.subtitle { text-align: center; color: #606770; margin-bottom: 24px; font-size: 14px; }
		.error { color: #dc3545; background: #f8d7da; padding: 12px; border-radius: 6px; margin-bottom: 20px; text-align: center; font-size: 14px; border: 1px solid #f5c6cb; }
	</style>
</head>
<body>
	<div class="login-box">
		<h2>SearchInlet Admin</h2>
		<p class="subtitle">Please sign in to manage tokens</p>
		{{if .}}
		<div class="error">{{.}}</div>
		{{end}}
		<form action="/login" method="POST">
			<input type="text" name="username" placeholder="Username" required autocomplete="username">
			<input type="password" name="password" placeholder="Password" required autocomplete="current-password">
			<button type="submit">Sign In</button>
		</form>
	</div>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.New("login").Parse(tmplStr))
	tmpl.Execute(w, errorMsg)
}

func (d *Dashboard) authenticate(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("admin_session")
	if err != nil || cookie.Value != d.sessionToken {
		return false
	}
	return true
}

func (d *Dashboard) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	if !d.authenticate(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		name := r.FormValue("name")
		if name == "" {
			name = "Token " + time.Now().Format("2006-01-02 15:04:05")
		}

		token, err := d.tm.CreateToken(name)
		if err != nil {
			http.Error(w, "Failed to create token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:   "token_created",
			Value:  token.Value,
			MaxAge: 60,
			Path:   "/",
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	d.HandleHome(w, r)
}

func (d *Dashboard) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	if !d.authenticate(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Token ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}

	if err := d.tm.RevokeToken(uint(id)); err != nil {
		http.Error(w, "Failed to revoke token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (d *Dashboard) getUsageStats() ([]UsageStat, error) {
	var stats []UsageStat

	var tokens []db.Token
	if err := d.db.Find(&tokens).Error; err != nil {
		return nil, err
	}

	for _, token := range tokens {
		var count int64
		d.db.Model(&db.UsageLog{}).Where("token_id = ?", token.ID).Count(&count)

		var lastLog db.UsageLog
		lastUsed := time.Time{}
		if err := d.db.Model(&db.UsageLog{}).Where("token_id = ?", token.ID).Order("created_at DESC").First(&lastLog).Error; err == nil {
			lastUsed = lastLog.CreatedAt
		}

		stats = append(stats, UsageStat{
			TokenName: token.Name,
			TotalUses: count,
			LastUsed:  lastUsed,
		})
	}

	return stats, nil
}
