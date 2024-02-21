package flag

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	rv8 "github.com/go-redis/redis/v8"
)

var (
	cache *rv8.Client

	// use sync.Once to ensure that the initialization code is executed only once.
	once sync.Once
)

type ReleaseFlagModel struct {
	Flag      string    `json:"flag" db:"flag"`
	IsEnable  bool      `json:"is_enable" db:"is_enable"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	EnabledAt time.Time `json:"enabled_at" db:"enabled_at"`
}

func (ReleaseFlagModel) TableName() string {
	return "release_flags"
}

func Init(dbConn *sql.DB, _cache *rv8.Client) {

	var releaseFlags []ReleaseFlagModel

	rows, err := dbConn.Query("SELECT * FROM release_flags WHERE is_enable = true")
	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatal("Error when get release flags from database: ", err)
		}

	}

	for rows.Next() {
		var flag ReleaseFlagModel
		err := rows.Scan(&flag.Flag, &flag.IsEnable, &flag.CreatedAt, &flag.EnabledAt)
		if err != nil {
			log.Fatal("Error when scan release flags from database: ", err)
		}
	}

	ctx := context.Background()

	once.Do(func() {
		if cache == nil {
			cache = _cache
		}

		for _, flag := range releaseFlags {
			b, err := json.Marshal(flag)
			if err != nil {
				continue
			}
			cache.Set(ctx, flag.Flag, string(b), 0)
		}
	})
}

func InitMock(_cache *rv8.Client) {
	cache = _cache
}

// IsEnable check if flag is enabled
// this function use for logic that need to be enable/disable
// example use:
//
//	if flagger.IsEnable("promo_segmentation_buyer") {
//	    do something here
//	}
func IsEnable(flag string) bool {

	data, err := cache.Get(context.Background(), flag).Result()
	if err != nil {
		return false
	}

	if data != "" {
		return true
	}

	return false
}

// create release flag middleware with specific flag
// example use:
// http.HandleFunc("/endpoint", flagger.ReleaseFlag("promo_segmentation_buyer", handler))
func ReleaseFlag(flag string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if !IsEnable(flag) {
			http.Error(w, "Flag is not enabled", http.StatusForbidden)
			return
		}

		next(w, r)
	}

}
