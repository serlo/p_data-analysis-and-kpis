package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"time"
)

type table interface {
	name() string
	create() error
	load(maxID int, rowLimit int) (int, error)
	save() error
}

func openAthene2DB(config *mysqlConfig) (*sql.DB, error) {
	dbConnect := fmt.Sprintf("%s:%s@tcp(%s)/serlo?parseTime=true", config.User, config.Password, config.URL)

	log.Logger.Info().Msgf("open database [%s]", config.URL)
	db, err := sql.Open("mysql", dbConnect)

	if err != nil {
		return nil, fmt.Errorf("open database [%s] error [%s]", config.URL, err.Error())
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("open database [%s] ping error [%s]", config.URL, err.Error())
	}

	db.SetConnMaxLifetime(time.Second * 600)

	log.Logger.Info().Msgf("open database [%s] successful", config.URL)

	return db, nil
}

func databaseDoesNotExist(err error) bool {
	match, _ := regexp.MatchString(".*database.*does not exist", err.Error())
	return match
}

func openKPIDatabase(config *postgresConfig) (*sql.DB, error) {
	dbConnect := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
	
	log.Logger.Info().Msgf("open database [%s] [%s]", config.Host, config.DBName)
	db, err := sql.Open("postgres", dbConnect)
	if err != nil {
		if databaseDoesNotExist(err) {
			db, err = createKPIDatabase(config)
			if err != nil {
				return nil, err
			}
			return db, nil
		}
		return nil, fmt.Errorf("open database [%s] [%s] error [%s]", config.Host, config.DBName, err.Error())
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		if databaseDoesNotExist(err) {
			db, err = createKPIDatabase(config)
			if err != nil {
				return nil, err
			}
			return db, nil
		}
		return nil, fmt.Errorf("open database [%s] [%s] ping error [%s]", config.Host, config.DBName, err.Error())
	}

	log.Logger.Info().Msgf("open database [%s] [%s] successful", config.Host, config.DBName)

	db.SetConnMaxLifetime(time.Second * 3600)

	return db, nil
}

func createKPIDatabase(config *postgresConfig) (*sql.DB, error) {
	log.Logger.Info().Msgf("create database [%s] [%s]", config.Host, config.DBName)
	serverConnect := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.SSLMode)
	log.Logger.Info().Msgf("open server connection [%s]", config.Host)
	server, err := sql.Open("postgres", serverConnect)
	if err != nil {
		return nil, fmt.Errorf("create database cannot connect to server [%s] error [%s]", config.Host, err.Error())
	}

	defer server.Close()

	_, err = server.Exec(fmt.Sprintf("CREATE DATABASE %s", config.DBName))
	if err != nil {
		return nil, fmt.Errorf("cannot create kpi database [%s] on postgres server [%s] error [%s]", config.DBName, config.Host, err.Error())
	}
	server.Close()

	//open kpi database
	kpiDBConnect := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
	config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	log.Logger.Info().Msgf("open database [%s] [%s]", config.DBName, config.Host)
	db, err := sql.Open("postgres", kpiDBConnect)
	return db, nil
}



func isTableCreated(targetDB *sql.DB, name string) (bool, error) {
	stmt := fmt.Sprintf(`SELECT EXISTS (
		SELECT 1
		FROM   information_schema.tables 
		WHERE  table_schema = 'public'
		AND    table_name = '%s'
		)`, name)
	row := targetDB.QueryRow(stmt)
	result := false
	err := row.Scan(&result)
	if err != nil {
		return false, err
	}

	return result, nil
}

func checkForUpdates(sourceDB *sql.DB, targetDB *sql.DB, name string) (bool, int, error) {
	sourceMaxID, err := getSourceMaxID(sourceDB, name)
	if err != nil {
		return false, 0, err
	}

	targetMaxID, err := getTargetMaxID(targetDB, name)
	if err != nil {
		return false, 0, err
	}

	if sourceMaxID == targetMaxID {
		log.Logger.Info().Str("table", name).Int("sourceMaxID", sourceMaxID).Int("targetMaxID", targetMaxID).Msg("up-to-date")
		return false, sourceMaxID, nil
	}

	if targetMaxID > sourceMaxID {
		return false, 0, fmt.Errorf("inconsistence in table [%s] targetMaxID [%d] greater than sourceMaxID [%d]", name, targetMaxID, sourceMaxID)
	}
	log.Logger.Info().Str("table", name).Int("sourceMaxID", sourceMaxID).Int("targetMaxID", targetMaxID).Msg("import required")
	return true, targetMaxID, nil
}

func getSourceMaxID(sourceDB *sql.DB, name string) (int, error) {
	rows, err := sourceDB.Query(fmt.Sprintf("SELECT COALESCE(max(id), 0) FROM serlo.%s", name))
	if err != nil {
		return 0, fmt.Errorf("cannot get max id from source db table %s [%s]", name, err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		id := 0
		err = rows.Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("cannot get max id from targetdb table %s [%s]", name, err.Error())
		}
		return int(id), nil
	}
	return 0, fmt.Errorf("did not get the max(id) from table %s [%s]", name, err.Error())
}

func getTargetMaxID(targetDB *sql.DB, name string) (int, error) {
	rows, err := targetDB.Query(fmt.Sprintf("SELECT COALESCE(max(id), 0) FROM public.%s", name))
	if err != nil {
		return 0, fmt.Errorf("cannot get max id from table %s [%s]", name, err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		id := 0
		err = rows.Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("cannot get max id from table %s [%s]", name, err.Error())
		}
		return int(id), nil
	}
	return 0, fmt.Errorf("did not get the max(id) from table %s [%s]", name, err.Error())
}

func createTable(db *sql.DB, name string, statements []string) error {
	created, err := isTableCreated(db, name)
	if err != nil {
		return err
	}
	if !created {
		log.Logger.Info().Msgf("'create %s\n", name)
		for _, stmt := range statements {
			_, err := db.Exec(stmt)
			if err != nil {
				return fmt.Errorf("create table %s failed [%s]", name, err.Error())
			}
		}
	}

	return nil
}