package utils

import (
	"net/http"
)

func SetCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   "localhost",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   false,
	})
}
