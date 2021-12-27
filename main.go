package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/pkg/errors"
)

const IsoTimeFormat = "2006-01-02T15:04:05-0700"

type Config struct {
	Name                  string
	NotionToken           string
	NotionDbId            string
	TwitterConsumerKey    string
	TwitterConsumerSecret string
	TwitterToken          string
	TwitterTokenSecret    string
	WaitTimeMinutes       *int
	WaitTime              *time.Duration
}

func main() {
	log.Println("Starting up!")

	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Panic("(open file) ", err)
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Panic("(read bytes) ", err)
	}
	var configs []Config
	err = json.Unmarshal(byteValue, &configs)
	if err != nil {
		log.Panic("(unmarshal) ", err)
	}
	jsonFile.Close()

	for _, conf := range configs {
		log.Println("starting", conf)
		go start(conf)
	}

	for {
		// keep alive
	}
}

func start(config Config) {
	waitTime := 5 * time.Minute
	if config.WaitTimeMinutes != nil {
		wtim := *config.WaitTimeMinutes
		waitTime = time.Duration(wtim) * time.Minute
	}
	for {
		run(config)
		time.Sleep(waitTime)
	}
}

func run(config Config) {
	log.Println(fmt.Sprintf("(%v) Starting run...", config.Name))
	var page *Page
	client := NotionClient{
		Token: config.NotionToken,
	}
	now := time.Now().Format(IsoTimeFormat)

	// Create a DB query to fetch all records where 'Share On' is less than now
	filter := QueryFilter{
		Filter: &Filter{
			And: &[]Filter{
				{
					Property: StrToPtr("Share At"),
					Date: &DateFilter{
						Before: &now,
					},
				},
				{
					Property: StrToPtr("Ready"),
					Checkbox: &CheckboxFilter{
						Equals: BoolToPtr(true),
					},
				},
			},
		},
		Sorts: &[]Sort{
			{
				Property:  StrToPtr("Share At"),
				Direction: StrToPtr("ascending"),
			},
		},
	}

	// Grab the ID of the first page & run it through the processing
	results, err := client.QueryDatabase(config.NotionDbId, &filter)
	if err != nil {
		log.Fatal(config.Name, err)
	}

	if results.Results != nil && len(*results.Results) > 0 {
		res := *results.Results
		page = &res[0]
	}

	if page != nil {
		content := page.GetTitle()
		log.Println(fmt.Sprintf("(%v) Tweeting: %v", config.Name, content))

		if page.Cover != nil && page.Cover.File != nil && page.Cover.File.Url != nil {
			err = SendTweet(config, content, page.Cover.File.Url)
		} else {
			err = SendTweet(config, content, nil)
		}

		if err != nil {
			log.Fatal(config.Name, err)
		}

		shareCount := 1
		props := *page.Properties
		shareCountP := props["Share Count"]
		if shareCountP.Number != nil {
			shareCount += *shareCountP.Number
		}

		// Update "Last Shared" and "Share On"
		future := "2100-01-01"
		updates := Page{
			Properties: &map[string]Property{
				"Last Shared": {
					Date: &DateProp{
						Start: &now,
					},
				},
				"Share At": {
					Date: &DateProp{
						Start: &future,
					},
				},
				"Share Count": {
					Number: IntToPtr(shareCount),
				},
				"Ready": {
					Checkbox: BoolToPtr(false),
				},
			},
		}

		err = client.UpdatePage(*page.Id, updates)
		if err != nil {
			log.Fatal(config.Name, err)
		}

		log.Println(fmt.Sprintf("(%v) Done!", config.Name))

	} else {
		log.Println(fmt.Sprintf("(%v) Nothing to send.", config.Name))
	}
}

func SendTweet(config Config, content string, imageUrl *string) error {
	oauthConfig := oauth1.NewConfig(config.TwitterConsumerKey, config.TwitterConsumerSecret)
	token := oauth1.NewToken(config.TwitterToken, config.TwitterTokenSecret)
	httpClient := oauthConfig.Client(oauth1.NoContext, token)

	// Twitter client
	twitterClient := twitter.NewClient(httpClient)

	var params twitter.StatusUpdateParams

	if imageUrl != nil {
		mediaService := NewMediaService(httpClient)
		media, err := mediaService.UploadMedia(httpClient, *imageUrl)
		if err != nil {
			return errors.Wrap(err, "(SendTweet) uploading media")
		}
		params = twitter.StatusUpdateParams{
			MediaIds: []int64{
				*media.MediaId,
			},
		}
	}

	_, _, err := twitterClient.Statuses.Update(content, &params)
	if err != nil {
		return errors.Wrap(err, "(SendTweet) sending tweet")
	}

	return nil
}
