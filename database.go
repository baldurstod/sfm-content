package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	_ "github.com/lib/pq"
)

var db *sql.DB

var sortFields = map[string]string{
	"index":         "publishedfileid",
	"name":          "title",
	"subscriptions": "subscriptions",
	"updated":       "time_updated",
	"created":       "time_created",
}

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
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14 )
	ON CONFLICT (publishedfileid) DO UPDATE SET
	title = $2,
	time_created = $3,
	time_updated = $4,
	creator = $5,
	tags = $6,
	file_size = $7,
	file_url = $8,
	preview_url = $9,
	subscriptions = $10,
	consumer_appid = $11,
	maybe_inappropriate_sex = $12,
	maybe_inappropriate_violence = $13,
	detail = $14
	`,
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

type itemFilter struct {
	Name string   `json:"name" mapstructure:"name"`
	Tags []string `json:"tags" mapstructure:"tags"`
}

type itemSort struct {
	Field     string `json:"field" mapstructure:"field"`
	Ascending bool   `json:"ascending" mapstructure:"ascending"`
}

type itemParams struct {
	Filter itemFilter `json:"filter" mapstructure:"filter"`
	Sort   itemSort   `json:"sort" mapstructure:"sort"`
}

func getItems(params itemParams) ([]WorkshopItemSummary, error) {
	if db == nil {
		return nil, errors.New("database is not initialized. Did you forgot to call openPostgre ?")
	}

	if len(params.Filter.Tags) > 5 {
		return nil, errors.New("too much tags")
	}

	values := []any{}

	sortField, found := sortFields[params.Sort.Field]
	if !found {
		sortField = "time_created"
	}

	var sortDirection = "desc"
	if params.Sort.Ascending {
		sortDirection = "asc"
	}

	valueIndex := 1
	namePredicate := ""
	if params.Filter.Name != "" {
		// title
		namePredicate = " AND title ILIKE $" + strconv.Itoa(valueIndex)
		valueIndex++
		values = append(values, "%"+strings.ReplaceAll(strings.ReplaceAll(params.Filter.Name, "%", `\%`), "_", `\_`)+"%")
	}

	keys := []string{}
	for i, v := range params.Filter.Tags {
		keys = append(keys, " AND $"+strconv.Itoa(i+valueIndex)+"=ANY(tags)")
		values = append(values, v)
	}

	tagsPredicate := strings.Join(keys, "")

	query := `SELECT publishedfileid, title, preview_url FROM items WHERE TRUE ` + namePredicate + tagsPredicate + ` AND 'Model'=ANY(tags) ORDER BY ` + sortField + ` ` + sortDirection + ` LIMIT(1000);`
	res, err := db.Query(query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query "+query+"in getItems: <%w>", err)
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

func getItemUrl(id uint64) (string, error) {
	row := db.QueryRow(`SELECT file_url FROM items WHERE publishedfileid = $1;`, id)

	file_url := ""

	err := row.Scan(&file_url)
	if err != nil {
		return "", fmt.Errorf("failed to scan row in getItems: <%w>", err)
	}

	return file_url, nil
}
