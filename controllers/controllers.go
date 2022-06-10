package controllers

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/its-me-debk007/auth-backend/database"
	"github.com/its-me-debk007/auth-backend/models"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *fiber.Ctx) error {
	body := new(models.User)

	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	body.Email = formatEmail(body.Email)
	body.Password = strings.Trim(body.Password, " ")

	user := new(models.User)

	database.DB.First(user, "email = ?", body.Email)

	if user.Email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(models.Message{"user not signed up"})
	}

	if !user.IsVerified {
		return c.Status(fiber.StatusUnauthorized).JSON(models.Message{"user is not verified"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{
			"invalid credentials"})
	}

	accessToken, err := generateToken(body.Email, 60)
	if err != nil {
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	refreshToken, err := generateToken(body.Email, 180)
	if err != nil {
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func Signup(c *fiber.Ctx) error {
	body := new(models.User)

	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	body.Email = formatEmail(body.Email)
	body.Password = strings.Trim(body.Password, " ")
	body.Name = strings.Trim(body.Name, " ")

	user := new(models.User)

	database.DB.First(user, "email = ?", body.Email)
	if user.Email != "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{"email already registered"})
	}

	if _, err := mail.ParseAddress(body.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{"invalid email format"})
	}

	if validation := isValidPassword(body.Password); validation != "ok" {
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{validation})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{err.Error()})
	}
	body.Password = string(hashedPassword)

	if err := database.DB.Create(&body); err.Error != nil {
		log.Println(err.Error)
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{"error in creating user"})
	}

	return c.JSON(models.Message{"user successfully signed up"})
}

func SendOtp(c *fiber.Ctx) error {
	body := new(struct {
		Email string `json:"email"`
	})

	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{err.Error()})
	}

	body.Email = formatEmail(body.Email)

	sender := "11testee11@gmail.com"
	password := os.Getenv("SMTP_PASSWORD")
	receiver := []string{body.Email}
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	bigInt, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		panic(err)
	}
	otp := bigInt.Int64() + 100000
	message := []byte(fmt.Sprint(otp))

	auth := smtp.PlainAuth("", sender, password, smtpHost)

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, sender, receiver, message)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{err.Error()})
	}

	table := new(models.Otps)

	database.DB.First(table, "email = ?", body.Email)
	if table.Email == "" {
		table.Email = body.Email
		table.Otp = int(otp)
		table.CreatedAt = time.Now()

		if err := database.DB.Create(table); err.Error != nil {
			log.Println(err.Error)
			return c.Status(fiber.StatusBadRequest).JSON(models.Message{"error in storing the email in database"})
		}
	} else {
		table.Otp = int(otp)
		table.CreatedAt = time.Now()
		if err := database.DB.Save(table); err.Error != nil {
			log.Println(err.Error)
			return c.Status(fiber.StatusBadRequest).JSON(models.Message{"error in storing the email in database"})
		}
	}

	return c.JSON(models.Message{"sent otp successfully"})
}

func ResetPassword(c *fiber.Ctx) error {
	body := new(models.User)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{err.Error()})
	}

	body.Email = formatEmail(body.Email)

	user := new(models.User)

	database.DB.First(user, "email = ?", body.Email)

	if user.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{"no matching email found"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{err.Error()})
	}

	user.Password = string(hashedPassword)
	database.DB.Save(user)

	return c.JSON(models.Message{"password changed successfully"})
}

func VerifyOtp(c *fiber.Ctx) error {
	body := new(struct {
		Email string `json:"email"`
		Otp   int    `json:"otp"`
	})

	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Message{err.Error()})
	}

	body.Email = formatEmail(body.Email)

	table := new(models.Otps)

	database.DB.First(table, "email = ?", body.Email)

	if body.Otp != table.Otp {
		return c.Status(fiber.StatusUnauthorized).JSON(models.Message{"incorrect otp"})

	} else {
		timeDiff := time.Now().Sub(table.CreatedAt)
		if timeDiff > (time.Minute * 5) {
			return c.Status(fiber.StatusUnauthorized).JSON(models.Message{"otp has expired"})
		}
		user := new(models.User)
		database.DB.First(user, "email = ?", body.Email)
		if !user.IsVerified {
			user.IsVerified = true
			database.DB.Save(user)
		}

		return c.JSON(models.Message{"email has been verified successfully"})
	}
}

func Refresh(c *fiber.Ctx) error {
	body := new(struct {
		Refresh string `json:"refresh_token"`
	})

	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.Message{err.Error()})
	}

	claims, err := verifyToken(body.Refresh)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.Message{err.Error()})
	}

	accessToken, err := generateToken(claims.Issuer, 60)
	if err != nil {
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"accessToken": accessToken,
	})
}

func isValidPassword(password string) string {
	isDigit, isLowercase, isUppercase, isSpecialChar := 0, 0, 0, 0
	for _, ch := range password {
		switch {
		case ch >= '0' && ch <= '9':
			isDigit = 1

		case ch >= 'a' && ch <= 'z':
			isLowercase = 1

		case ch >= 'A' && ch <= 'Z':
			isUppercase = 1

		case ch == '$' || ch == '!' || ch == '@' || ch == '#' || ch == '%' || ch == '&' || ch == '^' || ch == '*' || ch == '/' || ch == '\\':
			isSpecialChar = 1
		}
	}

	switch {
	case len(password) < 8:
		return "password must be at least 8 characters long"

	case isDigit == 0:
		return "password must contain at-least one numeric digit"

	case isLowercase == 0:
		return "password must contain at-least one lowercase alphabet"

	case isUppercase == 0:
		return "password must contain at-least one uppercase alphabet"

	case isSpecialChar == 0:
		return "password must contain at-least one special character"

	default:
		return "ok"
	}
}

func generateToken(username string, expirationTime int) (string, error) {
	standardClaims := jwt.StandardClaims{
		Issuer:    username,
		ExpiresAt: time.Now().Add(time.Minute * time.Duration(expirationTime)).Unix(),
	}
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, standardClaims)

	secretKey := os.Getenv("SECRET_KEY")
	token, err := claims.SignedString([]byte(secretKey))

	if err != nil {
		return token, err
	}

	return token, nil
}

func verifyToken(tokenString string) (*jwt.StandardClaims, error) {
	secretKey := os.Getenv("SECRET_KEY")

	claims := new(jwt.StandardClaims)

	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return claims, nil
}

func formatEmail(email string) string {
	email = strings.Trim(email, " ")
	email = strings.ToLower(email)

	return email
}
