package handlers

import (
	"html/template"
	"CHATX/models"
	"net/http"
)

// RegisterUser handles user registration
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("templates/register.html"))
		tmpl.Execute(w, nil)
		return
	}

	// Handle form POST
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	role := r.FormValue("role") // get role from form


	if username == "" || email == "" || password == "" {
		http.Error(w, "All fields required", http.StatusBadRequest)
		return
	}

	// Call models.RegisterUser to handle DB insertion
	err := models.RegisterUser(username, email, password, role)
	if err != nil {
		http.Error(w, "Registration failed: "+err.Error(), http.StatusInternalServerError)
		return
	}


	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Combined handler for /login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tmpl := template.Must(template.ParseFiles("templates/login.html"))
		tmpl.Execute(w, nil)

	case "POST":
		r.ParseForm()
		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := models.GetUserByEmail(email)
		if err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		if !models.CheckPasswordHash(password, user.PasswordHash) {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Generate and store session token
		sessionToken, err := models.GenerateSessionToken()
		if err != nil {
			http.Error(w, "Error generating session", http.StatusInternalServerError)
			return
		}

		err = models.StoreSessionToken(user.ID, sessionToken)
		if err != nil {
			http.Error(w, "Server error storing session", http.StatusInternalServerError)
			return
		}

		// Set cookie with session token
		cookie := http.Cookie{
			Name:     "session",
			Value:    sessionToken,
			Path:     "/",
			HttpOnly: true, // Prevents JavaScript access for security
			Secure:   true, // Requires HTTPS for better security
		}
		http.SetCookie(w, &cookie)

		http.Redirect(w, r, "/hub", http.StatusSeeOther)
	}
}


