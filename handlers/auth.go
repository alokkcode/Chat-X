package handlers

import (
	"CHATX/models"
	"html/template"
	"net/http"
)

// Combined handler for /register
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/register.html"))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		role := r.FormValue("role")

		err := models.RegisterUser(username, password, role)
		if err != nil {
			http.Error(w, "Registration failed", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// Combined handler for /login
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/login.html"))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := models.GetUserByUsername(username)
		if err != nil || !models.CheckPasswordHash(password, user.Password) {
			http.Error(w, "Invalid login", http.StatusUnauthorized)
			return
		}

		// Set session cookies
		http.SetCookie(w, &http.Cookie{
			Name:  "username",
			Value: user.Username,
			Path:  "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "role",
			Value: user.Role,
			Path:  "/",
		})

		http.Redirect(w, r, "/hub", http.StatusSeeOther)
	}
}
