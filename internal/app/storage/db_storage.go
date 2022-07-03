package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ikashurnikov/shortener/internal/app/urlencoder"
	"time"

	_ "github.com/lib/pq"
)

type DBStorage struct {
	db *sql.DB

	baseStorage
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := openDB(dsn)
	if err != nil {
		return nil, err
	}

	s := new(DBStorage)
	s.db = db
	s.urlEncoder = urlencoder.NewZBase32Encoder()
	s.baseStorage.storageImpl = s

	return s, nil
}

func (s *DBStorage) insertLongURL(userID *UserID, url string) (uint32, error) {
	// Добавляем запись в таблицу urls
	q := `
	WITH ins AS(
    	INSERT INTO urls ("url") VALUES ($1) 
    		ON CONFLICT("url") DO NOTHING
    	RETURNING url_id
	)
	SELECT * FROM ins
	UNION
		SELECT url_id FROM urls WHERE url=$1;`

	var urlID uint32
	row := s.db.QueryRow(q, url)
	if err := row.Scan(&urlID); err != nil {
		return 0, err
	}

	// Добавляем запись в таблицу users
	q = `
 	INSERT INTO users
	SELECT 
	WHERE NOT EXISTS(SELECT user_id FROM users WHERE user_id = $1) RETURNING user_id;
`
	row = s.db.QueryRow(q, *userID)
	err := row.Scan(userID)
	// Если пользователя нет, будет добавлена новая запись и возварщена
	// инчае не будет возвращаено ничего - sql.ErrNoRows.
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	// Добавляем запись в таблицу user_urls
	q = `
	INSERT INTO user_urls("user_id", "url_id") VALUES ($1, $2) 
	ON CONFLICT("user_id", "url_id") DO NOTHING `

	if _, err := s.db.Exec(q, *userID, urlID); err != nil {
		return 0, err
	}

	return urlID, nil
}

func (s *DBStorage) selectLongURL(id uint32) (string, error) {
	row := s.db.QueryRow("SELECT url FROM urls WHERE url_id=$1", id)
	var url string
	if err := row.Scan(&url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrDecodingShortURL
		}
		return "", err
	}
	return url, nil
}

func (s *DBStorage) getUserURLs(id UserID, newURLInfo newURLInfoFunc) ([]URLInfo, error) {
	q := `
	SELECT urls.url_id, urls.url FROM urls 
	  INNER JOIN user_urls ON urls.url_id = user_urls.url_id
	WHERE user_urls.user_id=$1`

	rows, err := s.db.Query(q, id)
	if err != nil {
		return nil, err
	}

	res := make([]URLInfo, 0)
	for rows.Next() {
		var id uint32
		var url string

		err = rows.Scan(&id, &url)
		if err != nil {
			return nil, err
		}

		uinfo, err := newURLInfo(id, url)
		if err != nil {
			return nil, err
		}
		res = append(res, uinfo)
	}

	return res, nil
}

func (s *DBStorage) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return ErrStorage
	}
	return nil
}

func (s *DBStorage) close() error {
	return s.db.Close()
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	createTable := func(q string) error {
		_, err := db.Exec(q)
		if err != nil {
			db.Close()
			return err
		}
		return nil
	}

	urlsTable := `CREATE TABLE IF NOT EXISTS urls(
	 	url_id SERIAL NOT NULL,
	 	url TEXT NOT NULL,
		UNIQUE(url),
     	PRIMARY KEY (url_id))`

	if err = createTable(urlsTable); err != nil {
		return nil, err
	}

	usersTable := `CREATE TABLE IF NOT EXISTS users(
		user_id SERIAL NOT NULL,
		PRIMARY KEY (user_id))`

	if err = createTable(usersTable); err != nil {
		return nil, err
	}

	userURLsTable := `CREATE TABLE IF NOT EXISTS user_urls(
		user_id INTEGER NOT NULL,
		url_id  INTEGER NOT NULL,
		UNIQUE(user_id, url_id),
		CONSTRAINT fk_user_id
      		FOREIGN KEY(user_id) 
	  		REFERENCES users(user_id)
	  		ON DELETE CASCADE,
		CONSTRAINT fk_url_id
			FOREIGN KEY(url_id)
			REFERENCES urls(url_id)
			ON DELETE CASCADE
	)`

	if err = createTable(userURLsTable); err != nil {
		return nil, err
	}
	return db, nil
}

func cleanDB(db *sql.DB) {
	q := `
DROP TABLES urls;
DROP TABLES users;
DROP TABLES user_urls
`
	db.Exec(q)
}
