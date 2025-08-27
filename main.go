package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Account struct {
	RealName string `json:"realname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var rdb *redis.Client

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "redis-10846.c256.us-east-1-2.ec2.redns.redis-cloud.com:10846",
		Password: "UrYNfii39pHxYuQpHiv3kMHuouUkoZ9M",
		DB:       0,
	})

	app := fiber.New()

	app.Post("/register", register)
	app.Post("/login", login)

	log.Fatal(app.Listen(":3000"))
}

func register(c *fiber.Ctx) error {
	type Req struct {
		Username string `json:"username"`
		RealName string `json:"realname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	hash := sha1.New()
	hash.Write([]byte(req.Password))
	hashedPassword := hex.EncodeToString(hash.Sum(nil))

	user := Account{
		RealName: req.RealName,
		Email:    req.Email,
		Password: hashedPassword,
	}

	key := fmt.Sprintf("login_%s", req.Username)
	data, _ := json.Marshal(user)
	if err := rdb.Set(ctx, key, data, 0).Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal menyimpan account"})
	}

	return c.JSON(fiber.Map{"message": "sudah terdaftar"})
}

func login(c *fiber.Ctx) error {
	type Req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req Req

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	key := fmt.Sprintf("login_%s", req.Username)
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return c.Status(401).JSON(fiber.Map{"error": "user tidak ditemukan"})
	} else if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "redis error"})
	}

	var user Account
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to parse data"})
	}

	hash := sha1.New()
	hash.Write([]byte(req.Password))
	hashedInput := hex.EncodeToString(hash.Sum(nil))

	if hashedInput != user.Password {
		return c.Status(401).JSON(fiber.Map{"error": "password salah"})
	}

	return c.JSON(fiber.Map{
		"message":  "login success",
		"realname": user.RealName,
		"email":    user.Email,
	})
}
