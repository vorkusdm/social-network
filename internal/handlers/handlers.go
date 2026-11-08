package handlers

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "strconv"

    "golang.org/x/crypto/bcrypt"
    "social_network/internal/db"
    "social_network/internal/models"
)

type Handler struct{}

func NewHandler() *Handler {
    return &Handler{}
}

// --- Middleware для проверки токена ---
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
	    http.Error(w, "Unauthorized", http.StatusUnauthorized)
	    return
	}

	var userID int
	err := db.Pool.QueryRow(r.Context(), 
	    "SELECT user_id FROM sessions WHERE token = $1 AND expires_at > NOW()", token).Scan(&userID)
	
	if err != nil {
	    http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
	    return
	}
	
	// Передаем userID в контексте, чтобы хендлер знал, кто авторизован
	ctx := r.Context()
	ctx = context.WithValue(ctx, "userID", userID)
	next.ServeHTTP(w, r.WithContext(ctx))
    }
}

// --- Регистрация ---
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
	http.Error(w, "Invalid JSON", http.StatusBadRequest)
	return
    }

    // Хешируем пароль (безопасное хранение)
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
	return
    }

    // Параметризованный запрос (защита от SQL-инъекций)
    query := `INSERT INTO users (login, first_name, last_name, birth_date, gender, interests, city, password_hash) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
    
    err = db.Pool.QueryRow(r.Context(), query, 
	user.Login, user.FirstName, user.LastName, user.BirthDate, 
	user.Gender, user.Interests, user.City, string(hashedPassword)).Scan(&user.ID)

    if err != nil {
	http.Error(w, "User already exists or invalid data", http.StatusConflict)
	return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]int{"id": user.ID})
}

// --- Авторизация ---
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var req models.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	http.Error(w, "Invalid JSON", http.StatusBadRequest)
	return
    }

    var userID int
    var passwordHash string
    
    query := `SELECT id, password_hash FROM users WHERE login = $1`
    err := db.Pool.QueryRow(r.Context(), query, req.Login).Scan(&userID, &passwordHash)
    if err != nil {
	http.Error(w, "Invalid login or password", http.StatusUnauthorized)
	return
    }

    // Проверка пароля
    if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
	http.Error(w, "Invalid login or password", http.StatusUnauthorized)
	return
    }

    // Генерируем случайный токен
    tokenBytes := make([]byte, 32)
    rand.Read(tokenBytes)
    token := hex.EncodeToString(tokenBytes)

    // Сохраняем сессию в БД на 24 часа
    expiresAt := time.Now().Add(24 * time.Hour)
    query = `INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)`
    _, err = db.Pool.Exec(r.Context(), query, token, userID, expiresAt)
    if err != nil {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
	return
    }

    json.NewEncoder(w).Encode(models.LoginResponse{Token: token})
}

// --- Получение анкеты ---
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    
    var id int
    if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
	http.Error(w, "Invalid ID", http.StatusBadRequest)
	return
    }

    var user models.User
    // Параметризованный запрос
    query := `SELECT id, login, first_name, last_name, birth_date, gender, interests, city 
              FROM users WHERE id = $1`
    
    err := db.Pool.QueryRow(r.Context(), query, id).Scan(
	&user.ID, &user.Login, &user.FirstName, &user.LastName, 
	&user.BirthDate, &user.Gender, &user.Interests, &user.City,
    )

    if err != nil {
	http.Error(w, "User not found", http.StatusNotFound)
	return
    }

    json.NewEncoder(w).Encode(user)
}

// --- Поиск анкет по префиксу имени и фамилии ---
func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
    firstNamePrefix := r.URL.Query().Get("first_name")
    lastNamePrefix := r.URL.Query().Get("last_name")
    limitStr := r.URL.Query().Get("limit")

    if firstNamePrefix == "" || lastNamePrefix == "" {
	http.Error(w, "first_name and last_name are required", http.StatusBadRequest)
	return
    }

    limit := 50
    if limitStr != "" {
	if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
	    limit = parsedLimit
	}
    }

    // сначала находим id через индекс, потом подтягиваем данные
    query := `SELECT u.id, u.login, u.first_name, u.last_name, u.birth_date, u.gender, u.interests, u.city 
              FROM (
                  SELECT id FROM users 
                  WHERE first_name LIKE $1 AND last_name LIKE $2 
                  ORDER BY id 
                  LIMIT $3
              ) AS subq
              JOIN users u ON u.id = subq.id
              ORDER BY u.id`

    rows, err := db.Pool.Query(r.Context(), query, firstNamePrefix+"%", lastNamePrefix+"%", limit)
    if err != nil {
	http.Error(w, "Database error", http.StatusInternalServerError)
	return
    }
    defer rows.Close()

    var users []models.User
    for rows.Next() {
	var u models.User
	err := rows.Scan(&u.ID, &u.Login, &u.FirstName, &u.LastName, &u.BirthDate, &u.Gender, &u.Interests, &u.City)
	if err != nil {
	    http.Error(w, "Scan error", http.StatusInternalServerError)
	    return
	}
	users = append(users, u)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}