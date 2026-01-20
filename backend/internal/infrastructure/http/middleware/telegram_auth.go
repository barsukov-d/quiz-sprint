package middleware

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

// TelegramAuthMiddleware validates Telegram Mini App init data
// Expects "Authorization: tma <base64-encoded-init-data-raw>" header
func TelegramAuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 1. Extract Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Missing Authorization header")
		}

		// 2. Parse "tma <init-data-raw>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "tma" {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid Authorization header format. Expected: 'tma <init-data-raw>'")
		}

		encodedInitData := parts[1]
		if encodedInitData == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Empty init data")
		}

		// 3. Decode from Base64
		decodedInitData, err := base64.StdEncoding.DecodeString(encodedInitData)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid Base64 in init data: "+err.Error())
		}
		initDataRaw := string(decodedInitData)

		// 4. Get bot token from environment
		botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
		if botToken == "" {
			return fiber.NewError(fiber.StatusInternalServerError, "Bot token not configured")
		}

		// 5. Validate init data signature and expiration
		// Telegram considers init data valid for 1 hour by default
		expiration := 1 * time.Hour

		fmt.Printf("🔍 Validating init data:\n")
		// Only log first 100 chars for brevity
		loggableInitData := initDataRaw
		if len(loggableInitData) > 100 {
			loggableInitData = loggableInitData[:100] + "..."
		}
		fmt.Printf("  Raw Decoded: %s\n", loggableInitData)
		fmt.Printf("  Bot token: %s...\n", botToken[:min(20, len(botToken))])
		fmt.Printf("  Expiration: %v\n", expiration)

		if err := initdata.Validate(initDataRaw, botToken, expiration); err != nil {
			fmt.Printf("❌ Validation failed: %v\n", err)
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired init data: "+err.Error())
		}

		fmt.Println("✅ Validation successful!")

		// 6. Parse validated init data
		parsedData, err := initdata.Parse(initDataRaw)
		if err != nil {
			fmt.Printf("❌ Parse failed: %v\n", err)
			return fiber.NewError(fiber.StatusUnauthorized, "Failed to parse init data: "+err.Error())
		}

		fmt.Printf("✅ Parsed init data: user_id=%d, username=%s\n", parsedData.User.ID, parsedData.User.Username)

		// 7. Store parsed init data in request context for handlers
		c.Locals("telegram_init_data", &parsedData)

		fmt.Println("✅ Stored in context, calling next handler...")

		return c.Next()
	}
}

// GetTelegramInitData retrieves validated init data from request context
func GetTelegramInitData(c fiber.Ctx) *initdata.InitData {
	data, ok := c.Locals("telegram_init_data").(*initdata.InitData)
	if !ok {
		return nil
	}
	return data
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
