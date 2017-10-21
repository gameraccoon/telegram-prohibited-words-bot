package database

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	testDbPath = "./testDb.db"
)

func dropDatabase(fileName string) {
	os.Remove(fileName)
}

func clearDb() {
	dropDatabase(testDbPath)
}

func connectDb(t *testing.T) *Database {
	assert := require.New(t)
	db := &Database{}

	err := db.Connect(testDbPath)
	if err != nil {
		assert.Fail("Problem with creation db connection:" + err.Error())
		return nil
	}
	return db
}

func createDbAndConnect(t *testing.T) *Database {
	clearDb()
	return connectDb(t)
}

func TestConnection(t *testing.T) {
	assert := require.New(t)
	dropDatabase(testDbPath)

	db := &Database{}

	assert.False(db.IsConnectionOpened())

	err := db.Connect(testDbPath)
	defer dropDatabase(testDbPath)
	if err != nil {
		assert.Fail("Problem with creation db connection:" + err.Error())
		return
	}

	assert.True(db.IsConnectionOpened())

	db.Disconnect()

	assert.False(db.IsConnectionOpened())
}

func TestDatabaseVersion(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}

	{
		version := db.GetDatabaseVersion()
		assert.Equal(latestVersion, version)
	}

	db.SetDatabaseVersion("1.2")

	{
		version := db.GetDatabaseVersion()
		assert.Equal("1.2", version)
	}

	db.SetDatabaseVersion("1.4")
	db.Disconnect()

	{
		db = connectDb(t)
		version := db.GetDatabaseVersion()
		assert.Equal("1.4", version)
		db.Disconnect()
	}
}

func TestGetUserId(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	var chatId1 int64 = 321
	var chatId2 int64 = 123

	name := "testName"

	id1 := db.GetUserId(chatId1, name)
	id2 := db.GetUserId(chatId1, name)
	id3 := db.GetUserId(chatId2, name)

	assert.Equal(id1, id2)
	assert.NotEqual(id1, id3)

	assert.Equal(chatId1, db.GetUserChatId(id1))
	assert.Equal(chatId2, db.GetUserChatId(id3))
}

func TestSanitizeString(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	testText := "text'test''test\"test\\"

	db.SetDatabaseVersion(testText)
	assert.Equal(testText, db.GetDatabaseVersion())
}

func TestAddAndRemoveProhibitedWord(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	prohibitedWord := "testWord"

	db.AddProhibitedWord(prohibitedWord)
	db.AddProhibitedWord(prohibitedWord)
	assert.Equal(prohibitedWord, db.GetProhibitedWords()[0])
	db.RemoveProhibitedWord(prohibitedWord)
	assert.Equal(0, len(db.GetProhibitedWords()))
}

func TestScoringUsers(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	var chatId1 int64 = 123
	var chatId2 int64 = 321

	userId1 := db.GetUserId(chatId1, "testName1")
	userId2 := db.GetUserId(chatId2, "testName2")

	assert.Equal(0, db.GetUserScore(userId1))
	assert.Equal(0, db.GetUserScore(userId2))

	db.AddUserScore(userId1, 1)
	db.AddUserScore(userId2, 5)

	assert.Equal(1, db.GetUserScore(userId1))
	assert.Equal(5, db.GetUserScore(userId2))

	ids, names, score := db.GetUsersList()

	for idx, id := range ids {
		if id == userId1 {
			assert.Equal(1, score[idx])
			assert.Equal("testName1", names[idx])
		} else if id == userId2 {
			assert.Equal(5, score[idx])
			assert.Equal("testName2", names[idx])
		}
	}
}
