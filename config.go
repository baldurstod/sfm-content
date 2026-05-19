package main

type Config struct {
	HTTP     `json:"http"`
	Database `json:"database"`
	Steam    `json:"steam"`
}

type HTTP struct {
	Port          int    `json:"port"`
	HttpsKeyFile  string `json:"https_key_file"`
	HttpsCertFile string `json:"https_cert_file"`
}

type Steam struct {
	ApiKey string `json:"api_key"`
}

type Database struct {
	DataSource string `json:"datasource"`
}
