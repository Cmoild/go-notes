package api

import (
	"encoding/json"
	"gonotes/internal/db"
	"gonotes/internal/http"
	"regexp"
	"strconv"
)

type UserStruct struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func User(w http.ResponseWriter, r *http.Request) {
	db := db.Connect()
	defer db.Close()

	switch r.Method {
	case "POST":
		var usr UserStruct
		err := json.Unmarshal(r.Body, &usr)
		if err != nil {
			http.SendBadRequest400(w, r)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.SendInternalServerError500(w, r)
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec(
			"INSERT INTO Users (username, email, user_password) VALUES ($1, $2, $3)",
			usr.Username,
			usr.Email,
			usr.Password)
		if err != nil {
			http.SendConflict409(w, r)
			return
		}

		var id int
		row := tx.QueryRow(
			"SELECT id FROM Users WHERE username = $1 AND email = $2",
			usr.Username,
			usr.Email,
		)
		_ = row.Scan(&id)

		tx.Commit()

		resp := map[string]any{
			"id": id,
		}
		w.SetStatus("201 Created")
		w.SetHeader("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	case "GET":
		http.SendOK200(w, r)
	}

}

func UserDetail(w http.ResponseWriter, r *http.Request) {
	db := db.Connect()
	defer db.Close()

	switch r.Method {
	case "GET":
		var id int
		var usr UserStruct

		re := regexp.MustCompile(`^/api/users/(\d+)$`)
		id, _ = strconv.Atoi(re.FindStringSubmatch(r.Path)[1])
		row := db.QueryRow(
			"SELECT username, email FROM Users WHERE id = $1",
			id,
		)

		err := row.Scan(&usr.Username, &usr.Email)
		if err != nil {
			http.SendNotFound404(w, r)
			return
		}

		resp := map[string]any{
			"username": usr.Username,
			"email":    usr.Email,
		}
		w.SetStatus("200 OK")
		w.SetHeader("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
