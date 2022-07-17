package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"time"

	"github.com/ikashurnikov/shortener/internal/app/model"
)

type dbRepo struct {
	db *sql.DB
}

func NewDBRepo(dsn string) (*dbRepo, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	//cleanDB(db)
	if err = initDatabase(db); err != nil {
		return nil, err
	}
	return &dbRepo{db: db}, nil
}

func (repo *dbRepo) AddUser() (model.UserID, error) {
	row := repo.db.QueryRow("INSERT INTO users VALUES(default) RETURNING user_id;")

	var id model.UserID
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (repo *dbRepo) SaveOriginalURL(userID model.UserID, origURL string) (model.LinkID, error) {
	var alreadyExists bool
	linkIDs, err := repo.saveOriginalURLs(userID, []string{origURL}, &alreadyExists)
	if err != nil {
		return 0, err
	}

	if len(linkIDs) != 1 {
		panic(fmt.Sprintf("Invalid number of linkIDs. Count=%v", len(linkIDs)))
	}

	res := linkIDs[0]
	if alreadyExists {
		return res, model.ErrLinkAlreadyExists
	}

	return res, nil
}

func (repo *dbRepo) SaveOriginalURLs(userID model.UserID, origURLs []string) ([]model.LinkID, error) {
	return repo.saveOriginalURLs(userID, origURLs, nil)
}

func (repo *dbRepo) saveOriginalURLs(userID model.UserID, origURLs []string, alreadyExists *bool) ([]model.LinkID, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	linkIDs, err := repo.doSaveOriginalURLs(tx, origURLs, alreadyExists)
	if err != nil {
		return nil, err
	}

	if err = repo.saveUserLinks(tx, userID, linkIDs); err != nil {
		return nil, err
	}

	return linkIDs, tx.Commit()
}

func (repo *dbRepo) doSaveOriginalURLs(tx *sql.Tx, origURLs []string, alreadyExists *bool) ([]model.LinkID, error) {
	if len(origURLs) == 0 {
		return nil, nil
	}

	q := `WITH ins AS(
    	INSERT INTO links ("original_url") VALUES ($1) 
    		ON CONFLICT("original_url") DO NOTHING
    	RETURNING link_id, true as is_new
	)
	SELECT * FROM ins
	UNION
	  SELECT link_id, false as is_new FROM links WHERE original_url=$1;`

	stmt, err := tx.Prepare(q)
	if err != nil {
		return nil, err
	}

	res := make([]model.LinkID, 0, len(origURLs))
	for _, origURL := range origURLs {
		row := stmt.QueryRow(origURL)

		var id model.LinkID
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

func (repo *dbRepo) saveUserLinks(tx *sql.Tx, userID model.UserID, linkIDs []model.LinkID) error {
	if len(linkIDs) == 0 {
		return nil
	}

	q := `
	INSERT INTO user_links("user_id", "link_id") VALUES ($1, $2) 
		ON CONFLICT("user_id", "link_id") 
	DO UPDATE SET deleted=FALSE WHERE user_links.user_id=$1 AND user_links.link_id=$2`

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

func (repo *dbRepo) GetOriginalURLByID(id model.LinkID) (string, error) {
	row := repo.db.QueryRow("SELECT original_url FROM links WHERE link_id=$1", id)

	var origURL string
	err := row.Scan(&origURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", model.ErrLinkNotFound
		}
		return "", err
	}

	row = repo.db.QueryRow("SELECT user_id FROM user_links WHERE link_id=$1 AND deleted=FALSE LIMIT 1", id)
	var userID int
	err = row.Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return origURL, model.ErrLinkRemoved
		}
		return "", err
	}

	return origURL, nil
}

func (repo *dbRepo) GetOriginalURLsByUserID(id model.UserID) (map[string]model.LinkID, error) {
	q := `
	SELECT links.link_id, links.original_url FROM links 
	  INNER JOIN user_links ON links.link_id = user_links.link_id
	WHERE user_links.user_id=$1 AND deleted=FALSE`

	rows, err := repo.db.Query(q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	res := make(map[string]model.LinkID)
	for rows.Next() {
		var id model.LinkID
		var origURL string

		err = rows.Scan(&id, &origURL)
		if err != nil {
			return nil, err
		}
		res[origURL] = id
	}

	return res, nil
}

func (repo *dbRepo) DeleteURLs(userID model.UserID, linkIDs []model.LinkID) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `UPDATE user_links SET deleted=TRUE WHERE user_id=$1 AND link_id=$2`

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}

	for _, linkID := range linkIDs {
		if _, err := stmt.Exec(userID, linkID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repo *dbRepo) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := repo.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (repo *dbRepo) Close() error {
	return repo.db.Close()
}

func initDatabase(db *sql.DB) error {
	createTable := func(q string) error {
		var err error
		_, err = db.Exec(q)
		return err
	}

	urlsTable := `CREATE TABLE IF NOT EXISTS links(
	 	link_id SERIAL NOT NULL,
	 	original_url TEXT NOT NULL UNIQUE,
     	PRIMARY KEY (link_id))`

	if err := createTable(urlsTable); err != nil {
		return err
	}

	usersTable := `CREATE TABLE IF NOT EXISTS users(
		user_id SERIAL NOT NULL,
		PRIMARY KEY (user_id))`

	if err := createTable(usersTable); err != nil {
		return err
	}

	userURLsTable := `CREATE TABLE IF NOT EXISTS user_links(
		user_id INTEGER NOT NULL,
		link_id  INTEGER NOT NULL,
		deleted BOOLEAN NOT NULL DEFAULT FALSE,
		UNIQUE(user_id, link_id),
		CONSTRAINT fk_user_id
      		FOREIGN KEY(user_id) REFERENCES users(user_id)
	  		ON DELETE CASCADE,
		CONSTRAINT fk_link_id
			FOREIGN KEY(link_id) REFERENCES links(link_id)
			ON DELETE CASCADE
	)`

	if err := createTable(userURLsTable); err != nil {
		return err
	}
	return nil
}

func cleanDB(db *sql.DB) {
	dropTable := func(table string) {
		_, _ = db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		_, _ = db.Exec(fmt.Sprintf("DROP TABLE %s CASCADE", table))
	}

	dropTable("table")
	dropTable("users")
	dropTable("user_links")
	dropTable("links")
}
