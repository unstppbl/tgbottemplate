package main

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

const (
	users = `CREATE TABLE IF NOT EXISTS users (
		id INT UNIQUE NOT NULL,
		first_name TEXT,
		last_name TEXT,
		username TEXT,
		tz TEXT,
		location TEXT,
		language_code TEXT,
		meta TEXT
	);`

	groups = `CREATE TABLE IF NOT EXISTS groups (
		id INT UNIQUE NOT NULL,
		title TEXT,
		tz TEXT,
		location TEXT,
		meta TEXT
	);`
)

func (app *appStruct) initDB() (err error) {
	app.db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	if err := app.db.Ping(); err != nil {
		return err
	}
	if _, err := app.db.Exec(users); err != nil {
		return err
	}
	if _, err := app.db.Exec(groups); err != nil {
		return err
	}
	return nil
}

func (app *appStruct) dbSaveUser(user *userStruct) (err error) {
	if _, err := app.db.Exec(`INSERT INTO users (id, first_name, last_name, username, language_code, tz, location, meta) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (id) DO UPDATE
		SET tz = EXCLUDED.tz, first_name = EXCLUDED.first_name, last_name = ESCLUDED.last_name, username = EXCLUDED.username, language_code = EXCLUDED.language_code
		location = EXCLUDED.location, meta = EXCLUDED.meta`, user.ID, user.FirstName, user.LastName, user.UserName, user.LanguageCode, user.Tz, user.Location, user.meta); err != nil {
		return err
	}
	return nil
}

func (app *appStruct) dbSaveGroup(group *groupStruct) (err error) {
	if _, err := app.db.Exec(`INSERT INTO groups (id, title, tz, location, meta) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO UPDATE
		SET tz = EXCLUDED.tz, title = EXCLUDED.title, location = EXCLUDED.location, meta = EXCLUDED.meta`, group.ID, group.Title, group.Tz, group.Location, group.meta); err != nil {
		return err
	}
	return nil
}

func (app *appStruct) dbUpdateUser(user *userStruct) (err error) {
	if _, err := app.db.Exec(`UPDATE users SET 
			first_name = $1,
			last_name = $2,
			username = $3,
			tz = $4,
			language_code = $5,
			meta = $7
			WHERE id = $6`,
		user.FirstName, user.LastName, user.UserName, user.Tz, user.LanguageCode, user.meta, user.ID); err != nil {
		return err
	}
	return nil
}

func (app *appStruct) dbUpdateUserTz(id int, tz string) (err error) {
	if _, err := app.db.Exec(`UPDATE users SET tz = $1 WHERE id = $2`, tz, id); err != nil {
		return err
	}
	return nil
}

func (app *appStruct) dbGetUser(id int) (user *userStruct, err error) {
	user = &userStruct{
		ID: id,
	}
	if err := app.db.QueryRow(`SELECT first_name, last_name, username, tz, language_code FROM users WHERE id = $1`, id).Scan(&user.FirstName, &user.LastName, &user.UserName, &user.Tz, &user.LanguageCode); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (app *appStruct) dbGetGroup(id int64) (group *groupStruct, err error) {
	group = &groupStruct{
		ID: id,
	}
	if err := app.db.QueryRow(`SELECT title, tz, location FROM groups WHERE id = $1`, id).Scan(&group.Title, &group.Tz, &group.Location); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return group, nil
}
