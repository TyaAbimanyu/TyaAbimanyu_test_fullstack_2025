package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// Struktur data pengguna
type User struct {
	RealName string `json:"realname"`
	Email    string `json:"email"`
	Password string `json:"password"` // Password akan di-hash SHA1
}

// Struktur permintaan login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Struktur respons login sukses
type LoginResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	RealName string `json:"realname,omitempty"`
	Email    string `json:"email,omitempty"`
}

// Struktur respons error
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

// Inisialisasi koneksi Redis
func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Alamat server Redis
		Password: "",               // Kosong jika tidak pakai password
		DB:       0,                // Gunakan database default
	})

	// Tes koneksi
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Gagal koneksi ke Redis: %v", err)
	}
	log.Println("Berhasil terhubung ke Redis")
}

// Fungsi untuk meng-hash password
func hashPassword(password string) string {
	hasher := sha1.New()
	hasher.Write([]byte(password))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// Simpan data pengguna ke Redis
func storeUser(username string, user User) error {
	key := fmt.Sprintf("login_%s", username)
	user.Password = hashPassword(user.Password)

	userData, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return redisClient.Set(ctx, key, userData, 0).Err()
}

// Ambil data pengguna dari Redis
func getUser(username string) (*User, error) {
	key := fmt.Sprintf("login_%s", username)
	userData, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal([]byte(userData), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Verifikasi password
func verifyPassword(hashedPassword, inputPassword string) bool {
	return hashedPassword == hashPassword(inputPassword)
}

// Handler untuk endpoint login
func loginHandler(c *fiber.Ctx) error {
	var loginReq LoginRequest

	if err := c.BodyParser(&loginReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Message: "Format permintaan tidak valid",
		})
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Message: "Username dan password wajib diisi",
		})
	}

	user, err := getUser(loginReq.Username)
	if err != nil {
		if err == redis.Nil {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
				Success: false,
				Message: "Username atau password salah",
			})
		}
		log.Printf("Gagal mengambil data pengguna: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Success: false,
			Message: "Terjadi kesalahan server",
		})
	}

	if !verifyPassword(user.Password, loginReq.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Success: false,
			Message: "Username atau password salah",
		})
	}

	return c.JSON(LoginResponse{
		Success:  true,
		Message:  "Login berhasil",
		RealName: user.RealName,
		Email:    user.Email,
	})
}

// Tambahkan data pengguna contoh
func insertSampleUsers() {
	users := []struct {
		username string
		user     User
	}{
		{
			username: "john_doe",
			user: User{
				RealName: "John Doe",
				Email:    "john.doe@example.com",
				Password: "password123",
			},
		},
		{
			username: "jane_smith",
			user: User{
				RealName: "Jane Smith",
				Email:    "jane.smith@example.com",
				Password: "mypassword",
			},
		},
		{
			username: "admin",
			user: User{
				RealName: "Administrator",
				Email:    "admin@example.com",
				Password: "admin123",
			},
		},
	}

	for _, u := range users {
		err := storeUser(u.username, u.user)
		if err != nil {
			log.Printf("Gagal menyimpan user %s: %v", u.username, err)
		} else {
			log.Printf("Berhasil menyimpan user: %s", u.username)
		}
	}
}

// Jalankan server login
func RunLoginSystem() {
	initRedis()
	insertSampleUsers()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(ErrorResponse{
				Success: false,
				Message: err.Error(),
			})
		},
	})

	app.Post("/login", loginHandler)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"message": "Sistem login berjalan normal",
		})
	})

	log.Println("Menjalankan server di port 3000...")
	log.Fatal(app.Listen(":3000"))
}
