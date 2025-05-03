package main

import (
	api "gonotes/internal/api"
	dbase "gonotes/internal/db"
	http "gonotes/internal/http"
)

func main() {
	http.HandleFunc(`^/api/users$`, api.User)
	http.HandleFunc(`^/api/users/([0-9]*)$`, api.UserDetail)
	http.HandleFunc(`^/api/auth/login`, api.Login)
	http.HandleFunc(`^/api/notes$`, api.Notes)
	http.HandleFunc(`^/api/notes/([0-9]*)$`, api.NotesDetail)

	db := dbase.Connect()
	db.Close()

	http.ListenAndServe(":8080")
}
