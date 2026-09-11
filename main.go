package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Book struct {
	IDBuku   int    `json:"id_buku"`
	Judul    string `json:"judul"`
	Penulis  string `json:"penulis"`
	Tahun    int    `json:"tahun"`
	Stok     int    `json:"stok"`
	Penerbit string `json:"penerbit,omitempty"`
	Harga    int    `json:"harga,omitempty"`
}

// Pointer membedakan field yang tidak dikirim dengan nilai 0 pada PUT.
type bookInput struct {
	IDBuku   *int    `json:"id_buku"`
	Judul    *string `json:"judul"`
	Penulis  *string `json:"penulis"`
	Tahun    *int    `json:"tahun"`
	Stok     *int    `json:"stok"`
	Penerbit *string `json:"penerbit"`
	Harga    *int    `json:"harga"`
}

const dataFile = "data/books.json"

// ponytail: satu proses mengunci seluruh akses file; gunakan database jika perlu banyak proses.
var booksMu sync.Mutex

func readJSON() ([]Book, error) {
	data, err := os.ReadFile(dataFile)
	if os.IsNotExist(err) {
		return []Book{}, nil
	}
	if err != nil {
		return nil, err
	}
	books := []Book{}
	if err := json.Unmarshal(data, &books); err != nil {
		return nil, err
	}
	if books == nil {
		return nil, fmt.Errorf("data buku harus berupa array")
	}
	return books, nil
}

func writeJSON(books []Book) error {
	data, err := json.MarshalIndent(books, "", "    ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// Tulis ke file sementara agar kegagalan menulis tidak memotong data lama.
	file, err := os.CreateTemp(dir, ".books-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(append(data, '\n'))
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), dataFile)
}

func writeResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("gagal menulis response: %v", err)
	}
}

func writeMessage(w http.ResponseWriter, status int, message string) {
	writeResponse(w, status, map[string]string{"message": message})
}

func readInput(w http.ResponseWriter, r *http.Request) (bookInput, error) {
	var input bookInput
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return input, fmt.Errorf("gagal membaca body atau body melebihi 1 MiB")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return input, fmt.Errorf("body harus berisi satu objek JSON")
	}
	for name, value := range fields {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return input, fmt.Errorf("%s tidak boleh null", name)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, fmt.Errorf("tipe data atau nama field tidak valid")
	}
	return input, nil
}

func checkBook(input bookInput, requireAll bool) error {
	for _, field := range []struct {
		name  string
		value *int
		min   int
	}{
		{"id_buku", input.IDBuku, 1},
		{"tahun", input.Tahun, 1},
		{"stok", input.Stok, 0},
		{"harga", input.Harga, 1},
	} {
		if field.value == nil {
			if requireAll || field.name == "id_buku" {
				return fmt.Errorf("field %s wajib ada", field.name)
			}
		} else if *field.value < field.min {
			return fmt.Errorf("%s minimal %d", field.name, field.min)
		}
	}
	for _, field := range []struct {
		name  string
		value *string
	}{
		{"judul", input.Judul},
		{"penulis", input.Penulis},
		{"penerbit", input.Penerbit},
	} {
		if field.value == nil {
			if requireAll {
				return fmt.Errorf("field %s wajib ada", field.name)
			}
		} else if strings.TrimSpace(*field.value) == "" {
			return fmt.Errorf("%s tidak boleh kosong", field.name)
		}
	}
	return nil
}

func applyInput(book *Book, input bookInput) {
	if input.IDBuku != nil {
		book.IDBuku = *input.IDBuku
	}
	if input.Judul != nil {
		book.Judul = *input.Judul
	}
	if input.Penulis != nil {
		book.Penulis = *input.Penulis
	}
	if input.Tahun != nil {
		book.Tahun = *input.Tahun
	}
	if input.Stok != nil {
		book.Stok = *input.Stok
	}
	if input.Penerbit != nil {
		book.Penerbit = *input.Penerbit
	}
	if input.Harga != nil {
		book.Harga = *input.Harga
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeMessage(w, http.StatusNotFound, "route tidak ditemukan")
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeMessage(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}
	writeResponse(w, http.StatusOK, map[string]any{
		"message": "API Sistem Perpustakaan",
		"routes": []string{
			"GET /books",
			"GET /books?id=<id_buku>",
			"GET /books?penulis=<nama_penulis>",
			"POST /books",
			"PUT /books",
			"DELETE /books",
		},
	})
}

func getBooksHandler(w http.ResponseWriter, r *http.Request, books []Book) {
	query := r.URL.Query()
	if query.Has("id") {
		id, err := strconv.Atoi(query.Get("id"))
		if err != nil || id < 1 {
			writeMessage(w, http.StatusBadRequest, "id_buku harus berupa angka positif")
			return
		}
		for _, book := range books {
			if book.IDBuku == id {
				writeResponse(w, http.StatusOK, book)
				return
			}
		}
		writeMessage(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	if query.Has("penulis") {
		result := []Book{}
		for _, book := range books {
			if strings.EqualFold(book.Penulis, query.Get("penulis")) {
				result = append(result, book)
			}
		}
		if len(result) == 0 {
			writeMessage(w, http.StatusNotFound, "data tidak ditemukan")
			return
		}
		writeResponse(w, http.StatusOK, result)
		return
	}
	writeResponse(w, http.StatusOK, books)
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete:
	default:
		w.Header().Set("Allow", "GET, POST, PUT, DELETE")
		writeMessage(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}

	var input bookInput
	if r.Method != http.MethodGet {
		var err error
		input, err = readInput(w, r)
		if err == nil {
			err = checkBook(input, r.Method == http.MethodPost)
		}
		if err != nil {
			writeMessage(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	booksMu.Lock()
	defer booksMu.Unlock()
	books, err := readJSON()
	if err != nil {
		writeMessage(w, http.StatusInternalServerError, "gagal membaca data")
		return
	}
	if r.Method == http.MethodGet {
		getBooksHandler(w, r, books)
		return
	}

	index := -1
	for i, book := range books {
		if book.IDBuku == *input.IDBuku {
			index = i
			break
		}
	}
	status, message := http.StatusOK, ""
	switch r.Method {
	case http.MethodPost:
		if index >= 0 {
			writeMessage(w, http.StatusConflict, "id_buku sudah ada")
			return
		}
		var book Book
		applyInput(&book, input)
		books = append(books, book)
		status, message = http.StatusCreated, "Buku berhasil ditambahkan"
	case http.MethodPut, http.MethodDelete:
		if index < 0 {
			writeMessage(w, http.StatusNotFound, "data tidak ditemukan")
			return
		}
		if r.Method == http.MethodPut {
			applyInput(&books[index], input)
			message = "Data berhasil diperbarui"
		} else {
			books = append(books[:index], books[index+1:]...)
			message = "Data berhasil dihapus"
		}
	}
	if err := writeJSON(books); err != nil {
		writeMessage(w, http.StatusInternalServerError, "gagal menyimpan data")
		return
	}
	writeMessage(w, status, message)
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/books", booksHandler)
	return mux
}

func main() {
	server := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           newHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Println("Server berjalan di http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
