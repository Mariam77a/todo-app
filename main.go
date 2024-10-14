package main

import (
    "database/sql"
    "log"
    "net/http"
    "strconv"
    "time"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/golang-jwt/jwt/v4"
    _ "github.com/mattn/go-sqlite3"
)

var jwtKey = []byte("your_secret_key")
var db *sql.DB

type Credentials struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type Claims struct {
    Email string `json:"email"`
    jwt.StandardClaims
}

type Todo struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
}

func initDB() {
    var err error
    db, err = sql.Open("sqlite3", "./todo.db")
    if err != nil {
        log.Fatal(err)
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            email TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS todos (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL
        );
    `)
    if err != nil {
        log.Fatal(err)
    }
}

func register(c echo.Context) error {
    var creds Credentials
    if err := c.Bind(&creds); err != nil {
        return c.JSON(http.StatusBadRequest, "Invalid input")
    }

    _, err := db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", creds.Email, creds.Password)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Could not register user: "+err.Error())
    }

    expirationTime := time.Now().Add(5 * time.Minute)
    claims := &Claims{
        Email: creds.Email,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Could not create token")
    }

    return c.JSON(http.StatusOK, echo.Map{"token": tokenString})
}

func login(c echo.Context) error {
    var creds Credentials
    if err := c.Bind(&creds); err != nil {
        return c.JSON(http.StatusBadRequest, "Invalid input")
    }

    var password string
    err := db.QueryRow("SELECT password FROM users WHERE email = ?", creds.Email).Scan(&password)
    if err != nil {
        if err == sql.ErrNoRows {
            return c.JSON(http.StatusUnauthorized, "Invalid credentials")
        }
        return c.JSON(http.StatusInternalServerError, "Database error: "+err.Error())
    }

    if password != creds.Password {
        return c.JSON(http.StatusUnauthorized, "Invalid credentials")
    }

    expirationTime := time.Now().Add(5 * time.Minute)
    claims := &Claims{
        Email: creds.Email,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Could not create token")
    }

    return c.JSON(http.StatusOK, echo.Map{"token": tokenString})
}

func getTodos(c echo.Context) error {
    rows, err := db.Query("SELECT id, title FROM todos")
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Database error: "+err.Error())
    }
    defer rows.Close()

    var todos []Todo
    for rows.Next() {
        var todo Todo
        if err := rows.Scan(&todo.ID, &todo.Title); err != nil {
            return c.JSON(http.StatusInternalServerError, "Database error: "+err.Error())
        }
        todos = append(todos, todo)
    }

    return c.JSON(http.StatusOK, todos)
}

func addTodo(c echo.Context) error {
    var todo Todo
    if err := c.Bind(&todo); err != nil {
        return c.JSON(http.StatusBadRequest, "Invalid input")
    }

    _, err := db.Exec("INSERT INTO todos (title) VALUES (?)", todo.Title)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Could not add todo: "+err.Error())
    }

    return c.JSON(http.StatusCreated, todo)
}

func updateTodo(c echo.Context) error {
    idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil {
        return c.JSON(http.StatusBadRequest, "Invalid ID")
    }

    var todo Todo
    if err := c.Bind(&todo); err != nil {
        return c.JSON(http.StatusBadRequest, "Invalid input")
    }

    _, err = db.Exec("UPDATE todos SET title = ? WHERE id = ?", todo.Title, id)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Could not update todo: "+err.Error())
    }

    return c.JSON(http.StatusOK, todo)
}

func deleteTodo(c echo.Context) error {
    idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil {
        return c.JSON(http.StatusBadRequest, "Invalid ID")
    }

    _, err = db.Exec("DELETE FROM todos WHERE id = ?", id)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, "Could not delete todo: "+err.Error())
    }

    return c.NoContent(http.StatusNoContent)
}

func main() {
    initDB()

    e := echo.New()

    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
    }))

    e.POST("/login", login)
    e.POST("/register", register)

    jwtGroup := e.Group("")
    jwtGroup.Use(middleware.JWTWithConfig(middleware.JWTConfig{
        SigningKey: jwtKey,
    }))

    jwtGroup.GET("/todos", getTodos)
    jwtGroup.POST("/todos", addTodo)
    jwtGroup.PUT("/todos/:id", updateTodo)
    jwtGroup.DELETE("/todos/:id", deleteTodo)

    e.Logger.Fatal(e.Start(":8080"))
}
