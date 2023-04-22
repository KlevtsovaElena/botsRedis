package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"strconv"
	"strings"
	"time"

	_ "github.com/garyburd/redigo/redis"

	_ "github.com/go-sql-driver/mysql"
)

type ResponseT struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID           int    `json:"id"`
				IsBot        bool   `json:"is_bot"`
				FirstName    string `json:"first_name"`
				LastName     string `json:"last_name"`
				Username     string `json:"username"`
				LanguageCode string `json:"language_code"`
			} `json:"from"`
			Chat struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				Type      string `json:"type"`
			} `json:"chat"`
			Date int    `json:"date"`
			Text string `json:"text"`
		} `json:"message"`
	} `json:"result"`
}

// type UserT struct {
// 	ID         int
// 	Username   string
// 	FirstName  string
// 	LastName   string
// 	Messages   []MessageT
// 	ReqDate    int
// 	LastVisite int
// }

type UserT struct {
	ID        int
	Username  string
	FirstName string
	LastName  string
	ReqDate   int
}

// func (u *UserT) addMessage(text string, messageTime int) {

// 	message := MessageT{}
// 	message.Content = text
// 	message.Date = messageTime
// 	u.Messages = append(u.Messages, message)

// }

func (m *MessageT) addMessage(text string, messageTime int, chatId int) {

	m.UserID = chatId
	m.Content = text
	m.Date = messageTime

}

type MessageT struct {
	UserID  int
	Content string
	Date    int
}

var host string = "https://api.telegram.org/bot"

var tokens = [5]string{
	"6131123688:AAGV7bDvX4aX4_n-ShaiKjXlpUvlnfXsQFY",
	"6266036859:AAGLaQvcjIR8BgkymXNwP0rSfqx2lzQvdmA",
	"6114246715:AAHeEIQBYooYdGG-Dgjqv0jLxPH6zxGJRNY",
	"6089892871:AAHBVa5OpNIg0WYzvIDXj7x8nWqX3n0h6EQ",
	"6025286750:AAHWYyfw1g4-QCP6iopsR5xkMprILA3vdkI",
}

var lastMessage = [5]int{0, 0, 0, 0, 0}

var importansWord = [9]string{
	"срочно",
	"помогите",
	"помощь",
	"помочь",
	"важно",
	"конфликт",
	"неприятн",
	"паник",
	"sos",
}

// создали базу данных сообщений
var MessagesDB = []MessageT{}

// создали базу данных юзеров
var UsersDB = make(map[int]UserT)

// создаем соединение с БД
var Db, Err = sql.Open("mysql", "root:nordic123@tcp(mysql:3306)/inordic")

func main() {

	if Err != nil {
		fmt.Println("НЕ подключились к БД", Err)
	}

	//получили юзеров из ДБ в оперативную память
	rows, err := Db.Query("select * from `users`")

	if err != nil {
		fmt.Println("Что-то не так с rows", err)
	}

	for rows.Next() {
		u := UserT{}
		err := rows.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.ReqDate)
		if err != nil {
			fmt.Println("ошибка в users rows", err)
			continue
		}
		UsersDB[u.ID] = u
	}

	//считываем из бд при включении
	// dataDb, _ := ioutil.ReadFile("db.json")
	// json.Unmarshal(dataDb, &UsersDB)
	// fmt.Println(dataDb)

	getConnectionRedis()

	for range time.Tick(time.Second * 1) {
		//отправляем запрос к Telegram API на получение сообщений длЯ каждого бота
		for j := 0; j < len(tokens); j++ {
			handleBot(j)

		}
	}

}

func getConnectionRedis() {
	//коннектимся к редис
	connRedis1, errRedis1 := redis.Dial("tcp", "redis:6379")
	if errRedis1 != nil {
		fmt.Println("Connect to redis error", errRedis1)
		return
	}
	fmt.Println("подключились к редис")

}

// функция отправки сообщения пользователю
func sendMessage(chatId int, text string, token string) {
	http.Get(host + token + "/sendMessage?chat_id=" + strconv.Itoa(chatId) + "&text=" + text)
}

// функция проверки на важные слова
func checkImportant(text string) bool {

	// проверка сообщений на важные слова
	for i := 0; i < len(importansWord); i++ {
		if strings.Contains(strings.ToLower(text), importansWord[i]) {
			return true
		}
	}
	return false
}

// функция обработки обращений
func handleBot(j int) {

	var url string = host + tokens[j] + "/getUpdates?offset=" + strconv.Itoa(lastMessage[j])

	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	data, _ := ioutil.ReadAll(response.Body)

	//парсим данные из json
	var responseObj ResponseT
	json.Unmarshal(data, &responseObj)

	//считаем количество новых сообщений
	number := len(responseObj.Result)

	//если сообщений нет - то дальше код не выполняем
	if number < 1 {
		return
	}

	fmt.Println("сообщения из ", tokens[j])

	//в цикле доставать инормацию по каждому сообщению
	for i := 0; i < number; i++ {

		text := responseObj.Result[i].Message.Text
		chatId := responseObj.Result[i].Message.From.ID
		messageTime := responseObj.Result[i].Message.Date
		username := responseObj.Result[i].Message.From.Username
		firstName := responseObj.Result[i].Message.From.FirstName
		lastName := responseObj.Result[i].Message.From.LastName

		//определяем зарегистрирован ли пользователь, бд
		// exists := db.QueryRow("select count(id) from `users` where id=?", chatId)
		// if err != nil {
		// 	fmt.Println("Что-то не так с rows", err)
		// }

		// fmt.Println(chatId, " ", exists)

		// var exist int
		// exists.Scan(&exist)

		// fmt.Println(chatId, " ", exist)
		//определяем зарегистрирован ли пользователь
		_, exist := UsersDB[chatId]

		if exist == false {
			user := UserT{}
			user.ID = chatId
			user.Username = username
			user.FirstName = firstName
			user.LastName = lastName
			// user.LastVisite = messageTime
			user.ReqDate = messageTime
			// user.addMessage(text, messageTime)

			//если не зарегистрирован - добавляем в БД и сохраняем в ОП
			_, err := Db.Query("INSERT INTO `users`(`id`,`username`,`first_name`,`last_name`, `date_req`) VALUES(?,?, ?, ?,?)", chatId, username, firstName, lastName, messageTime)
			if err != nil {
				fmt.Println("Ошибка сохранения пользователя ", err)
			} else {
				fmt.Println("пользователь добавлен")
			}

			UsersDB[chatId] = user

		}
		//else {

		//MessagesDB[chatId] = addMessage(text, messageTime)
		// 	user, _ := UsersDB[chatId]
		// 	// user.LastVisite = messageTime
		// 	// user.addMessage(text, messageTime)

		// 	UsersDB[chatId] = user
		//}

		//проверим сообщение на пустоту
		if text == "" {
			continue
		}
		fmt.Println("непустое сообщение")

		is_important := 0
		if checkImportant(text) {
			is_important = 1
		}

		//запись сообщений с БД и оперативку
		_, err := Db.Query("INSERT INTO `messages`(`user_id`,`content`,`c_time`, `bot_id`, `is_important`) VALUES(?,?, ?,?,?)", chatId, text, messageTime, j+1, is_important)
		if err != nil {
			fmt.Println("Ошибка сохранения сообщения ", err)
		} else {
			fmt.Println("сообщение " + text + " добавлено")
		}

		//запишем сообщения в ОП
		message := MessageT{}
		message.addMessage(text, messageTime, chatId)
		MessagesDB = append(MessagesDB, message)

		//сохраняем в файл
		file, _ := os.Create("dbUsers.json")
		jsonUsers, _ := json.Marshal(UsersDB)
		file.Write(jsonUsers)
		file.Close()
		file2, _ := os.Create("dbMessages.json")
		jsonMessages, _ := json.Marshal(MessagesDB)
		file2.Write(jsonMessages)
		file2.Close()
		// file3, err := os.Create("http://localhost:8080/dbUsers.json")
		// if err != nil {
		// 	fmt.Println("Ошибка создания файла ", err)
		// }

		// file3.Write(jsonMessages)
		// file3.Close()
		//отвечаем пользователю на его сообщение
		go sendMessage(chatId, text, tokens[j])

	}

	//запоминаем update_id  последнего сообщения
	lastMessage[j] = responseObj.Result[number-1].UpdateID + 1
	fmt.Println(MessagesDB)
}
