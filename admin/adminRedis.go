package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type User struct {
	Id          string
	FirstName   string
	LastName    string
	PhoneNumber string
}

func GetUser(rdb redis.Client, key string) {
	respUser, err := rdb.Get(context.Background(), key).Result()
	if err != nil {
		log.Fatal(err)
		return
	}
	var newUser User
	if err := json.Unmarshal([]byte(respUser), &newUser); err != nil {
		log.Fatal(err)
		return
	} else {
		fmt.Println(newUser)
	}
}

func GetAll(r redis.Client) {
	keys, err := r.Keys(context.Background(), "*").Result()
	if err != nil {
		fmt.Println("Get keys error")
		return 
	}
	fmt.Println("All Keys:")
	for _, key := range keys {
		fmt.Println(key)
	}



	for _, userID := range keys {
		respUser, err := r.Get(context.Background(), userID).Result()
		if err != nil {
			log.Println("Error retrieving user with ID:", userID, err)
			continue
		}

		var newUser User
		if err := json.Unmarshal([]byte(respUser), &newUser); err != nil {
			log.Println("Error unmarshalling user with ID:", userID, err)
			continue
		}
		fmt.Sprintf("{\nUserID: %s\nFirstName: %s\n LastName: %s\n PhoneNumber: %s\n}", newUser.Id, newUser.FirstName, newUser.LastName, newUser.PhoneNumber)
	}
}

func DelUser(r redis.Client, key string) {
	res := r.Del(context.Background(), key)
	fmt.Println(res)
}



func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	for {
		var step string
		fmt.Print("GETALL -> 1\n GET -> 2\n DEL -> 3: \n EXIT -> 4\n Enter:")
		fmt.Scan(&step)
		if step == "1" {
			GetAll(*rdb)
		} else if step == "2" {
			fmt.Print("Enter user id: ")
			var ukey string
			fmt.Scan(&ukey)
			GetUser(*rdb, ukey)
		} else if step == "3" {
			fmt.Print("Enter user id: ")
			var ukey string
			fmt.Scan(&ukey)
			DelUser(*rdb, ukey)
		} else if step == "4" {
			fmt.Println("Program End !")
			break
		} else {
			fmt.Println("unknown enter command!\n")
		}
	}
}
