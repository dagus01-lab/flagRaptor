package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
)

// AUTHENTICATION HANDLERS
func verifyAuthentication(r *http.Request) (string, bool, bool) {
	authToken := r.Header.Get("X-Auth-Token")

	if authToken == cfg.ServerConf.APIToken {
		//log.Println("Authorization successful!")
		user, ok := getAuthenticatedUsername(r)
		if !ok {
			return "", true, false
		} else {
			return user, true, true
		}
	} else {
		//log.Println("Authorization not successful!")
		user, ok := getAuthenticatedUsername(r)
		if !ok {
			return "", false, false
		} else {
			return user, false, true
		}
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		if cfg.AuthConf.AuthEnable {
			rows, err := db.Query("SELECT * FROM users WHERE username=", username, " AND PASSWORD=", password)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			if !rows.Next() {
				http.Error(w, "Invalid password", http.StatusUnauthorized)
				return
			}

		} else {
			if cfg.AuthConf.WebPassword != password {
				http.Error(w, "Invalid password", http.StatusUnauthorized)
				return
			}
		}

		// Set the session cookie upon successful login
		setAuthenticatedUsername(r, w, username)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
		return
	}
}

func checkAuthHandler(w http.ResponseWriter, r *http.Request) {
	username, validAuth, validUser := verifyAuthentication(r)
	if validAuth && validUser {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"username": username})
	} else {
		http.Error(w, "Unauhorized", http.StatusUnauthorized)
	}

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	clearAuthenticatedUsername(r, w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout successful"})
}

//UTILITY FUNCTIONS FOR AUTHENTICATION

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func isAuthenticated(r *http.Request) bool {
	session, err := store.Get(r, cfg.ServerConf.SessionName)
	if err != nil {
		return false
	}

	userName, ok := session.Values["authenticatedUserName"].(string)
	return ok && userName != ""
}

func getAuthenticatedUsername(r *http.Request) (string, bool) {
	session, _ := store.Get(r, cfg.ServerConf.SessionName)
	if isAuthenticated(r) {
		userID, ok := session.Values["authenticatedUserName"].(string)
		return userID, ok
	} else {
		return "", false
	}
}

func setAuthenticatedUsername(r *http.Request, w http.ResponseWriter, username string) {
	session, err := store.Get(r, cfg.ServerConf.SessionName)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = cfg.ServerConf.SessionLifetime
	session.Values["authenticatedUserName"] = username
	session.Save(r, w)
}

func clearAuthenticatedUsername(r *http.Request, w http.ResponseWriter) {
	session, _ := store.Get(r, cfg.ServerConf.SessionName)
	session.Options.MaxAge = -1
	session.Save(r, w)
}

func initSessionManager() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   cfg.ServerConf.SessionLifetime,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
