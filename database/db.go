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

	database.execQuery("PRAGMA foreign_keys = ON")

	database.execQuery("CREATE TABLE IF NOT EXISTS" +
		" global_vars(name TEXT PRIMARY KEY" +
		",integer_value INTEGER" +
		",string_value STRING);")

	database.execQuery("CREATE TABLE IF NOT EXISTS" +
		" users(id INTEGER NOT NULL PRIMARY KEY" +
		",chat_id INTEGER UNIQUE NOT NULL" +
		",score INTEGER NOT NULL" +
		",name STRING NOT NULL" +
		")")

	database.execQuery("CREATE UNIQUE INDEX IF NOT EXISTS" +
		" chat_id_index ON users(chat_id)")

	database.execQuery("CREATE TABLE IF NOT EXISTS" +
		" prohibited_words(id INTEGER NOT NULL PRIMARY KEY" +
		",word STRING NOT NULL" +
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

func (database *Database) GetUserId(chatId int64, name string) (userId int64) {
	database.execQuery(fmt.Sprintf("INSERT OR IGNORE INTO users(chat_id, score, name) "+
		"VALUES (%d, 0, '%s')", chatId, name))

	rows, err := database.conn.Query(fmt.Sprintf("SELECT id, name FROM users WHERE chat_id=%d", chatId))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer rows.Close()

	var oldName string

	if rows.Next() {
		err := rows.Scan(&userId, &oldName)
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

	if name != oldName {
		database.execQuery(fmt.Sprintf("UPDATE OR ROLLBACK users SET name='%s' WHERE id=%d", name, userId))
	}

	return
}

func (database *Database) GetUserChatId(userId int64) (chatId int64) {
	rows, err := database.conn.Query(fmt.Sprintf("SELECT chat_id FROM users WHERE id=%d", userId))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&chatId)
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

func (database *Database) AddProhibitedWord(word string) {
	database.execQuery(fmt.Sprintf("INSERT INTO prohibited_words (word) VALUES ('%s')", sanitizeString(word)))
}

func (database *Database) RemoveProhibitedWord(word string) {
	database.execQuery(fmt.Sprintf("DELETE FROM prohibited_words WHERE word='%s'", sanitizeString(word)))
}

func (database *Database) GetProhibitedWords() (words []string) {
	rows, err := database.conn.Query(fmt.Sprintf("SELECT word FROM prohibited_words"))
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

func (database *Database) GetUsersList() (ids []int64, names []string, scores []int) {
	rows, err := database.conn.Query(fmt.Sprintf("SELECT id, name, score FROM users"))
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

func (database *Database) GetUserScore(userId int64) (score int) {
	rows, err := database.conn.Query(fmt.Sprintf("SELECT score FROM users WHERE id=%d", userId))
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
		log.Fatal("No question found")
	}

	return
}

func (database *Database) AddUserScore(userId int64, addedScore int64) {
	database.execQuery(fmt.Sprintf("UPDATE OR ROLLBACK users SET score=score+%d WHERE id=%d", addedScore, userId))
}
