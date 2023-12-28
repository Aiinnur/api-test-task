package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

// Node струкртура, котоая содержит все необходимые поля для заметок
type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// db глобальная переменная,для взаимодействия с базой данных.
var db *sql.DB

// initialization выполняет инициализацию базы данных SQLite и создает таблицу 'notes', если она отсутствует.
// Подключается к базе данных SQLite, указанной в URL-строке "./notes.db".
// В случае ошибки при подключении или создании таблицы, программа завершает работу с фатальной ошибкой.
func initialization() {
	db, err := sql.Open("sqlite3", "./notes.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY AUTOINCREMENT, " +
		"title TEXT, content TEXT, created_at DATETIME);")
	if err != nil {
		log.Fatal(err)
	}
}

// createNote обрабатывает запрос на создание новой записи в базе данных.
// Извлекаются данные о новой заметке из тела запроса, затем устанавливается текущая дата и время создания.
// Далее выполняется запрос INSERT для добавления новой записи в таблицу 'notes'.
// В случае успешного добавления, возвращается HTTP-ответ со статусом 201 Created
// и содержащий данные о созданной заметке в формате JSON.
// Если при выполнении запроса INSERT произошла ошибка, возвращается HTTP-ответ со статусом 500 Internal Server Error,
// содержащий информацию об ошибке.
// Если произошла ошибка при декодировании тела запроса, возвращается HTTP-ответ со статусом 400 Bad Request
// и информацией об ошибке декодирования.
func createNote(w http.ResponseWriter, r *http.Request) {
	var note Note
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	note.CreatedAt = time.Now()

	result, err := db.Exec("INSERT INTO notes (title, content, created_at) VALUES (?, ?, ?)", note.Title, note.Content, note.CreatedAt)
	if err != nil {
		log.Printf("Error when inserting a new note into the database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	note.ID = int(id)

	resp, err := json.Marshal(note)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

// getNote обрабатывает запрос на получение одной записи из базы данных по указанному ID.
// Извлекается ID записи из URL запроса, и выполняется запрос SELECT для получения данных о заметке.
// Если запись с указанным ID не существует, возвращается HTTP-ответ со статусом 404 Not Found.
// Если при выполнении запроса SELECT произошла ошибка, возвращается HTTP-ответ со статусом 500 Internal Server Error,
// содержащий информацию об ошибке.
// В случае успешного выполнения запроса, возвращается HTTP-ответ со статусом 200 OK
// и содержащий данные о заметке в формате JSON.
func getNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var note Note
	err := db.QueryRow("SELECT id, title, content, created_at FROM notes WHERE id=?", id).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		log.Printf("Error when retrieving a note by ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(note)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// getNotes обрабатывает запрос на получение всех записей из базы данных.
// Извлекаются все записи из таблицы 'notes', и их данные преобразуются в формат JSON.
// В случае успешного выполнения запроса, возвращается HTTP-ответ со статусом 200 OK
// и содержащий список всех заметок в формате JSON.
// Если при выполнении запроса произошла ошибка, возвращается HTTP-ответ со статусом 500 Internal Server Error,
// содержащий информацию об ошибке.
func getNotes(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, content, created_at FROM notes")
	if err != nil {
		log.Printf("Error receiving notes: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)
		if err != nil {
			log.Printf("Error reading notes: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		notes = append(notes, note)
	}

	resp, err := json.Marshal(notes)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// updateNote обрабатывает запрос на обновление существующей записи в базе данных по указанному ID.
// Извлекается ID записи из URL запроса, а затем декодируется тело запроса, содержащее обновленные данные для записи.
// Если декодирование тела запроса завершается ошибкой, возвращается HTTP-ответ со статусом 400 Bad Request
// и информацией об ошибке декодирования.
// Затем выполняется запрос UPDATE, обновляя заголовок и содержимое записи в таблице 'notes' на основе переданных данных.
// В случае успешного обновления, возвращается HTTP-ответ со статусом 200 OK.
// Если при выполнении запроса UPDATE произошла ошибка, возвращается HTTP-ответ со статусом 500 Internal Server Error,
// содержащий информацию об ошибке.
func updateNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var note Note
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE notes SET title=?, content=? WHERE id=?", note.Title, note.Content, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// deleteNote обрабатывает запрос на удаление записи из базы данных по указанному ID.
// Если запись с указанным ID существует, она удаляется из таблицы 'notes'.
// В случае успешного удаления, возвращается HTTP-ответ со статусом 200 OK.
// В случае ошибки при выполнении запроса DELETE или если запись с указанным ID не существует,
// возвращается HTTP-ответ со статусом 500 Internal Server Error, содержащий информацию об ошибке.
func deleteNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := db.Exec("DELETE FROM notes WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func main() {
	initialization()
	defer db.Close()

	r := chi.NewRouter()

	r.Post("/note", createNote)
	r.Get("/note/{id}", getNote)
	r.Get("/notes", getNotes)
	r.Patch("/note/{id}", updateNote)
	r.Delete("/note/{id}", deleteNote)

	log.Fatal(http.ListenAndServe(":8080", r))
}
