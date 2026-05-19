package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/go-viper/mapstructure/v2"
	_ "github.com/lib/pq"
)

var db *sql.DB

func openPostgre(dataSourceName string) {
	var err error
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	// db.Open() only creates a connection pool, and doesn't actually establish
	// a connection. To ensure the connection works you need to do *something*
	// with a connection.
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func closePostgre() {
	if db != nil {
		db.Close()
	}
}

func insertItem(item map[string]any) error {
	if db == nil {
		return errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	var workshopItem WorkshopItem
	err := mapstructure.Decode(item, &workshopItem)
	if err != nil {
		return fmt.Errorf("an error occured while decoding workshop item : <%w>", err)
	}

	var tags = []string{}

	for _, t := range workshopItem.Tags {
		tags = append(tags, t.Tag)
	}

	j, err := json.Marshal(&item)
	if err != nil {
		return fmt.Errorf("failed to marshal json: <%w>", err)
	}

	fileId, err := strconv.ParseUint(workshopItem.Publishedfileid, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to convert Publishedfileid "+workshopItem.Publishedfileid+": <%w>", err)
	}

	fileSize, err := strconv.ParseUint(workshopItem.FileSize, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to convert file size "+workshopItem.FileSize+": <%w>", err)
	}

	_, err = db.Exec(`INSERT INTO items (
	publishedfileid, title, time_created, time_updated, creator, tags, file_size, file_url, preview_url, subscriptions, consumer_appid, maybe_inappropriate_sex, maybe_inappropriate_violence, detail )
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14 )`,
		fileId,
		workshopItem.Title,
		workshopItem.TimeCreated,
		workshopItem.TimeUpdated,
		workshopItem.Creator,
		tags,
		fileSize,
		workshopItem.FileUrl,
		workshopItem.PreviewUrl,
		workshopItem.Subscriptions,
		workshopItem.ConsumerAppid,
		workshopItem.MaybeInappropriateSex,
		workshopItem.MaybeInappropriateViolence,
		j,
	)

	if err != nil {
		return fmt.Errorf("failed to insert workshop item "+workshopItem.Publishedfileid+" : <%w>", err)
	}

	return nil
}

func getItems() ([]WorkshopItemSummary, error) {
	if db == nil {
		return nil, errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	res, err := db.Query(`SELECT publishedfileid, title, preview_url FROM items LIMIT(1000);`)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query in getItems: <%w>", err)
	}
	defer res.Close()

	result := []WorkshopItemSummary{}
	for {
		if ok := res.Next(); !ok {
			// Check if we have an error
			if err := res.Err(); err != nil {
				return nil, fmt.Errorf("failed to get next row in getItems: <%w>", err)
			}
			// If no error, exit the loop
			break
		}

		var publishedfileid uint64
		var title string
		var fileUrl string

		err = res.Scan(&publishedfileid, &title, &fileUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row in getItems: <%w>", err)
		}

		result = append(result, WorkshopItemSummary{publishedfileid, title, fileUrl})
	}

	return result, nil
}
