package api

import (
	"encoding/json"
	"gonotes/internal/db"
	"gonotes/internal/http"
	"log"
	"regexp"
	"strconv"
)

type Note struct {
	Content string `json:"content"`
	Id      int    `json:"id"`
}

func Notes(w http.ResponseWriter, r *http.Request) {
	db := db.Connect()
	defer db.Close()

	switch r.Method {
	case "POST":
		var note Note
		err := json.Unmarshal(r.Body, &note)
		if err != nil {
			http.SendBadRequest400(w, r)
			return
		}

		var apiToken string
		re := regexp.MustCompile(`^Authorization:\s*Bearer\s+([^\s]+)`)
		containsToken := false
		for _, header := range r.Headers {
			matches := re.FindStringSubmatch(header)
			if len(matches) > 1 {
				apiToken = matches[1]
				containsToken = true
			}
		}

		if !containsToken {
			http.SendForbidden403(w, r)
			return
		}

		var id int
		row := db.QueryRow(
			"SELECT user_id "+
				"FROM Tokens "+
				"WHERE token = $1 AND revoked = false",
			apiToken,
		)
		err = row.Scan(&id)

		if err != nil {
			http.SendForbidden403(w, r)
			return
		}

		_, err = db.Exec(
			"INSERT INTO Notes (note_text, user_id) VALUES ($1, $2)",
			note.Content,
			id,
		)

		if err != nil {
			http.SendInternalServerError500(w, r)
			return
		}

		w.SetStatus("201 Created")
		w.SetHeader("Content-Type", "text/plain")
		w.Write([]byte("Successfully created"))

	case "GET":
		var apiToken string
		re := regexp.MustCompile(`^Authorization:\s*Bearer\s+([^\s]+)`)
		containsToken := false
		for _, header := range r.Headers {
			matches := re.FindStringSubmatch(header)
			if len(matches) > 1 {
				apiToken = matches[1]
				containsToken = true
			}
		}

		if !containsToken {
			http.SendForbidden403(w, r)
			return
		}

		var id int
		row := db.QueryRow(
			"SELECT user_id "+
				"FROM Tokens "+
				"WHERE token = $1 AND revoked = false",
			apiToken,
		)
		err := row.Scan(&id)

		if err != nil {
			http.SendForbidden403(w, r)
			return
		}

		var notes []Note

		rows, err := db.Query(
			"SELECT id, note_text FROM Notes WHERE user_id = $1",
			id,
		)
		if err != nil {
			http.SendInternalServerError500(w, r)
			return
		}

		defer rows.Close()
		for rows.Next() {
			var note Note
			if err := rows.Scan(&note.Id, &note.Content); err != nil {
				http.SendInternalServerError500(w, r)
				return
			}
			notes = append(notes, note)
		}

		jsonData, err := json.Marshal(notes)
		if err != nil {
			http.SendInternalServerError500(w, r)
			return
		}

		w.SetStatus("200 OK")
		w.SetHeader("Content-Type", "application/json")
		w.Write(jsonData)
	}
}

func NotesDetail(w http.ResponseWriter, r *http.Request) {
	db := db.Connect()
	defer db.Close()

	switch r.Method {
	case "GET":
		var noteId int
		re := regexp.MustCompile(`^/api/notes/(\d+)$`)
		noteId, _ = strconv.Atoi(re.FindStringSubmatch(r.Path)[1])

		var apiToken string
		re = regexp.MustCompile(`^Authorization:\s*Bearer\s+([^\s]+)`)
		containsToken := false
		for _, header := range r.Headers {
			matches := re.FindStringSubmatch(header)
			if len(matches) > 1 {
				apiToken = matches[1]
				containsToken = true
			}
		}

		if !containsToken {
			http.SendForbidden403(w, r)
			return
		}

		var id int
		row := db.QueryRow(
			"SELECT user_id "+
				"FROM Tokens "+
				"WHERE token = $1 AND revoked = false",
			apiToken,
		)
		err := row.Scan(&id)

		if err != nil {
			http.SendForbidden403(w, r)
			return
		}

		var note Note
		log.Println(id)
		log.Println(noteId)

		row = db.QueryRow(
			"SELECT note_text, created_at FROM Notes WHERE user_id = $1 AND id = $2",
			id,
			noteId,
		)

		var cratedAt string
		err = row.Scan(&note.Content, &cratedAt)
		if err != nil {
			http.SendNotFound404(w, r)
			return
		}

		resp := map[string]any{
			"content":    note.Content,
			"created_at": cratedAt,
		}
		w.SetStatus("200 OK")
		w.SetHeader("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	case "DELETE":
		var noteId int
		re := regexp.MustCompile(`^/api/notes/(\d+)$`)
		noteId, _ = strconv.Atoi(re.FindStringSubmatch(r.Path)[1])

		var apiToken string
		re = regexp.MustCompile(`^Authorization:\s*Bearer\s+([^\s]+)`)
		containsToken := false
		for _, header := range r.Headers {
			matches := re.FindStringSubmatch(header)
			if len(matches) > 1 {
				apiToken = matches[1]
				containsToken = true
			}
		}

		if !containsToken {
			http.SendForbidden403(w, r)
			return
		}

		var id int
		row := db.QueryRow(
			"SELECT user_id "+
				"FROM Tokens "+
				"WHERE token = $1 AND revoked = false",
			apiToken,
		)
		err := row.Scan(&id)

		if err != nil {
			http.SendForbidden403(w, r)
			return
		}

		result, err := db.Exec(
			"DELETE FROM Notes WHERE user_id = $1 AND id = $2",
			id,
			noteId,
		)
		if err != nil {
			http.SendNotFound404(w, r)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			http.SendNotFound404(w, r)
			return
		}

		w.SetStatus("204 No Content")
		w.SetHeader("Content-Type", "text/plain")
		w.Write([]byte("Deleted"))

	case "PATCH":
		var note Note
		err := json.Unmarshal(r.Body, &note)
		if err != nil {
			http.SendBadRequest400(w, r)
			return
		}

		var noteId int
		re := regexp.MustCompile(`^/api/notes/(\d+)$`)
		noteId, _ = strconv.Atoi(re.FindStringSubmatch(r.Path)[1])

		var apiToken string
		re = regexp.MustCompile(`^Authorization:\s*Bearer\s+([^\s]+)`)
		containsToken := false
		for _, header := range r.Headers {
			matches := re.FindStringSubmatch(header)
			if len(matches) > 1 {
				apiToken = matches[1]
				containsToken = true
			}
		}

		if !containsToken {
			http.SendForbidden403(w, r)
			return
		}

		var id int
		row := db.QueryRow(
			"SELECT user_id "+
				"FROM Tokens "+
				"WHERE token = $1 AND revoked = false",
			apiToken,
		)
		err = row.Scan(&id)

		if err != nil {
			http.SendForbidden403(w, r)
			return
		}

		result, err := db.Exec(
			"UPDATE Notes SET note_text = $3 WHERE user_id = $1 AND id = $2",
			id,
			noteId,
			note.Content,
		)
		if err != nil {
			http.SendNotFound404(w, r)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			http.SendNotFound404(w, r)
			return
		}

		w.SetStatus("204 No Content")
		w.SetHeader("Content-Type", "text/plain")
		w.Write([]byte("Updated"))
	}
}
