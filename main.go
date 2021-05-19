package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var (
	targetDir string
	dataDir   string
	database  string
	username  string
	password  string
	host      string
	workers   int
	wg        sync.WaitGroup
)

func init() {
	flag.StringVar(&targetDir, "target-dir", "./", "Directory with backuped files")
	flag.StringVar(&dataDir, "data-dir", "/var/lib/mysql/", "Directory with data")
	flag.StringVar(&database, "database", "", "Database")
	flag.StringVar(&username, "username", "root", "User")
	flag.StringVar(&password, "password", "", "Password")
	flag.IntVar(&workers, "workers", 1, "Number of workers")
	flag.StringVar(&host, "host", "tcp(127.0.0.1:3306)", "Host")
	flag.Parse()
}

func main() {
	bufferTables := make(chan string)

	files, err := filterDirsGlob(targetDir, "*.frm")

	if err != nil {
		panic(err)
	}

	for w := 1; w <= workers; w++ {
		go worker(bufferTables)
	}

	for i := range files {
		tableName := filepath.Base(fileNameWithoutExtension(files[i]))
		wg.Add(1)
		bufferTables <- tableName
	}

	wg.Wait()
}

func worker(jobs chan string) {
	for {
		select {
		case tableName := <-jobs:
			fmt.Println(tableName)

			mysqlGroup, err := user.LookupGroup("mysql")
			if err != nil {
				panic(err)
			}
			mysqlGid, _ := strconv.Atoi(mysqlGroup.Gid)

			mysqlUser, err := user.Lookup("mysql")
			if err != nil {
				panic(err)
			}
			mysqlUid, _ := strconv.Atoi(mysqlUser.Uid)

			db, err := sql.Open("mysql", username+":"+password+"@"+host+"/"+database+"?multiStatements=true")
			db.SetMaxIdleConns(0)
			db.SetMaxOpenConns(100)
			db.SetConnMaxLifetime(10 * time.Minute)

			if err != nil {
				panic(err)
			}

			tableFiles, _ := filterDirsGlob(targetDir, tableName+".*")

			// copy files
			databaseDataPath := dataDir + database + "/"
			if isNotEqualsFiles(tableFiles, databaseDataPath) {
				if isInnoDB(tableFiles) {
					fmt.Println("discard", tableName)
					execQuery(db, "SET FOREIGN_KEY_CHECKS=0; ALTER TABLE "+database+"."+tableName+" DISCARD TABLESPACE;")
				}

				for j := range tableFiles {
					srcFile := tableFiles[j]
					destFile := databaseDataPath + filepath.Base(tableFiles[j])

					copy(srcFile, destFile)
					if err := os.Chmod(destFile, 0660); err != nil {
						panic(err)
					}
					if err := os.Chown(destFile, mysqlUid, mysqlGid); err != nil {
						panic(err)
					}
				}

				// TODO: check isl

				if isInnoDB(tableFiles) {
					fmt.Println("import", tableName)
					execQuery(db, "ALTER TABLE "+database+"."+tableName+" IMPORT TABLESPACE; SET FOREIGN_KEY_CHECKS=1;")
				}
			} else {
				fmt.Println("skip", tableName)
			}


			db.Close()

			wg.Done()
		}
	}
}

func isInnoDB(files []string) bool {
	for i := range files {
		if filepath.Ext(files[i]) == ".ibd" {
			return true
		}
	}
	return false
}

func isNotEqualsFiles(files []string, databaseDataPath string) bool {
	for j := range files {
		srcFile := files[j]
		destFile := databaseDataPath + filepath.Base(files[j])

		if !isFileEquals(srcFile, destFile) {
			return true
		}
	}

	return false
}

func execQuery(db *sql.DB, query string) {
	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}
