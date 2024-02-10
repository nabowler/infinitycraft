package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type (
	Item struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	Pair struct {
		First  string `db:"first"`
		Second string `db:"second"`
	}
)

func initDB(ctx context.Context) (*sqlx.DB, error) {
	db, err := sqlx.Open(dbDriver(), "infinitecraft.db")
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return db, err
	}

	err = func() error {
		_, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS item(
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			UNIQUE(name)
		)`)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pairs(
			first INTEGER NOT NULL,
			second INTEGER NOT NULL,
			result INTEGER NOT NULL,
			PRIMARY KEY(first, second),
			FOREIGN KEY(first) REFERENCES item(id),
			FOREIGN KEY(second) REFERENCES item(id),
			FOREIGN KEY(result) REFERENCES item(id)
		)`)
		if err != nil {
			return err
		}

		// Initial values the game starts with
		_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO item(name) VALUES ("Water"),("Fire"),("Wind"),("Earth")`)
		if err != nil {
			return err
		}

		return tx.Commit()
	}()
	if err != nil {
		_ = tx.Rollback()
		return db, err
	}

	return db, nil
}

func getRandomItem(ctx context.Context, db *sqlx.DB) (Item, error) {
	var item Item
	err := db.GetContext(ctx, &item, `SELECT id, name FROM item ORDER BY RANDOM() LIMIT 1`)
	return item, err
}

func getPair(ctx context.Context, db *sqlx.DB, first, second int) (Pair, error) {
	var pair Pair
	err := db.GetContext(ctx, &pair, db.Rebind(`SELECT first, second FROM pairs WHERE first =? AND second =?`), first, second)
	return pair, err
}

func getTwoUnpairedItems(ctx context.Context, db *sqlx.DB) (first Item, second Item, err error) {
	for ctx.Err() == nil {
		// this could probably be done in one hard-to-understand query instead of 3.
		// the only _real_ benefit would be that I no longer need an infinite-loop to guess-and-check.

		// getting one random item twice allows for the same item to be both first and second
		first, err = getRandomItem(ctx, db)
		if err != nil {
			// error getting the first item
			return
		}
		second, err = getRandomItem(ctx, db)
		if err != nil {
			// error getting the second item
			return
		}

		_, pairErr := getPair(ctx, db, first.ID, second.ID)
		if errors.Is(pairErr, sql.ErrNoRows) {
			// Success
			return
		}
		if pairErr != nil {
			// Sad trombone
			err = pairErr
			return
		}

		// try again
	}
	return
}

func insertNewPair(ctx context.Context, db *sqlx.DB, first, second Item, result Result) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	err = func() error {
		_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO item(name) VALUES (?)`, result.Result)
		if err != nil {
			return err
		}

		var resultId int
		err = tx.GetContext(ctx, &resultId, `SELECT id FROM item WHERE name =?`, result.Result)
		if err != nil {
			return err
		}

		_, err := tx.ExecContext(ctx, db.Rebind(`INSERT INTO pairs(first, second, result) VALUES (?,?,?)`), first.ID, second.ID, resultId)
		if err != nil {
			return nil
		}

		return tx.Commit()
	}()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}
