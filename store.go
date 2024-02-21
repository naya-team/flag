package flag

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	rv8 "github.com/go-redis/redis/v8"
)

type flagger struct {
	db    *sql.DB
	cache *rv8.Client
}

type Flagger interface {
	StoreFlag(flag string, isEnable bool) (ReleaseFlagModel, error)
	GetFlags(flag string) (flags []ReleaseFlagModel, err error)
	DisableFlag(flag string) (ReleaseFlagModel, error)
	EnableFlag(flag string) (ReleaseFlagModel, error)
}

func New(db *sql.DB, cache *rv8.Client) *flagger {
	return &flagger{
		db:    db,
		cache: cache,
	}
}

func (f *flagger) StoreFlag(flag string, isEnable bool) (ReleaseFlagModel, error) {

	var now = time.Now()

	var model = ReleaseFlagModel{
		Flag:      flag,
		IsEnable:  isEnable,
		CreatedAt: now,
	}

	if isEnable {
		model.EnabledAt = now
	}

	_, err := f.db.Exec("INSERT INTO release_flags (flag, is_enable, created_at, enabled_at) VALUES ($1, $2, $3, $4)", model.Flag, model.IsEnable, model.CreatedAt, model.EnabledAt)
	if err != nil {
		return model, err
	}

	// store to memory only if flag is enable
	if isEnable {

		b, err := json.Marshal(model)
		if err != nil {
			return model, err
		}

		if err := f.cache.Set(context.Background(), flag, string(b), 0).Err(); err != nil {
			return model, err
		}
	}

	return model, nil
}

func (f *flagger) GetFlags(flag string) (flags []ReleaseFlagModel, err error) {

	var rows *sql.Rows

	query := "SELECT * FROM release_flags"

	if flag != "" {
		query += " WHERE flag LIKE ?"
		rows, err = f.db.Query(query, flag)
		if err != nil {
			return nil, err
		}
	} else {
		rows, err = f.db.Query(query)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	for rows.Next() {
		var _flag ReleaseFlagModel
		err = rows.Scan(&_flag.Flag, &_flag.IsEnable, &_flag.CreatedAt, &_flag.EnabledAt)
		if err != nil {
			return nil, err

		}
	}

	return flags, nil
}

// disable flag
func (f *flagger) DisableFlag(flag string) (ReleaseFlagModel, error) {
	// update flag on db set is_enable to false
	var model ReleaseFlagModel

	err := f.db.QueryRow("SELECT * FROM release_flags WHERE flag = ?", flag).Scan(
		model,
	)
	if err != nil {
		return model, err
	}

	model.IsEnable = false
	_, err = f.db.Exec("UPDATE release_flags SET is_enable = ? WHERE flag = ?", model.IsEnable, flag)
	if err != nil {
		return model, err
	}

	if err := f.cache.Del(context.Background(), flag).Err(); err != nil {
		// check error not redis nil
		if err != rv8.Nil {
			return model, err
		}
	}

	return model, nil
}

// enable flag
func (f *flagger) EnableFlag(flag string) (ReleaseFlagModel, error) {

	var model ReleaseFlagModel
	err := f.db.QueryRow("SELECT * FROM release_flags WHERE flag = ?", flag).Scan(
		model,
	)
	if err != nil {
		return model, err
	}

	model.IsEnable = false
	model.EnabledAt = time.Now()
	_, err = f.db.Exec("UPDATE release_flags SET is_enable = ?, enabled_at = ? WHERE flag = ?", model.IsEnable, model.EnabledAt, flag)
	if err != nil {
		return model, err
	}
	b, err := json.Marshal(model)
	if err != nil {
		return model, err
	}
	// update flag on memory
	if err := f.cache.Set(context.Background(), flag, string(b), 0).Err(); err != nil {
		return model, err
	}

	return model, nil
}
