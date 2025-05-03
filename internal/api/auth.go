package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"gonotes/internal/db"
	"gonotes/internal/http"
)

func generateToken() string {
	bytes := make([]byte, 20)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func Login(w http.ResponseWriter, r *http.Request) {
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

		row := db.QueryRow(
			"SELECT id FROM Users WHERE email = $1 AND user_password = $2",
			usr.Email,
			usr.Password,
		)
		var id int
		err = row.Scan(&id)
		if err != nil {
			http.SendNotFound404(w, r)
			return
		}

		token := generateToken()

		tx, err := db.Begin()
		if err != nil {
			http.SendInternalServerError500(w, r)
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec(
			"UPDATE Tokens SET revoked = true WHERE user_id = $1",
			id,
		)

		if err != nil {
			http.SendInternalServerError500(w, r)
			return
		}

		_, err = tx.Exec(
			"INSERT INTO Tokens (user_id, token, revoked) VALUES ($1, $2, $3)",
			id,
			token,
			0,
		)
		if err != nil {
			http.SendConflict409(w, r)
			return
		}

		tx.Commit()

		resp := map[string]any{
			"id":           id,
			"access_token": token,
		}
		w.SetStatus("200 OK")
		w.SetHeader("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
