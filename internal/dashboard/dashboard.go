package dashboard

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/webdunesurfer/SearchInlet/internal/auth"
	"github.com/webdunesurfer/SearchInlet/internal/db"
	"gorm.io/gorm"
)

//go:embed templates/*.html
var dashboardFS embed.FS

type Dashboard struct {
	db          *gorm.DB
	tm          *auth.TokenManager
	templateDir string
	adminUser   string
	adminPass   string
}

type DashboardData struct {
	Tokens       []db.Token
	UsageStats   []UsageStat
	ActiveTokens int
	TotalTokens  int
}

type UsageStat struct {
	TokenName string
	TotalUses int64
	LastUsed  time.Time
}

func NewDashboard(db *gorm.DB, tm *auth.TokenManager, adminUser, adminPass string) *Dashboard {
	return &Dashboard{
		db:          db,
		tm:          tm,
		templateDir: "./templates",
		adminUser:   adminUser,
		adminPass:   adminPass,
	}
}

func (d *Dashboard) HandleHome(w http.ResponseWriter, r *http.Request) {
	if !d.authenticate(w, r) {
		d.loginForm(w)
		return
	}

	data := DashboardData{}
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
	if r.Method == "POST" {
		user := r.FormValue("username")
		pass := r.FormValue("password")

		if user == d.adminUser && pass == d.adminPass {
			http.SetCookie(w, &http.Cookie{
				Name:   "admin_session",
				Value:  "authenticated",
				MaxAge: 86400,
				Path:   "/",
			})
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	d.loginForm(w)
}

func (d *Dashboard) loginForm(w http.ResponseWriter) {
	tmplStr := `
<!DOCTYPE html>
<html>
<head>
	<title>Login</title>
	<style>
		body { font-family: sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; background: #f5f5f5; }
		.login-box { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
		input { width: 100%; padding: 10px; margin: 10px 0; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
		button { width: 100%; padding: 10px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
		button:hover { background: #0056b3; }
		h2 { text-align: center; margin-bottom: 20px; }
	</style>
</head>
<body>
	<div class="login-box">
		<h2>Admin Login</h2>
		<form method="POST">
			<input type="text" name="username" placeholder="Username" required>
			<input type="password" name="password" placeholder="Password" required>
			<button type="submit">Login</button>
		</form>
	</div>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.New("login").Parse(tmplStr))
	tmpl.Execute(w, nil)
}

func (d *Dashboard) authenticate(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("admin_session")
	if err != nil || cookie.Value != "authenticated" {
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

		var lastUsed time.Time
		d.db.Model(&db.UsageLog{}).Where("token_id = ?", token.ID).Order("created_at DESC").First(&lastUsed)

		stats = append(stats, UsageStat{
			TokenName: token.Name,
			TotalUses: count,
			LastUsed:  lastUsed,
		})
	}

	return stats, nil
}

func (d *Dashboard) Run(addr string) error {
	if d.adminUser == "" {
		d.adminUser = "admin"
	}
	if d.adminPass == "" {
		d.adminPass = os.Getenv("ADMIN_PASSWORD")
		if d.adminPass == "" {
			d.adminPass = "admin123"
		}
	}

	if d.templateDir == "" {
		d.templateDir = "./templates"
	}

	http.HandleFunc("/", d.HandleHome)
	http.HandleFunc("/login", d.HandleLogin)
	http.HandleFunc("/create-token", d.HandleCreateToken)
	http.HandleFunc("/revoke-token", d.HandleRevokeToken)

	log.Printf("Starting admin dashboard on %s", addr)
	return http.ListenAndServe(addr, nil)
}
