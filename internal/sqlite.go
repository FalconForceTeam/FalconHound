package internal

import (
	"database/sql"
	"log"
	"os"

	// _ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

func CheckAndCreateDB(dbPath string) error {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Println("Cache database not located at", dbPath, "not found, initializing...")
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		// Execute a simple SQL statement to force the creation of the database file
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS userCache (objectid TEXT PRIMARY KEY, name TEXT)")
		if err != nil {
			return err
		}
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS computerCache (objectid TEXT PRIMARY KEY, name TEXT)")
		if err != nil {
			return err
		}
	} else if err != nil {
		log.Println("Cache database located at", dbPath, "but could not be opened")
		return err
	}
	log.Println("Cache database located at", dbPath)
	return nil
}

func OpenDB(dbPath string) (*sql.DB, error) {
	return sql.Open("sqlite", dbPath)
}

func CloseDB(db *sql.DB) error {
	return db.Close()
}

func GetCachedUserByName(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT objectid FROM userCache WHERE name = ?", key).Scan(&value)
	if err != nil {
		//log.Println("Error getting cache for", key, ":", err)
		return "error", err
	}
	return value, nil
}

func GetCachedComputerByName(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT objectid FROM computerCache WHERE name = ?", key).Scan(&value)
	if err != nil {
		//log.Println("Error getting cache for", key, ":", err)
		return "error", err
	}
	return value, nil
}

func SetUserCache(db *sql.DB, key string, value string) error {
	log.Printf("Adding user %s to the cache with objectid %s", value, key)
	_, err := db.Exec("INSERT INTO userCache (objectid, name) VALUES (?, ?)", key, value)
	if err != nil {
		return err
	}
	return nil
}

func SetComputerCache(db *sql.DB, key string, value string) error {
	log.Printf("Adding computer %s to the cache with objectid %s", value, key)
	_, err := db.Exec("INSERT INTO computerCache (objectid, name) VALUES (?, ?)", key, value)
	if err != nil {
		return err
	}
	return nil
}

func GetCachedComputerByNames(db *sql.DB, keys []string) (map[string]string, error) {
	results := make(map[string]string)
	for _, key := range keys {
		value, err := GetCachedComputerByName(db, key)
		if err != nil {
			//return nil, err
			continue
		}
		results[key] = value
	}
	return results, nil
}

func DeleteCache(db *sql.DB, key string) error {
	_, err := db.Exec("DELETE FROM cache WHERE key = ?", key)
	if err != nil {
		return err
	}
	return nil
}

func ClearUserCache(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM userCache")
	if err != nil {
		return err
	}
	return nil
}

func GetCacheCount(db *sql.DB) (int, int, error) {
	var usercount int
	var computercount int
	err := db.QueryRow("SELECT COUNT(*) FROM userCache").Scan(&usercount)
	if err != nil {
		return 0, 0, err
	}
	cerr := db.QueryRow("SELECT COUNT(*) FROM computerCache").Scan(&computercount)
	if err != nil {
		return 0, 0, cerr
	}
	return usercount, computercount, nil
}
