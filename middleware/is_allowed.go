package middleware

import (
	"os"
	"strings"
	"log"
	"net/http"
	"omiro/helper"
	"omiro/redis"
	"time"
)

func AllowHandshake(ip string, limit int) (bool, error) {
	allowed, err := redis.CheckRateLimit(ip, limit, 1*time.Minute)
	if err != nil {
		return false, err
	}
	return allowed, nil
}

func EnsureUpgradeChecks(w http.ResponseWriter, r *http.Request) bool {

	origin := r.Header.Get("Origin")
	// TODO: Add allowed origins
	if !originAllowed(origin) {
		http.Error(w, "Forbidden origin", http.StatusForbidden)
		return false
	}

	ip := helper.GetRealIP(r)
	log.Printf("IP: %s", ip)
	ok, err := AllowHandshake(ip, 60)
	if err != nil || !ok {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return false
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("X-Session-Token")
	}
	log.Printf("Token: %s", token)
	if !VerifySessionToken(token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	log.Printf("Token verified")
	return true
}

func originAllowed(origin string) bool {
	if origin == "http://localhost:8080" || origin == "https://omiro.underthedesk.blog" {
		return true
	}
	for _, o := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" && o == origin {
			return true
		}
	}
	return false
}
