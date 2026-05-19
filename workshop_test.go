package main

import (
	"log"
	"testing"
)

func TestRefreshItems(t *testing.T) {

	config, err := readConfig()
	if err != nil {
		t.Error(err)
		return
	}

	setApiKey(config.ApiKey)

	openPostgre(config.DataSource)
	defer closePostgre()

	err = refreshItems()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGetItems(t *testing.T) {
	config, err := readConfig()
	if err != nil {
		t.Error(err)
		return
	}

	setApiKey(config.ApiKey)

	openPostgre(config.DataSource)
	defer closePostgre()

	result, err := getItems()
	if err != nil {
		t.Error(err)
		return
	}

	log.Println(result)
}
