package main

import (
	"database/sql"
	"distcache"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Todo represents a todo item
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Database connection successful")
	}
}

// createGroup creates a new distcache group
func createGroup() *distcache.Group {
	return distcache.NewGroup("todos", 2<<10, distcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			// Fetch todo from database
			todo, err := getTodoFromDB(key)
			if err != nil {
				return nil, err
			}
			// Marshal todo to JSON
			data, err := json.Marshal(todo)
			if err != nil {
				return nil, err
			}
			return data, nil
		}))
}

// getTodoFromDB fetches a todo item from the PostgreSQL database
func getTodoFromDB(key string) (*Todo, error) {
	var todo Todo
	err := db.QueryRow("SELECT id, title, completed, created_at, updated_at FROM todos WHERE id = $1", key).
		Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

// startCacheServer starts the distcache server
func startCacheServer(addr string, addrs []string, dist *distcache.Group) {
	peers := distcache.NewHTTPPool(addr)
	peers.Set(addrs...)
	dist.RegisterPeers(peers)

	log.Println("distcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], nil))
}

// startAPIServer starts the API server
func startAPIServer(apiAddr string, dist *distcache.Group) {
	r := mux.NewRouter()

	r.HandleFunc("/api/todos", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Handle GET request to fetch todos
			getTodosHandler(w, r, dist)
		case http.MethodPost:
			// Handle POST request to create todo
			createTodoHandler(w, r)
		// Add handlers for other CRUD operations: PUT, DELETE
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods(http.MethodGet, http.MethodPost)

	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], r))
}

// getTodosHandler handles GET request to fetch todos
// getTodosHandler handles GET request to fetch todos
func getTodosHandler(w http.ResponseWriter, r *http.Request, dist *distcache.Group) {
	key := r.URL.Query().Get("key")
	fmt.Println("key", key)

	// Check if the request is for all todos without a specific key
	if r.URL.Path == "/api/todos" && key == "" {
		// Fetch all todos from cache
		view, err := dist.Get("0")
		if err != nil {
			// If todos not found in cache, fetch from database
			allTodos, err := getAllTodosFromDB()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Marshal todos to JSON
			data, err := json.Marshal(allTodos)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Respond with all todos
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
			return
		}

		// Respond with all todos from cache
		w.Header().Set("Content-Type", "application/json")
		w.Write(view.ByteSlice())
		return
	}

	// Fetch single todo from cache
	view, err := dist.Get(key)
	if err != nil {
		// If todo not found in cache, fetch from database
		todo, err := getTodoFromDB(key)
		if err != nil {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}

		// Marshal todo to JSON
		data, err := json.Marshal(todo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Respond with the todo
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}

	// Respond with single todo from cache
	w.Header().Set("Content-Type", "application/json")
	w.Write(view.ByteSlice())
}

// getAllTodosFromDB fetches all todos from the PostgreSQL database
func getAllTodosFromDB() ([]Todo, error) {
	rows, err := db.Query("SELECT id, title, completed, created_at, updated_at FROM todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

// createTodoHandler handles POST request to create a todo
func createTodoHandler(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Insert the todo into the database
	if err := insertTodoIntoDB(todo); err != nil {
		http.Error(w, "Failed to create todo in database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// insertTodoIntoDB inserts a todo into the PostgreSQL database
func insertTodoIntoDB(todo Todo) error {
	_, err := db.Exec("INSERT INTO todos (title, completed, created_at, updated_at) VALUES ($1, $2, $3, $4)",
		todo.Title, todo.Completed, todo.CreatedAt, todo.UpdatedAt)
	return err
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "distcache server port")
	flag.BoolVar(&api, "api", false, "Start an API server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	dist := createGroup()
	if api {
		go startAPIServer(apiAddr, dist)
	}
	startCacheServer(addrMap[port], addrs, dist)
}
