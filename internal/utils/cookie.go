package utils

import (
	"net/http"
)

func SetCookie(w http.ResponseWriter, name, value string) {
	// TODO: add expiration
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

func DeleteCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   "localhost",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   false,
		MaxAge:   -1,
	})
}
