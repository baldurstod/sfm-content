package main

import (
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
