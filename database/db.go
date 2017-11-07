package database

import (
	"bytes"
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

	database.execQuery("CREATE TABLE IF NOT EXISTS" +
		" prohibited_words(id INTEGER NOT NULL PRIMARY KEY" +
		",chat_id INTEGER NOT NULL" +
		",word STRING NOT NULL" +
		",removed INTEGER" +
		",UNIQUE(chat_id, word)" +
		")")

	database.execQuery("CREATE TABLE IF NOT EXISTS" +
		" used_words(id INTEGER NOT NULL PRIMARY KEY" +
		",user_id INTEGER NOT NULL" +
		",chat_id INTEGER NOT NULL" +
		",word_id INTEGER NOT NULL" +
		",revoked INTEGER" +
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

	var safeVersion string

	if rows.Next() {
		err := rows.Scan(&safeVersion)
		if err != nil {
			log.Fatal(err.Error())
		}
		version = strings.Replace(safeVersion, "_", ".", -1)
	} else {
		// that means it's a new clean database
		version = latestVersion
	}

	return
}

func (database *Database) SetDatabaseVersion(version string) {
	database.execQuery("DELETE FROM global_vars WHERE name='version'")

	// because sqlite eff up with some stirngs that represent numbers with dots
	safeVersion := strings.Replace(sanitizeString(version), ".", "_", -1)
	database.execQuery(fmt.Sprintf("INSERT INTO global_vars (name, string_value) VALUES ('version', '%s')", safeVersion))
}

func (database *Database) AddProhibitedWord(chatId int64,word string) {
	// insert a new word if it doesn't exist
	database.execQuery(fmt.Sprintf("INSERT OR IGNORE INTO prohibited_words (chat_id, word) VALUES (%d, '%s')",
		chatId,
		sanitizeString(word),
	))

	// mark word not removed if have been presented already
	database.execQuery(fmt.Sprintf("UPDATE prohibited_words SET removed=NULL WHERE chat_id=%d and word='%s'",
		chatId,
		sanitizeString(word),
	))
}

func (database *Database) RemoveProhibitedWord(chatId int64, word string) {
	// mark a word as removed
	database.execQuery(fmt.Sprintf("UPDATE prohibited_words SET removed=1 WHERE chat_id=%d and word='%s'",
		chatId,
		sanitizeString(word),
	))
}

func (database *Database) GetProhibitedWords(chatId int64) (words []string) {
	rows, err := database.conn.Query(fmt.Sprintf("SELECT word FROM prohibited_words WHERE chat_id=%d AND removed IS NULL ORDER BY word ASC",
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
	rows, err := database.conn.Query(fmt.Sprintf("SELECT messenger_id, name, score FROM users WHERE chat_id=%d ORDER BY score DESC",
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
		"UPDATE users SET name='%s' WHERE messenger_id=%d",
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

func (database *Database) getWordIds(groupId int64, words []string) (ids []int64) {
	sanitizedWords := []string{}
	for _, word := range words {
		sanitizedWords = append(sanitizedWords, sanitizeString(word))
	}

	wordsSet := strings.Join(sanitizedWords, "','")
	rows, err := database.conn.Query(fmt.Sprintf("SELECT id, word FROM prohibited_words WHERE chat_id=%d AND word IN ('%s')",
		groupId,
		wordsSet,
	))

	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	wordsMap := map[string]int64{}

	for rows.Next() {
		var id int64
		var word string
		err := rows.Scan(&id, &word)
		if err != nil {
			log.Fatal(err.Error())
		}
		wordsMap[word] = id
	}

	for _, word := range words {
		// can crash
		ids = append(ids, wordsMap[word])
	}

	return
}

func (database *Database) AddWordsUsage(chatId int64, messengerUserId int64, words []string) {
	database.execQuery(fmt.Sprintf("UPDATE OR ROLLBACK users SET score=score+%d WHERE messenger_id=%d AND chat_id=%d",
		len(words),
		messengerUserId,
		chatId,
	))

	var buffer bytes.Buffer

	buffer.WriteString("INSERT INTO used_words (chat_id, user_id, word_id) VALUES ")

	wordIds := database.getWordIds(chatId, words)

	isFirst := true
	for _, wordId := range wordIds {
		if !isFirst {
			buffer.WriteString(",")
		}

		buffer.WriteString(fmt.Sprintf("(%d,%d,%d)", chatId, messengerUserId, wordId))

		isFirst = false
	}

	database.execQuery(buffer.String())
}

func (database *Database) RevokeLastUsedWords(chatId int64, wordsCount int) (words []string) {
	rows, err := database.conn.Query(fmt.Sprintf("SELECT u.id, p.word, u.user_id, IFNULL(u.revoked, 0) FROM used_words as u, prohibited_words as p WHERE u.chat_id=%d AND u.word_id=p.id ORDER BY u.id DESC LIMIT %d",
		chatId,
		wordsCount,
	))
	if err != nil {
		log.Fatal(err.Error())
	}

	var userId = int64(-1)
	revokedIds := []int64{}

	for rows.Next() {
		var word string
		var usedWordId int64
		var isRevoked int
		var lastUserId int64
		err := rows.Scan(&usedWordId, &word, &lastUserId, &isRevoked)
		if err != nil {
			log.Fatal(err.Error())
		}

		if userId != int64(-1) && userId != lastUserId {
			break
		}
		userId = lastUserId

		if isRevoked == 0 {
			words = append(words, word)
			revokedIds = append(revokedIds, usedWordId)
		}
	}

	rows.Close()

	if len(revokedIds) > 0 {
		database.execQuery(fmt.Sprintf("UPDATE OR ROLLBACK users SET score=score-%d WHERE messenger_id=%d AND chat_id=%d",
			len(revokedIds),
			userId,
			chatId,
		))

		ids_list := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(revokedIds)), ","), "[]")
		database.execQuery(fmt.Sprintf("UPDATE OR ROLLBACK used_words SET revoked=1 WHERE id in (%s)", ids_list))
	}

	return
}
