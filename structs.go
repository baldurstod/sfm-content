package main

type WorkshopItem struct {
	Publishedfileid            string        `json:"publishedfileid" mapstructure:"publishedfileid"`
	Title                      string        `json:"title" mapstructure:"title"`
	TimeCreated                uint64        `json:"time_created" mapstructure:"time_created"`
	TimeUpdated                uint64        `json:"time_updated" mapstructure:"time_updated"`
	Creator                    string        `json:"creator" mapstructure:"creator"`
	Tags                       []WorkshopTag `json:"tags" mapstructure:"tags"`
	FileSize                   string        `json:"file_size" mapstructure:"file_size"`
	FileUrl                    string        `json:"file_url" mapstructure:"file_url"`
	PreviewUrl                 string        `json:"preview_url" mapstructure:"preview_url"`
	Subscriptions              uint64        `json:"subscriptions" mapstructure:"subscriptions"`
	ConsumerAppid              uint64        `json:"consumer_appid" mapstructure:"consumer_appid"`
	MaybeInappropriateSex      bool          `json:"maybe_inappropriate_sex" mapstructure:"maybe_inappropriate_sex"`
	MaybeInappropriateViolence bool          `json:"maybe_inappropriate_violence" mapstructure:"maybe_inappropriate_violence"`
}

type WorkshopTag struct {
	DisplayName string `json:"display_name" mapstructure:"display_name"`
	Tag         string `json:"tag" mapstructure:"tag"`
}

type WorkshopItemSummary struct {
	Publishedfileid uint64   `json:"publishedfileid"`
	Title           string   `json:"title"`
	PreviewUrl      string   `json:"preview_url"`
	TimeCreated     uint64   `json:"time_created"`
	TimeUpdated     uint64   `json:"time_updated"`
	Subscriptions   uint64   `json:"subscriptions"`
	Tags            []string `json:"tags"`
}
