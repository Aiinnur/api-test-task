package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var note = Note{
	Title:   "Тестовое задание",
	Content: "Написать api для тестового задания в проект",
}

var updatedNote = Note{
	Title:   "Урок по алгоритмам",
	Content: "Посмотреть вебинар по деревам отрезков",
}

func createNoteHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	createNote(w, r)
}

func getNoteHandler(w http.ResponseWriter, r *http.Request) {
	getNote(w, r)
}

func getNotesHandler(w http.ResponseWriter, r *http.Request) {
	getNotes(w, r)
}

func updateNoteHandler(w http.ResponseWriter, r *http.Request) {
	updateNote(w, r)
}

func deleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	deleteNote(w, r)
}

// TestCreateNote проверяет функцию создания новой заметки.
// Тест создает временный HTTP-сервер, отправляет POST-запрос на создание заметки,
// проверяет успешный статус ответа и сравнивает созданную заметку с ожидаемой.
func TestCreateNote(t *testing.T) {
	initialization()

	//Создание тестового сервера и заметки в формате JSON
	ts := httptest.NewServer(http.HandlerFunc(createNoteHandler))
	defer ts.Close()

	body, err := json.Marshal(note)
	assert.NoError(t, err)

	//Отправка запроса
	resp, err := http.Post(ts.URL+"/note", "application/json", bytes.NewBuffer(body))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	//Парсин данных
	var createdNote Note
	err = json.NewDecoder(resp.Body).Decode(&createdNote)
	assert.NoError(t, err)

	// Проверка созданной заметки
	assert.NotZero(t, createdNote.ID)
	assert.Equal(t, note.Title, createdNote.Title)
	assert.Equal(t, note.Content, createdNote.Content)
	assert.WithinDuration(t, time.Now(), createdNote.CreatedAt, time.Second)
}

// TestGetNote проверяет функцию получения заметки по ее ID.
// Тест создает временный HTTP-сервер, отправляет GET-запрос на получение заметки по ID,
// проверяет успешный статус ответа и сравнивает полученную заметку с ожидаемой.
func TestGetNote(t *testing.T) {
	initialization()

	ts := httptest.NewServer(http.HandlerFunc(getNoteHandler))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/note/1")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var createdNote Note
	err = json.NewDecoder(resp.Body).Decode(&createdNote)
	assert.NoError(t, err)

	assert.Equal(t, 1, createdNote.ID)
	assert.Equal(t, note.Title, createdNote.Title)
	assert.Equal(t, note.Content, createdNote.Content)
	assert.WithinDuration(t, time.Now(), createdNote.CreatedAt, time.Second)
}

// TestGetNotes проверяет функцию получения списка всех заметок.
// Тест создает временный HTTP-сервер, отправляет GET-запрос на получение списка заметок,
// проверяет успешный статус ответа и сравнивает полученную заметку с ожидаемой.
func TestGetNotes(t *testing.T) {
	initialization()

	ts := httptest.NewServer(http.HandlerFunc(getNotesHandler))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/notes")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var notes []Note
	err = json.NewDecoder(resp.Body).Decode(&notes)
	assert.NoError(t, err)
	assert.Equal(t, 1, notes[0].ID)
	assert.Equal(t, note.Title, notes[0].Title)
	assert.Equal(t, note.Content, notes[0].Content)
	assert.WithinDuration(t, time.Now(), notes[0].CreatedAt, time.Second)
}

// TestUpdateNote проверяет функцию обновления заметки.
// Тест создает временный HTTP-сервер, отправляет PATCH-запрос на обновление заметки,
// проверяет успешный статус ответа и сравнивает обновленную заметку с ожидаемой.
func TestUpdateNote(t *testing.T) {
	initialization()

	ts := httptest.NewServer(http.HandlerFunc(updateNoteHandler))
	defer ts.Close()

	body, err := json.Marshal(updatedNote)
	assert.NoError(t, err)

	req, err := http.NewRequest("PATCH", ts.URL+"/note/1", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedResponse Note
	err = json.NewDecoder(resp.Body).Decode(&updatedResponse)
	assert.NoError(t, err)

	assert.Equal(t, 1, updatedResponse.ID)
	assert.Equal(t, updatedNote.Title, updatedResponse.Title)
	assert.Equal(t, updatedNote.Content, updatedResponse.Content)
	assert.WithinDuration(t, time.Now(), updatedResponse.CreatedAt, time.Second)
}

// TestDeleteNote проверяет функцию удаления заметки.
// Тест создает временный HTTP-сервер, отправляет DELETE-запрос на удаление заметки,
// проверяет успешный статус ответа и удостоверяется, что заметка удалена из базы данных.
func TestDeleteNote(t *testing.T) {
	initialization()

	ts := httptest.NewServer(http.HandlerFunc(deleteNoteHandler))

	req, err := http.NewRequest("DELETE", ts.URL+"/note/1", nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверяем, что заметка успешно удалена из базы данных
	checkDeletedNote := func() {
		getNoteResp, err := http.Get(ts.URL + "/note/1")
		assert.NoError(t, err)
		defer getNoteResp.Body.Close()

		//Должны получить http.StatusNotFound, так как заметка должна быть удалена
		assert.Equal(t, http.StatusNotFound, getNoteResp.StatusCode)
	}
	checkDeletedNote()
}
