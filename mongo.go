package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/abadojack/whatlanggo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DBName          = "mtg"
	cardsCollection = "cards"
)

func InsertTest(collection *mongo.Collection, card bson.M) primitive.ObjectID {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if empty := new(bson.M); &card == empty {
		card = bson.M{"name": "Shock", "printing": "ONS"}
	}
	res, err := collection.InsertOne(ctx, card)
	if err != nil {
		log.Fatal(err)
	}
	id := res.InsertedID
	fmt.Printf("inserted %v\n", id)

	return id.(primitive.ObjectID)
}

func searchCards(searchTerm string) (cardList []string) {
	// connecting
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.MongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	collection := client.Database(DBName).Collection(cardsCollection)

	// Only rus and eng cards in database
	langOpts := whatlanggo.Options{
		Whitelist: map[whatlanggo.Lang]bool{
			whatlanggo.Eng: true,
			whatlanggo.Rus: true,
		},
	}

	langInfo := whatlanggo.DetectWithOptions(searchTerm, langOpts)

	langCode := langInfo.Lang.Iso6391() // short one: "en", "ru", etc.

	log.Printf("Detected lang: %v\n", langCode)

	// Full-Text Search FTW!
	filter := bson.M{"$text": bson.M{"$search": searchTerm, "$language": langCode}}

	// sort by full text score
	findOpts := options.Find()
	score := bson.M{
		"score": bson.M{
			"$meta": "textScore",
		},
	}
	findOpts.Projection = score
	findOpts.Sort = score

	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	cards := make([]string, 0)
	for cur.Next(ctx) {
		var result Card
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		cards = append(cards, fmt.Sprintf("%s [%s]", result.Name, result.Printing))
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
		return cards
	}

	return cards
}

func findOneCard(args ...string) (card Card, err error) {
	searchTerm := args[len(args)-1]

	// connecting
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.MongoURI))
	if err != nil {
		return Card{}, err
	}

	defer client.Disconnect(ctx)
	collection := client.Database(DBName).Collection(cardsCollection)

	// Only rus and eng cards in database
	langOpts := whatlanggo.Options{
		Whitelist: map[whatlanggo.Lang]bool{
			whatlanggo.Eng: true,
			whatlanggo.Rus: true,
		},
	}

	langInfo := whatlanggo.DetectWithOptions(searchTerm, langOpts)

	langCode := langInfo.Lang.Iso6391() // short one: "en", "ru", etc.

	log.Printf("Detected lang: %v\n", langCode)

	// Full-Text Search FTW!
	fts := bson.M{}
	if strings.Contains(args[0], "e:") {
		setCode := strings.ToUpper(strings.Split(args[0], ":")[1])
		fmt.Printf("%v\n", setCode)
		fts = bson.M{
			"$text":    bson.M{"$search": searchTerm, "$language": langCode},
			"printing": bson.M{"$eq": setCode},
		}
	} else {
		fts = bson.M{
			"$text": bson.M{"$search": searchTerm, "$language": langCode},
		}
	}

	card = Card{}

	// sort by full text score
	findOpts := options.FindOne()
	score := bson.M{
		"score": bson.M{
			"$meta": "textScore",
		},
	}
	findOpts.Projection = score
	findOpts.Sort = score

	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	err = collection.FindOne(ctx, fts, findOpts).Decode(&card)
	if err != nil {
		return Card{}, err
	}

	fmt.Printf("%v\n", card)
	return card, nil
}
