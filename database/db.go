package database

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type Database struct {
	// connection
	conn *sql.DB
}

func sanitizeString(input string) (result string) {
	result = input
	result = strings.Replace(result, "'", "''", -1)
	return
}

func (database *Database) execQuery(query string) {
	_, err := database.conn.Exec(query)

	if err != nil {
		log.Fatal(err.Error())
	}
}

func (database *Database) Connect(fileName string) error {
	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}

	database.conn = db

	// database.execQuery("PRAGMA foreign_keys = ON")

	database.execQuery("CREATE TABLE IF NOT EXISTS" +
		" global_vars(name TEXT NOT NULL PRIMARY KEY" +
		",integer_value INTEGER" +
		",string_value STRING" +
		");")

	// the same user in a different chat treats as a different user
	database.execQuery("CREATE TABLE IF NOT EXISTS" +
		" users(messenger_id INTEGER NOT NULL" +
		",chat_id INTEGER NOT NULL" +
		",score INTEGER NOT NULL" +
		",name STRING NOT NULL" +
		",PRIMARY KEY (messenger_id, chat_id)" +
		")")

	// database.execQuery("CREATE UNIQUE INDEX IF NOT EXISTS" +
	// 	" chat_id_index ON users(chat_id)")

	database.execQuery("CREATE TABLE IF NOT EXISTS" +
		" prohibited_words(id INTEGER NOT NULL PRIMARY KEY" +
		",chat_id INTEGER NOT NULL" +
		",word STRING NOT NULL" +
		",UNIQUE(chat_id, word)" +
		")")

	return nil
}

func (database *Database) Disconnect() {
	database.conn.Close()
	database.conn = nil
}

func (database *Database) IsConnectionOpened() bool {
	return database.conn != nil
}

func (database *Database) createUniqueRecord(table string, values string) int64 {
	var err error
	if len(values) == 0 {
		_, err = database.conn.Exec(fmt.Sprintf("INSERT INTO %s DEFAULT VALUES ", table))
	} else {
		_, err = database.conn.Exec(fmt.Sprintf("INSERT INTO %s VALUES (%s)", table, values))
	}

	if err != nil {
		log.Fatal(err.Error())
		return -1
	}

	rows, err := database.conn.Query(fmt.Sprintf("SELECT id FROM %s ORDER BY id DESC LIMIT 1", table))

	if err != nil {
		log.Fatal(err.Error())
		return -1
	}
	defer rows.Close()

	if rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err.Error())
			return -1
		}

		return id
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal("No record created")
	return -1
}

func (database *Database) GetDatabaseVersion() (version string) {
	rows, err := database.conn.Query("SELECT string_value FROM global_vars WHERE name='version'")

	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&version)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		// that means it's a new clean database
		version = latestVersion
	}

	return
}

func (database *Database) SetDatabaseVersion(version string) {
	database.execQuery("DELETE FROM global_vars WHERE name='version'")
	database.execQuery(fmt.Sprintf("INSERT INTO global_vars (name, string_value) VALUES ('version', '%s')", sanitizeString(version)))
}

func (database *Database) AddProhibitedWord(chatId int64,word string) {
	database.execQuery(fmt.Sprintf("INSERT OR IGNORE INTO prohibited_words (chat_id, word) VALUES (%d, '%s')",
		chatId,
		sanitizeString(word),
	))
}

func (database *Database) RemoveProhibitedWord(chatId int64, word string) {
	database.execQuery(fmt.Sprintf("DELETE FROM prohibited_words WHERE chat_id=%d AND word='%s'",
		chatId,
		sanitizeString(word),
	))
}

func (database *Database) GetProhibitedWords(chatId int64) (words []string) {
	rows, err := database.conn.Query(fmt.Sprintf("SELECT word FROM prohibited_words WHERE chat_id=%d ORDER BY word ASC",
		chatId,
	))

	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var word string
		err := rows.Scan(&word)
		if err != nil {
			log.Fatal(err.Error())
		}
		words = append(words, word)
	}

	return
}

func (database *Database) GetUsersList(chatId int64) (ids []int64, names []string, scores []int) {
	rows, err := database.conn.Query(fmt.Sprintf("SELECT messenger_id, name, score FROM users WHERE chat_id=%d",
		chatId,
	))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		var score int
		err := rows.Scan(&id, &name, &score)
		if err != nil {
			log.Fatal(err.Error())
		}

		ids = append(ids, id)
		names = append(names, name)
		scores = append(scores, score)
	}

	return
}

func (database *Database) UpdateUser(chatId int64, messengerUserId int64, name string) {
	sanitizedName := sanitizeString(name)

	database.execQuery(fmt.Sprintf(
		"INSERT OR IGNORE INTO users(messenger_id, chat_id, name, score) VALUES (%d, %d, '%s', 0);" +
		"UPDATE users SET name='%s' where messenger_id=%d",
		messengerUserId,
		chatId,
		sanitizedName,
		sanitizedName,
		messengerUserId,
	))
}

func (database *Database) GetUserScore(chatId int64, messengerUserId int64) (score int) {
	rows, err := database.conn.Query(fmt.Sprintf("SELECT score FROM users WHERE messenger_id=%d AND chat_id=%d",
		messengerUserId,
		chatId,
	))

	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&score)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal("No user found")
	}

	return
}

func (database *Database) AddUserScore(chatId int64, messengerUserId int64, addedScore int) {
	database.execQuery(fmt.Sprintf("UPDATE OR ROLLBACK users SET score=score+%d WHERE messenger_id=%d AND chat_id=%d",
		addedScore,
		messengerUserId,
		chatId,
	))
}
