package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	MongoURI        string `toml:"mongo_uri"`
	BotToken        string `toml:"tg_bot_token"`
	CacheDir        string `toml:"cache_dir"`
	Endpoint        string `toml:"minio_endpoint"`
	Bucket          string `toml:"minio_bucket"`
	AccessKeyID     string `toml:"minio_access_key"`
	SecretAccessKey string `toml:"minio_secret"`
}

var conf Config

func init() {
	if _, err := toml.DecodeFile("meme.toml", &conf); err != nil {
		panic("Could not read config meme.toml")
	}

	st, err := NewStorage(conf.Endpoint, conf.AccessKeyID, conf.SecretAccessKey, false, conf.Bucket)
	if err != nil {
		panic("Could not connect to storage")
	}

	err = st.CreateBucket()
	if err != nil {
		panic("Could not create bucket")
	}
}
