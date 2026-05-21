package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var apiKey = ""

const sfmAppId = 1840
const retries = 10

const k_PublishedFileQueryType_RankedByLastUpdatedDate = 21

func refreshWorkshop() {

}

func setApiKey(key string) {
	apiKey = key
}

func startWorkshop() {
}

func runWorkshopTasks() {
	go func() {
		for {
			refreshItems()
			time.Sleep(5 * time.Minute)
		}
	}()
}

func refreshItems() error {
	log.Println("Starting items refresh")

	cursor := "*"
	page := 1

	for {
		log.Println("readWorkshopPage cursor "+cursor+" page", page)
		content, err := readWorkshopPage(cursor)
		if err != nil {
			return err
		}

		response, ok := content["response"].(map[string]any)
		if !ok {
			return errors.New("missing response key in api response")
		}

		next, ok := response["next_cursor"].(string)
		if !ok {
			return errors.New("next_cursor is missing or don't have an expected format")
		}

		if next == cursor {
			// Pagination end
			break
		}

		files, ok := response["publishedfiledetails"].([]any)
		if !ok {
			return errors.New("publishedfiledetails don't have an expected format")
		}

		ids := []string{}
		for _, f := range files {
			file, ok := f.(map[string]any)
			if !ok {
				return errors.New("publishedfiledetails don't have an expected format")
			}
			result, ok := file["result"].(float64)
			if !ok {
				return errors.New("publishedfiledetails don't have an expected format for key result")
			}
			if result == 0 {
				return errors.New("publishedfiledetails result is not 1")
			}
			if result != 1 {
				continue
			}
			id, ok := file["publishedfileid"].(string)
			if !ok {
				return errors.New("publishedfiledetails don't have an expected format for key publishedfileid")
			}

			ids = append(ids, id)
		}

		details, err := workshopGetDetails(ids)
		if err != nil {
			return fmt.Errorf("an error occured while reading items for cursor "+cursor+": <%w>", err)
		}

		items := details["response"].(map[string]any)["publishedfiledetails"].([]any)
		for _, item := range items {
			i, ok := item.(map[string]any)
			if !ok {
				return fmt.Errorf("unexpected item type <%w>", err)
			}

			err = insertItem(i)
			if err != nil {
				return fmt.Errorf("failed to insert item : <%w>", err)
			}
		}

		page++
		cursor = next
	}

	return nil
}

func writeJson(name string, data map[string]any) {
	j, _ := json.MarshalIndent(&data, "", "\t")

	os.WriteFile("./var/"+name+".json", j, 0666)
}

func readWorkshopPage(cursor string) (map[string]any, error) {
	if apiKey == "" {
		return nil, errors.New("missing api key. Did you forgot to call setApiKey ?")
	}

	numPerPage := 100
	queryType := k_PublishedFileQueryType_RankedByLastUpdatedDate

	params := url.Values{
		"appid":            {strconv.Itoa(sfmAppId)},
		"cursor":           {cursor},
		"numperpage":       {strconv.Itoa(numPerPage)},
		"query_type":       {strconv.Itoa(queryType)},
		"key":              {apiKey},
		"search_text":      {""},
		"return_previews":  {"1"},
		"return_vote_data": {"1"},
		"return_tags":      {"1"},
		"return_metadata":  {"1"},
	}

	u, err := url.Parse(steamApiQueryFiles)
	if err != nil {
		return nil, errors.New("can't parse steamApiQueryFiles")
	}
	u.RawQuery = params.Encode()
	//c.Redirect(http.StatusSeeOther, u.String())
	response, err := getApi(u)
	if err != nil {
		return nil, fmt.Errorf("failed to get api response in readWorkshopPage: <%w>", err)
	}

	return response, nil

}

func workshopGetDetails(items []string) (map[string]any, error) {
	//log.Println("processing items", items)
	if apiKey == "" {
		return nil, errors.New("missing api key. Did you forgot to call setApiKey ?")
	}

	params := url.Values{
		"appid": {strconv.Itoa(sfmAppId)},
		"key":   {apiKey},
	}

	for i, id := range items {
		params.Add("publishedfileids["+strconv.Itoa(i)+"]", id)
	}

	u, err := url.Parse(steamApiGetDetails)
	if err != nil {
		return nil, errors.New("can't parse steamApiGetDetails")
	}
	u.RawQuery = params.Encode()
	response, err := getApi(u)
	if err != nil {
		return nil, fmt.Errorf("failed to get api response in readWorkshopItems: <%w>", err)
	}

	return response, nil
}

func getApi(u *url.URL) (map[string]any, error) {
	resp, err := getPage(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get api response in getApi: <%w>", err)
	}
	defer resp.Body.Close()

	response := map[string]any{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json in getApi: <%w>", err)
	}

	return response, nil
}

func getPage(url string) (*http.Response, error) {
	var resp *http.Response
	for range retries {

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 429 { //Too Many Requests
			time.Sleep(60 * time.Second)
			continue
		}

		if resp.StatusCode != 200 { //Everything except 429 and 200
			return nil, fmt.Errorf("http request returned HTTP status code: %d", resp.StatusCode)
		}
		break
	}

	return resp, nil
}
