package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const botToken = "1869654383:AAG-9t6gTXd2EcThC_lgx8I9Wt7YDOl7Gjc"

func main() {
	fmt.Println("start")
	err := http.ListenAndServe(":8080", http.HandlerFunc(webHookHandler))
	if err != nil {
		log.Fatal(err)
		return
	}
}

func webHookHandler(rw http.ResponseWriter, req *http.Request) {

	// Create our web hook request body type instance
	body := &webHookReqBody{}

	// Decodes the incoming request into our cutom webhookreqbody type
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		log.Printf("An error occured (webHookHandler)")
		log.Panic(err)
		return
	}

	// If the command /joke is recieved call the sendReply function
	if strings.ToLower(body.Message.Text) == "/love" {
		//fmt.Println("")
		var now = time.Now()
		l, err := time.LoadLocation("Asia/Dushanbe")
		if err != nil {
			log.Println(err)
		}
		var nextDay = time.Date(now.Year(), now.Month(), now.Day(), 15, 58, 0, 0, l)
		sub := nextDay.Sub(now)
		fmt.Println(nextDay)
		//fmt.Println(sub)
		ctx, _ := context.WithCancel(context.Background())
		var wg sync.WaitGroup

		wg.Add(1)

		go Worker(ctx, &wg, sub, func() {
			//err = sendReplyForTime(body.Message.Chat.ID)
			//if err != nil {
			//	log.Printf("An error occured (webHookHandler)" + err.Error())
			//}
			err = sendReply(body.Message.Chat.ID)
			if err != nil {
				log.Printf("An error occured (webHookHandler)" + err.Error())
			}
			time.Sleep(time.Hour*24)
		})

		wg.Wait()

	}

}

type webHookReqBody struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

func sendReply(chatID int64) error {
	fmt.Println("sendReply called")

	// calls the joke fetcher fucntion and gets a random joke from the API
	text, err := jokeFetcher()
	if err != nil {
		return err
	}

	//Creates an instance of our custom sendMessageReqBody Type
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   text,
	}

	// Convert our custom type into json format
	reqBytes, err := json.Marshal(reqBody)

	if err != nil {
		return err
	}

	// Make a request to send our message using the POST method to the telegram bot API
	resp, err := http.Post(
		"https://api.telegram.org/bot"+botToken+"/"+"sendMessage",
		"application/json",
		bytes.NewBuffer(reqBytes),
	)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("unexpected status" + resp.Status)
	}

	return err
}

func jokeFetcher() (string, error) {

	l, err := time.LoadLocation("Asia/Dushanbe")
	if err != nil {
		log.Println(err)
	}

	t := time.Date(2022, 7, 5, 8, 0, 0, 0, l)
	c := &joke{}
	since := time.Since(t)
	s := math.Floor(since.Hours() / 24)
	c.Value.Joke = strconv.Itoa(int(s)*(-1)) + " дней осталось \n" + strconv.Itoa(int(since.Hours()*(-1))) + " часов осталось\n" + strconv.Itoa(int(since.Minutes()*(-1))) + " минут осталось"
	return c.Value.Joke, err
}

type joke struct {
	Value struct {
		Joke string `json:"joke"`
	} `json:"value"`
}

type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func Worker(ctx context.Context, wg *sync.WaitGroup, period time.Duration, callback func()) {
	fmt.Println(period)
	if period > (-1)*time.Minute && 0*time.Second > period {
		return
	}

	ticker := time.NewTicker(period) // 5 * time.Minute
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("StatusCheck - worker is interrupted")
			wg.Done()
			return
		case <-ticker.C:
			callback()
		}
	}
}

func sendReplyForTime(chatID int64) error {
	fmt.Println("sendReply called")

	// calls the joke fetcher fucntion and gets a random joke from the API
	text := "Счетчик заработал, спасибо за это Благородному Рустаму"

	//Creates an instance of our custom sendMessageReqBody Type
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   text,
	}

	// Convert our custom type into json format
	reqBytes, err := json.Marshal(reqBody)

	if err != nil {
		return err
	}

	// Make a request to send our message using the POST method to the telegram bot API
	resp, err := http.Post(
		"https://api.telegram.org/bot"+botToken+"/"+"sendMessage",
		"application/json",
		bytes.NewBuffer(reqBytes),
	)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("unexpected status" + resp.Status)
	}

	return err
}


// Вызов переданной функции раз в сутки в указанное время.
func callAt(hour, min, sec int, f func()) error {
    loc, err := time.LoadLocation("Local")
    if err != nil {
        return err
    }

    // Вычисляем время первого запуска.
    now := time.Now().Local()
    firstCallTime := time.Date(
        now.Year(), now.Month(), now.Day(), hour, min, sec, 0, loc)
    if firstCallTime.Before(now) {
        // Если получилось время раньше текущего, прибавляем сутки.
        firstCallTime = firstCallTime.Add(time.Hour * 24)
    }

    // Вычисляем временной промежуток до запуска.
    duration := firstCallTime.Sub(time.Now().Local())

    go func() {
        time.Sleep(duration)
        for {
            f()
            // Следующий запуск через сутки.
            time.Sleep(time.Hour * 24)
        }
    }()

    return nil
}