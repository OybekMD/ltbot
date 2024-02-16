package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/redis/go-redis/v9"
)

// Login start
const (
	cooldownDuration = time.Minute
	waitMessage      = "🕐 Wait, you already got a code. \n1 minute later you might get a code"
)

type UserCooldown struct {
	LastSent time.Time
	sync.Mutex
}

var cooldowns = make(map[int]UserCooldown)
var cooldownsMutex sync.Mutex

func canSendCode(userID int) bool {
	cooldownsMutex.Lock()
	defer cooldownsMutex.Unlock()

	if cooldown, ok := cooldowns[userID]; ok {
		cooldown.Lock()
		defer cooldown.Unlock()
		if time.Since(cooldown.LastSent) < cooldownDuration {
			return false
		}
	}

	return true
}

func updateCooldown(userID int) {
	cooldownsMutex.Lock()
	defer cooldownsMutex.Unlock()

	if cooldown, ok := cooldowns[userID]; ok {
		cooldown.Lock()
		defer cooldown.Unlock()
		cooldown.LastSent = time.Now()
	} else {
		cooldowns[userID] = UserCooldown{LastSent: time.Now()}
	}
}

func generateCode(length int) string {
	const charset = "0123456789"
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}

	return string(code)
}

// Login end

// Sign Start
func GetAll(r redis.Client, ukey string) bool {
	keys, err := r.Keys(context.Background(), "*").Result()
	if err != nil {
		fmt.Println("Get keys error")
		return false
	}
	fmt.Println("All Keys:")
	for _, key := range keys {
		if ukey == key {
			return true
		}
	}
	return false
}

// Sign End

type User struct {
	Id          string
	FirstName   string
	LastName    string
	PhoneNumber string
}

func main() {
	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Replace "YOUR_BOT_TOKEN" with your actual bot token
	bot, err := tgbotapi.NewBotAPI("Your_token")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Set up updates configuration
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Get updates channel
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore non-Message updates
			continue
		}
		if update.Message.IsCommand() && update.Message.Command() == "start" {
			username := update.Message.From.FirstName

			// Format the message with the username
			responseMessage := fmt.Sprintf("Hello👋 %s\nRegister of Articlely\n Sign for /sign \n Login for /login", username)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseMessage)
			_, err := bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
		}

		if update.Message.IsCommand() && update.Message.Command() == "login" {
			userID := update.Message.From.ID
			if canSendCode(userID) {
				randomNumber := generateCode(6)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("🔐 your code: %s", randomNumber))
				_, err := bot.Send(msg)
				if err != nil {
					log.Println("Error sending message:", err)
				}

				updateCooldown(userID)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, waitMessage)
				_, err := bot.Send(msg)
				if err != nil {
					log.Println("Error sending wait message:", err)
				}
			}
		}

		if update.Message.IsCommand() && update.Message.Command() == "sign" {
			skey := strconv.Itoa(update.Message.From.ID)
			if GetAll(*rdb, skey) {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You aready signed!")
				_, err := bot.Send(msg)
				if err != nil {
					log.Println(err)
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please provide your contact:")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButtonContact("Share Contact"),
					),
				)
				bot.Send(msg)
			}
		}

		if update.Message.Contact != nil {
			contact := update.Message.Contact
			userID := contact.UserID
			firstName := contact.FirstName
			lastName := contact.LastName
			phoneNumber := contact.PhoneNumber

			//{}
			// Send message to user
			// result := fmt.Sprintf("User ID: %d\nphoneNumber: %s\nfirstName: %s\nlastName: %s\n", userID, phoneNumber, firstName, lastName)
			// message := tgbotapi.NewMessage(update.Message.Chat.ID, result)
			// See result in console
			// log.Printf("Received update: %+v", update)
			// _, err := bot.Send(message)
			// if err != nil {
			// 	log.Println(err)
			// }
			//{}

			user := User{
				Id:          strconv.Itoa(userID),
				FirstName:   firstName,
				LastName:    lastName,
				PhoneNumber: phoneNumber,
			}

			// Marshal
			byteData, err := json.Marshal(&user)
			if err != nil {
				log.Fatal(err)
				return
			}

			// Create Data
			err = rdb.Set(context.Background(), user.Id, byteData, 0).Err()
			if err != nil {
				fmt.Println("\n\n\n\nerrror\n")
				log.Fatal(err)
				return
			}
		}
	}
}
