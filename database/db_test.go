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

func TestUpdateUser(t *testing.T) {
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

	var userId1 int64 = 1234
	var userId2 int64 = 4321

	db.UpdateUser(chatId1, userId1, "test1")
	db.UpdateUser(chatId2, userId2, "test2")

	db.UpdateUser(chatId1, userId2, "test3")
	db.UpdateUser(chatId2, userId1, "test4")

	db.UpdateUser(chatId2, userId1, "test5")

	{
		ids, names, scores := db.GetUsersList(chatId1)

		for idx, id := range ids {
			if id == userId1 {
				assert.Equal("test5", names[idx])
				assert.Equal(0, scores[idx])
			} else if id == userId2 {
				assert.Equal("test3", names[idx])
				assert.Equal(0, scores[idx])
			}
		}
	}

	{
		ids, names, scores := db.GetUsersList(chatId2)

		for idx, id := range ids {
			if id == userId1 {
				assert.Equal("test5", names[idx])
				assert.Equal(0, scores[idx])
			} else if id == userId2 {
				assert.Equal("test3", names[idx])
				assert.Equal(0, scores[idx])
			}
		}
	}
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

	var chatId1 int64 = 321
	var chatId2 int64 = 123

	prohibitedWord1 := "testWord1"
	prohibitedWord2 := "testWord2"

	{
		db.AddProhibitedWord(chatId1, prohibitedWord1)
		assert.Equal(1, len(db.GetProhibitedWords(chatId1)))
		db.AddProhibitedWord(chatId1, prohibitedWord1)
		assert.Equal(1, len(db.GetProhibitedWords(chatId1)))
		assert.Equal(prohibitedWord1, db.GetProhibitedWords(chatId1)[0])
		db.RemoveProhibitedWord(chatId1, prohibitedWord1)
		assert.Equal(0, len(db.GetProhibitedWords(chatId1)))
	}

	{
		db.AddProhibitedWord(chatId2, prohibitedWord2)
		assert.Equal(1, len(db.GetProhibitedWords(chatId2)))
		assert.Equal(prohibitedWord2, db.GetProhibitedWords(chatId2)[0])
		db.AddProhibitedWord(chatId2, prohibitedWord1)
		assert.Equal(2, len(db.GetProhibitedWords(chatId2)))
		db.RemoveProhibitedWord(chatId2, prohibitedWord2)
		assert.Equal(1, len(db.GetProhibitedWords(chatId2)))
	}
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

	var chatId int64 = 123

	var userId1 int64 = 1234
	var userId2 int64 = 4321

	db.UpdateUser(chatId, userId1, "testName1")
	db.UpdateUser(chatId, userId2, "testName2")

	assert.Equal(0, db.GetUserScore(chatId, userId1))
	assert.Equal(0, db.GetUserScore(chatId, userId2))

	db.AddUserScore(chatId, userId1, 1)
	db.AddUserScore(chatId, userId2, 5)

	assert.Equal(1, db.GetUserScore(chatId, userId1))
	assert.Equal(5, db.GetUserScore(chatId, userId2))

	ids, names, score := db.GetUsersList(chatId)

	assert.Equal(2, len(ids))
	if len(ids) > 1 {
		// sorted by score DESC
		assert.Equal(5, score[0])
		assert.Equal("testName2", names[0])
		assert.Equal(1, score[1])
		assert.Equal("testName1", names[1])
	}
}
