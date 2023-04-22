package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type UserT struct {
	ID        int
	Username  string
	FirstName string
	LastName  string
	ReqDate   int
}

type ApiBotsT struct {
	UsersCount   int
	BotsContents []BotsContentT
}

type BotsContentT struct {
	Name     string
	Messages []MessagesBotT
}

type MessagesBotT struct {
	UserId      int
	Username    string
	FirstName   string
	Last_name   string
	Content     string
	DateTime    int
	IsImportant int8
}

type BotsMySql struct {
	Id        int
	Bot_id    string
	Name      string
	Is_active int
}

var UsersDB = make(map[int]UserT)

func main() {

	ApiBots := ApiBotsT{}

	//подключение к БД
	db, err := sql.Open("mysql", "root:nordic123@tcp(mysql:3306)/inordic")
	if err != nil {
		fmt.Println("НЕ подключились к БД", err)
	}
	fmt.Println("подключились к БД")
	//делаем запросы к базе, чтобы получить данные и построить API

	//получим данные ботов
	rows, err := db.Query("select * from bots")
	if err != nil {
		fmt.Println("Что-то не так с rows", err)
	}
	fmt.Println("получили ботов")
	defer rows.Close()

	bots := []BotsMySql{}

	for rows.Next() {
		b := BotsMySql{}
		err := rows.Scan(&b.Id, &b.Bot_id, &b.Name, &b.Is_active)
		if err != nil {
			fmt.Println("ошибка в BotsMySql rows", err)
			continue
		}
		bots = append(bots, b)
	}
	fmt.Println("собрали bots")

	// выводим нашу апишку при запросе с указанного адреса
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		//считываем из файла, чтобы посчитать кол-во юзеров
		dataDb, err := ioutil.ReadFile("../backend/dbUsers.json")
		if err != nil {
			fmt.Println("не могу считатть файл", err)
		}
		json.Unmarshal(dataDb, &UsersDB)
		fmt.Println(dataDb)
		ApiBots.UsersCount = 20
		//ApiBots.UsersCount = len(UsersDB)
		//переменная, куда будем собирать апи
		botsAPI := []BotsContentT{}

		//теперь для каждого бота соберём его сообщения
		for i := 0; i < len(bots); i++ {

			//сначала возьмём имя бота

			bot := BotsContentT{}

			bot.Name = bots[i].Name

			//сделаем запрос в базу по bot_id
			rows, err := db.Query("select messages.user_id, users.username, users.first_name, users.last_name, messages.content, messages.c_time, messages.is_important from messages, users where users.id=messages.user_id and bot_id=?", bots[i].Id)

			if err != nil {
				fmt.Println("Что-то не так с rows", err)
			}

			//?????
			defer rows.Close()

			//соберём сообщения этого бота в апишку
			for rows.Next() {

				m := MessagesBotT{}
				err := rows.Scan(&m.UserId, &m.Username, &m.FirstName, &m.Last_name, &m.Content, &m.DateTime, &m.IsImportant)
				if err != nil {
					fmt.Println(err)
					continue
				}

				//если юзернейм пустой, то берём данные из другого поля и дублируем их в юзернейм
				if m.Username == "" {
					if m.FirstName != "" {
						m.Username = m.FirstName
					} else if m.Last_name != "" {
						m.Username = m.Last_name
					} else {
						m.Username = strconv.Itoa(m.UserId)
					}
				}

				bot.Messages = append(bot.Messages, m)
			}

			botsAPI = append(botsAPI, bot)
		}
		ApiBots.BotsContents = botsAPI

		//распарсим апишку
		JsonBotsAPI, _ := json.Marshal(ApiBots)

		//разрешим подключаться из браузера
		w.Header().Set("Access-Control-Allow-Origin", "*")

		//выдаём апишку
		fmt.Fprintf(w, string(JsonBotsAPI))

	})
	http.ListenAndServe(":80", nil)

}
