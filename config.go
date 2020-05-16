package main

import "github.com/BurntSushi/toml"

type Config struct {
	MongoURI string `toml:"mongo_uri"`
	BotToken string `toml:"tg_bot_token"`
	CacheDir string `toml:"cache_dir"`
}

var conf Config

func init() {
	if _, err := toml.DecodeFile("meme.toml", &conf); err != nil {
		panic("Could not read config meme.toml")
	}
}
