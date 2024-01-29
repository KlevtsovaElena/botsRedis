package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
)

type UserT struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ReqDate   int    `json:"date_req"`
}

type MessagesBotT struct {
	UserId      int
	Username    string
	FirstName   string
	LastName    string
	Content     string
	DateTime    int
	IsImportant int8
}

type ApiBotsT struct {
	UsersCount    int
	MessagesCount int
	BotsContents  []BotsContentT
}

type BotsContentT struct {
	Name     string
	Messages []MessagesBotT
}

type BotsMySqlT struct {
	Id        int
	Bot_id    string
	Name      string
	Is_active int
}

type countMessagesT struct {
	CountMessage int
	CountUser    int
}

type OldResponseT struct {
	NotNewMessages bool
}

var UsersDB = make(map[int]UserT)

// подключение к БД
var Db, Err = sql.Open("mysql", "root:admin@tcp(mysql:3306)/messages")

// подключение к редис
var RedisDB, _ = redis.Dial("tcp", "redis:6379")

func main() {

	// проверка подключились ли к БД
	if Err != nil {
		fmt.Println("НЕ подключились к БД", Err)
	} else {
		fmt.Println("подключились к БД")
	}
	// начальное количество сообщений и юзеров устанавливаем в ноль
	CountMessages := countMessagesT{CountMessage: 0, CountUser: 0}

	ApiBots := ApiBotsT{}

	//делаем запросы к базе или redis, чтобы получить данные и построить API
	// **** ДАННЫЕ ЮЗЕРОВ **** //
	// записываем общее кол-во юзеров
	count, errRedis := redis.String(RedisDB.Do("GET", "countUsers"))
	if errRedis != nil {
		fmt.Println("не смогли получить кол-во юзеров из редис", errRedis)
	} else {
		countInt, _ := strconv.Atoi(count)
		CountMessages.CountUser = countInt
		fmt.Println("взяли кол-во юзеров из редис", countInt)
	}

	// **** ДАННЫЕ СООБЩЕНИЙ **** //
	// записываем общее кол-во cсообщений
	count, errRedis = redis.String(RedisDB.Do("GET", "countMessages"))
	if errRedis != nil {
		fmt.Println("не смогли получить кол-во сообщений из редис", errRedis)
	} else {
		countInt, _ := strconv.Atoi(count)
		CountMessages.CountMessage = countInt
		fmt.Println("взяли кол-во сообщений из редис", countInt)
	}

	// **** ДАННЫЕ БОТОВ **** //
	// получим данные БОТОВ
	rows, err := Db.Query("select * from bots")
	if err != nil {
		fmt.Println("Не смогли получить ботов из БД", err)
	}
	fmt.Println("получили ботов")
	defer rows.Close()

	bots := []BotsMySqlT{}

	// разнесем по полям и соберём их всех в переменную bots
	for rows.Next() {
		b := BotsMySqlT{}
		err := rows.Scan(&b.Id, &b.Bot_id, &b.Name, &b.Is_active)
		if err != nil {
			fmt.Println("ошибка сканирования в BotsMySql данных одного бота", err)
			continue
		}
		bots = append(bots, b)
	}
	fmt.Println("собрали bots")
	var JsonBotsAPI []byte

	// выводим нашу апишку при запросе с указанного адреса
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// возьмём кол-во сообщений из redis
		count, errRedis = redis.String(RedisDB.Do("GET", "countMessages"))
		var countInt int
		if errRedis != nil {
			fmt.Println("не смогли получить кол-во сообщений из редис в HandleFunc", errRedis)
		} else {
			countInt, _ = strconv.Atoi(count)
			fmt.Println("взяли кол-во сообщений из редис", countInt)
		}

		fmt.Printf(" CountMessages.CountMessage - %v, countInt - %v, ApiBots.MessagesCount - %v", CountMessages.CountMessage, countInt, ApiBots.MessagesCount)
		// сверяем кол-во из redis со старым кол-вом сообщений в переменной
		// если они разные, значит в базе есть новые сообщения и мы их оттуда достаём
		// если нет - то отправляем ответ, что новых сообщений нет
		if CountMessages.CountMessage == countInt && ApiBots.MessagesCount != 0 {
			fmt.Println("Выдаём старую апишку без запроса в базу")

			//разрешим подключаться из браузера
			w.Header().Set("Access-Control-Allow-Origin", "*")

			//выдаём старую апишку
			fmt.Fprintf(w, string(JsonBotsAPI))

		} else {
			fmt.Println("Выдаём новую апишку с запросом в базу")

			CountMessages.CountMessage = countInt
			// возьмём кол-во юзеров из redis
			count, errRedis = redis.String(RedisDB.Do("GET", "countUsers"))
			var countUsers int
			if errRedis != nil {
				fmt.Println("не смогли получить кол-во users из редис в HandleFunc", errRedis)
			} else {
				countUsers, _ = strconv.Atoi(count)
			}

			ApiBots.UsersCount = countUsers
			ApiBots.MessagesCount = CountMessages.CountMessage

			//переменная, куда будем собирать апи
			botsAPI := []BotsContentT{}

			//теперь для каждого бота соберём его сообщения
			for i := 0; i < len(bots); i++ {

				//сначала возьмём имя бота
				bot := BotsContentT{}
				bot.Name = bots[i].Name

				//сделаем запрос в базу по bot_id
				rows, err := Db.Query("select messages.user_id, users.username, users.first_name, users.last_name, messages.content, messages.c_time, messages.is_important from messages, users where users.id=messages.user_id and bot_id=?", bots[i].Id)
				if err != nil {
					fmt.Println("Не смогли получить сообщения бота", err)
				}
				defer rows.Close()

				//соберём сообщения этого бота в апишку
				for rows.Next() {
					m := MessagesBotT{}
					err := rows.Scan(&m.UserId, &m.Username, &m.FirstName, &m.LastName, &m.Content, &m.DateTime, &m.IsImportant)
					if err != nil {
						fmt.Println(err)
						continue
					}

					//если юзернейм пустой, то берём данные из другого поля и дублируем их в юзернейм
					if m.Username == "" {
						if m.FirstName != "" {
							m.Username = m.FirstName
						} else if m.LastName != "" {
							m.Username = m.LastName
						} else {
							m.Username = strconv.Itoa(m.UserId)
						}
					}

					bot.Messages = append(bot.Messages, m)
				}

				botsAPI = append(botsAPI, bot)
			}
			ApiBots.BotsContents = botsAPI

			//закодируем в json данные, чтобы выдавать в апи
			JsonBotsAPI, _ = json.Marshal(ApiBots)

			//разрешим подключаться из браузера
			w.Header().Set("Access-Control-Allow-Origin", "*")

			//выдаём апишку
			fmt.Fprintf(w, string(JsonBotsAPI))

		}

	})
	http.ListenAndServe(":80", nil)

}
