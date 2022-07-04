package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := openDB(dsn)
	if err != nil {
		return nil, err
	}

	s := new(DBStorage)
	s.db = db

	return s, nil
}

func (s *DBStorage) InsertUser() (UserID, error) {
	row := s.db.QueryRow("INSERT INTO users VALUES(default) RETURNING user_id;")

	var id UserID
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *DBStorage) InsertLink(userID UserID, link string) (LinkID, error) {
	var alreadyExists bool
	linkIDs, err := s.insertLinks(userID, []string{link}, &alreadyExists)
	if err != nil {
		return 0, err
	}

	if len(linkIDs) != 1 {
		panic(fmt.Sprintf("Invalid number of linkIDs. Count=%v", len(linkIDs)))
	}

	res := linkIDs[0]
	if alreadyExists {
		return res, ErrLinkAlreadyExists
	}

	return res, nil
}

func (s *DBStorage) InsertLinks(userID UserID, links []string) ([]LinkID, error) {
	return s.insertLinks(userID, links, nil)
}

func (s *DBStorage) insertLinks(userID UserID, links []string, alreadyExists *bool) ([]LinkID, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	linkIDs, err := s.doInsertLinks(tx, links, alreadyExists)
	if err != nil {
		return nil, err
	}

	if err = s.insertUserLinks(tx, userID, linkIDs); err != nil {
		return nil, err
	}

	return linkIDs, tx.Commit()
}

func (s *DBStorage) doInsertLinks(tx *sql.Tx, links []string, alreadyExists *bool) ([]LinkID, error) {
	if len(links) == 0 {
		return nil, nil
	}

	q := `WITH ins AS(
    	INSERT INTO links ("link") VALUES ($1) 
    		ON CONFLICT("link") DO NOTHING
    	RETURNING link_id, true as is_new
	)
	SELECT * FROM ins
	UNION
	  SELECT link_id, false as is_new FROM links WHERE link=$1;`

	stmt, err := tx.Prepare(q)
	if err != nil {
		return nil, err
	}

	res := make([]LinkID, 0, len(links))
	for _, link := range links {
		row := stmt.QueryRow(link)

		var id LinkID
		var isNew bool

		if err := row.Scan(&id, &isNew); err != nil {
			return nil, err
		}

		if !isNew && alreadyExists != nil {
			*alreadyExists = true
		}

		res = append(res, id)
	}

	return res, nil
}

func (s *DBStorage) insertUserLinks(tx *sql.Tx, userID UserID, linkIDs []LinkID) error {
	if len(linkIDs) == 0 {
		return nil
	}

	q := `
	INSERT INTO user_links("user_id", "link_id") VALUES ($1, $2) 
	ON CONFLICT("user_id", "link_id") DO NOTHING`

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}

	for _, linkID := range linkIDs {
		if _, err := stmt.Exec(userID, linkID); err != nil {
			return err
		}
	}

	return nil
}

func (s *DBStorage) SelectLink(id LinkID) (string, error) {
	row := s.db.QueryRow("SELECT link FROM links WHERE link_id=$1", id)
	var link string
	if err := row.Scan(&link); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrLinkNotFound
		}
		return "", err
	}
	return link, nil
}

func (s *DBStorage) SelectUserLinks(id UserID) (map[string]LinkID, error) {
	q := `
	SELECT links.link_id, links.link FROM links 
	  INNER JOIN user_links ON links.link_id = user_links.link_id
	WHERE user_links.user_id=$1`

	rows, err := s.db.Query(q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	res := make(map[string]LinkID)
	for rows.Next() {
		var id LinkID
		var link string

		err = rows.Scan(&id, &link)
		if err != nil {
			return nil, err
		}
		res[link] = id
	}

	return res, nil
}

func (s *DBStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (s *DBStorage) Close() error {
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

	urlsTable := `CREATE TABLE IF NOT EXISTS links(
	 	link_id SERIAL NOT NULL,
	 	link TEXT NOT NULL UNIQUE,
     	PRIMARY KEY (link_id))`

	if err = createTable(urlsTable); err != nil {
		return nil, err
	}

	usersTable := `CREATE TABLE IF NOT EXISTS users(
		user_id SERIAL NOT NULL,
		PRIMARY KEY (user_id))`

	if err = createTable(usersTable); err != nil {
		return nil, err
	}

	userURLsTable := `CREATE TABLE IF NOT EXISTS user_links(
		user_id INTEGER NOT NULL,
		link_id  INTEGER NOT NULL,
		UNIQUE(user_id, link_id),
		CONSTRAINT fk_user_id
      		FOREIGN KEY(user_id) REFERENCES users(user_id)
	  		ON DELETE CASCADE,
		CONSTRAINT fk_link_id
			FOREIGN KEY(link_id) REFERENCES links(link_id)
			ON DELETE CASCADE
	)`

	if err = createTable(userURLsTable); err != nil {
		return nil, err
	}
	return db, nil
}

func cleanDB(db *sql.DB) {
	q := `
DROP TABLE likns;
DROP TABLE users;
DROP TABLE user_links;
`
	db.Exec(q)
}
