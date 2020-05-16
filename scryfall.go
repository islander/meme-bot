package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const scryfallTag = "scryfall"
const BaseURL string = "https://api.scryfall.com"

type SFClient struct {
	BaseURL   *url.URL
	UserAgent string

	httpClient *http.Client
}

func NewSFClient() *SFClient {
	return &SFClient{httpClient: &http.Client{}}
}

type Card struct {
	ID         primitive.ObjectID `scryfall:"id" bson:"_id" json:"id,omitempty"`
	Name       string             `scryfall:"name" json:"name"`
	Number     string             `scryfall:"number" bson:"number"`
	Printing   string             `scryfall:"code" json:"printing"`
	Side       string             `scryfall:"side" json:"side,omitempty"`
	ScryfallID string             `scryfall:"objectid" bson:"scryfallId"`
	Language   string             `scryfall:"lang" json:"lang,omitempty"`
	Face       string             `scryfall:"face" json:"face,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at,omitempty"`
}

// format api url by struct tags
func (card *Card) formatApiUrl(format string) string {
	v := reflect.ValueOf(card).Elem()

	args := make([]string, v.NumField()*2)

	for i, pos := 0, 0; i < v.NumField(); i++ {
		tags := strings.Split(v.Type().Field(i).Tag.Get(scryfallTag), ",")
		key := tags[0]
		val := v.Field(i)
		args[pos] = "{" + key + "}"
		args[pos+1] = fmt.Sprint(val)
		pos = pos + 2
	}

	return strings.ToLower(strings.NewReplacer(args...).Replace(format))
}

func (sf *SFClient) GetImage(ctx context.Context, card Card) (string, error) {
	// double-faced cards is a pain =)
	switch card.Side {
	case "a":
		card.Face = "front"
	case "b":
		card.Face = "back"
	default:
		card.Face = "front"
	}

	fileName := conf.CacheDir + "/" + card.ScryfallID + "[" + card.Language + "][" + card.Face + "].jpg"

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		base, err := url.Parse(BaseURL)

		relativeUrl := card.formatApiUrl("/cards/{code}/{number}/{lang}")
		u, err := url.Parse(relativeUrl)
		if err != nil {
			log.Fatal(err)
		}
		queryString := u.Query()
		queryString.Set("face", card.Face)
		queryString.Set("format", "image")
		queryString.Set("version", "normal")
		u.RawQuery = queryString.Encode()

		req, err := http.NewRequest("GET", base.ResolveReference(u).String(), nil)
		if err != nil {
			return "", err
		}
		body, err := sf.do(ctx, req)
		if err != nil {
			return "", err
		}

		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		_, err = file.Write(body)
		if err != nil {
			log.Fatal(err)
		}
	}
	return fileName, nil
}

func (sf *SFClient) do(ctx context.Context, req *http.Request) ([]byte, error) {
	req = req.WithContext(ctx)

	resp, err := sf.httpClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}

	// close the HTTP response body and
	// drain the response body before closing it
	defer func() {
		maxCopySize := 2 << 10
		io.CopyN(ioutil.Discard, resp.Body, int64(maxCopySize))
		defer resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if 200 != resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()
		return nil, fmt.Errorf("%s", body)
	}
	return body, nil
}
