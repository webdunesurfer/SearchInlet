package dashboard

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/webdunesurfer/SearchInlet/internal/auth"
	"github.com/webdunesurfer/SearchInlet/internal/db"
	"github.com/webdunesurfer/SearchInlet/internal/distiller"
	"gorm.io/gorm"
)

//go:embed templates/*.html
var dashboardFS embed.FS

type DownloadStatus struct {
	Model      string  `json:"model"`
	Status     string  `json:"status"`
	Percentage float64 `json:"percentage"`
	Active     bool    `json:"active"`
}

type Dashboard struct {
	db               *gorm.DB
	tm               *auth.TokenManager
	ll               *auth.LoginLimiter
	templateDir      string
	adminUser        string
	adminPass        string
	sessionToken     string
	distiller        *distiller.OllamaClient
	downloadProgress map[string]DownloadStatus
	progressMu       sync.RWMutex
}

type DashboardData struct {
	Tokens              []db.Token
	UsageStats          []UsageStat
	ActiveTokens        int
	TotalTokens         int
	ErrorMessage        string
	SuccessToken        string
	HostURL             string
	DistillationEnabled bool
	DistillationModel   string
	DistillationPrompt  string
	DownloadedModels    []string
}

type UsageStat struct {
	TokenName        string
	TotalUses        int64
	LastUsed         time.Time
	AvgSearchMS      int64
	AvgDistillMS     int64
	AvgCompression   float64
	DistillationUsed int64
}

func NewDashboard(db *gorm.DB, tm *auth.TokenManager, adminUser, adminPass, ollamaURL string) *Dashboard {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	sessionToken := hex.EncodeToString(bytes)

	log.Printf("Initializing Dashboard with Admin User: %s (Password length: %d)", adminUser, len(adminPass))

	return &Dashboard{
		db:               db,
		tm:               tm,
		ll:               auth.NewLoginLimiter(db),
		templateDir:      "./templates",
		adminUser:        adminUser,
		adminPass:        adminPass,
		sessionToken:     sessionToken,
		distiller:        distiller.NewOllamaClient(ollamaURL),
		downloadProgress: make(map[string]DownloadStatus),
	}
}

func (d *Dashboard) getSetting(key, defaultValue string) string {
	var setting db.GlobalSetting
	if err := d.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return defaultValue
	}
	return setting.Value
}

func (d *Dashboard) saveSetting(key, value string) error {
	var setting db.GlobalSetting
	if err := d.db.Where("key = ?", key).First(&setting).Error; err != nil {
		setting = db.GlobalSetting{Key: key, Value: value}
		return d.db.Create(&setting).Error
	}
	setting.Value = value
	return d.db.Save(&setting).Error
}

func (d *Dashboard) HandleHome(w http.ResponseWriter, r *http.Request) {
	if !d.authenticate(w, r) {
		d.loginForm(w, "")
		return
	}

	data := DashboardData{}
	
	// Determine the base URL for the instructions
	scheme := "https"
	if r.TLS == nil && !strings.Contains(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "http"
	}
	data.HostURL = scheme + "://" + r.Host

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

	// Distillation Settings
	data.DistillationEnabled = d.getSetting("distillation_enabled", "false") == "true"
	data.DistillationModel = d.getSetting("distillation_model", "qwen2.5:1.5b")
	data.DistillationPrompt = d.getSetting("distillation_prompt", "Summarize and extract the most relevant information from the following search results. Be concise and maintain technical accuracy.")

	// Fetch downloaded models
	models, err := d.distiller.ListModels(r.Context())
	if err != nil {
		log.Printf("Failed to list models: %v", err)
	} else {
		data.DownloadedModels = models
	}

	tmpl := template.New("dashboard")
	tmpl, err = template.ParseFS(dashboardFS, "templates/dashboard.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func (d *Dashboard) HandleSaveSettings(w http.ResponseWriter, r *http.Request) {
	if !d.authenticate(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	distillationEnabled := r.FormValue("distillation_enabled") == "on"
	distillationModel := r.FormValue("distillation_model")
	customModel := r.FormValue("custom_model")

	if distillationModel == "custom" && customModel != "" {
		distillationModel = customModel
	}

	distillationPrompt := r.FormValue("distillation_prompt")

	oldModel := d.getSetting("distillation_model", "")

	d.saveSetting("distillation_enabled", strconv.FormatBool(distillationEnabled))
	d.saveSetting("distillation_model", distillationModel)
	d.saveSetting("distillation_prompt", distillationPrompt)

	// Trigger pull if model changed or if distillation was just enabled
	if distillationEnabled && distillationModel != oldModel {
		log.Printf("Triggering background pull for model: %s", distillationModel)
		go func(modelName string) {
			d.progressMu.Lock()
			d.downloadProgress[modelName] = DownloadStatus{
				Model:      modelName,
				Status:     "Initializing...",
				Percentage: 0,
				Active:     true,
			}
			d.progressMu.Unlock()

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
			defer cancel()

			err := d.distiller.PullModel(ctx, modelName, func(p distiller.PullProgress) {
				d.progressMu.Lock()
				status := d.downloadProgress[modelName]
				status.Status = p.Status
				if p.Total > 0 {
					status.Percentage = float64(p.Completed) / float64(p.Total) * 100
				} else if p.Status == "success" {
					status.Percentage = 100
				}
				d.downloadProgress[modelName] = status
				d.progressMu.Unlock()
			})

			d.progressMu.Lock()
			status := d.downloadProgress[modelName]
			if err != nil {
				log.Printf("ERROR: Failed to pull model %s: %v", modelName, err)
				status.Status = "Error: " + err.Error()
				status.Active = false
			} else {
				log.Printf("SUCCESS: Model %s pulled successfully", modelName)
				status.Active = false
				status.Percentage = 100
				status.Status = "Completed"
			}
			d.downloadProgress[modelName] = status
			d.progressMu.Unlock()

			// Always keep status visible for a while so user sees it finished
			time.Sleep(30 * time.Second)
			d.progressMu.Lock()
			delete(d.downloadProgress, modelName)
			d.progressMu.Unlock()
		}(distillationModel)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (d *Dashboard) HandleDownloadStatus(w http.ResponseWriter, r *http.Request) {
	if !d.authenticate(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	d.progressMu.RLock()
	defer d.progressMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d.downloadProgress)
}

func (d *Dashboard) HandleDeleteModel(w http.ResponseWriter, r *http.Request) {
	if !d.authenticate(w, r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	modelName := r.URL.Query().Get("model")
	if modelName == "" {
		http.Error(w, "Model name required", http.StatusBadRequest)
		return
	}

	log.Printf("Deleting model: %s", modelName)
	if err := d.distiller.DeleteModel(r.Context(), modelName); err != nil {
		log.Printf("Failed to delete model %s: %v", modelName, err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (d *Dashboard) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Extract IP (handling potential proxy headers if needed later)
	ip := strings.Split(r.RemoteAddr, ":")[0]

	if d.ll.IsBanned(ip) {
		log.Printf("Login blocked: IP %s is banned", ip)
		http.Error(w, "Too many failed attempts. Please try again later.", http.StatusTooManyRequests)
		return
	}

	if r.Method == "POST" {
		user := r.FormValue("username")
		pass := r.FormValue("password")

		log.Printf("Login attempt from %s - User: %s | Expected: %s", ip, user, d.adminUser)

		if user == d.adminUser && pass == d.adminPass {
			log.Printf("Login success for user: %s", user)
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

		log.Printf("Login failed for user: %s", user)
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

		plaintext, _, err := d.tm.CreateToken(name)
		if err != nil {
			http.Error(w, "Failed to create token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:   "token_created",
			Value:  plaintext,
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

		var metrics struct {
			AvgSearch    float64
			AvgDistill   float64
			TotalInput   int64
			TotalOutput  int64
			DistillCount int64
		}

		d.db.Model(&db.UsageLog{}).
			Select("AVG(search_latency_ms) as avg_search, AVG(distill_latency_ms) as avg_distill, SUM(input_tokens) as total_input, SUM(output_tokens) as total_output, SUM(CASE WHEN distillation_enabled THEN 1 ELSE 0 END) as distill_count").
			Where("token_id = ?", token.ID).
			Scan(&metrics)

		compression := 0.0
		if metrics.TotalInput > 0 {
			compression = (1.0 - float64(metrics.TotalOutput)/float64(metrics.TotalInput)) * 100
		}

		stats = append(stats, UsageStat{
			TokenName:        token.Name,
			TotalUses:        count,
			LastUsed:         lastUsed,
			AvgSearchMS:      int64(metrics.AvgSearch),
			AvgDistillMS:     int64(metrics.AvgDistill),
			AvgCompression:   compression,
			DistillationUsed: metrics.DistillCount,
		})
	}

	return stats, nil
}
