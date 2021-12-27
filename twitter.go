package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"
)

type Media struct {
	MediaId       *int64  `json:"media_id,omitempty"`
	MediaIdString *string `json:"media_id_string,omitempty"`
}

type MediaService struct {
	sling *sling.Sling
}

func NewMediaService(httpClient *http.Client) MediaService {
	s := sling.New().Client(httpClient).Base("https://upload.twitter.com/1.1/")

	return MediaService{
		sling: s,
	}
}

type UploadMediaRequest struct {
	MediaData string `url:"media_data,omitempty"`
}

func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func (s *MediaService) UploadMedia(httpClient *http.Client, mediaUrl string) (*Media, error) {
	resp, err := http.Get(mediaUrl)
	if err != nil {
		return nil, errors.Wrap(err, "(UploadMedia) http get mediaUrl")
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.New(fmt.Sprintf("(UploadMedia) failed to get mediaUrl with status %v.", resp.StatusCode))
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "(UploadMedia) ioutil readall img response")
	}
	base64Encoding := toBase64(bytes)

	formData := url.Values{
		"media_data": {base64Encoding},
	}
	req, err := http.NewRequest("POST", "https://upload.twitter.com/1.1/media/upload.json", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "(UploadMedia) new twitter upload req")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = httpClient.Do(req)
	if err != nil {
		return nil, errors.New("(UploadMedia) exec post to twitter")
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.New(fmt.Sprintf("(UploadMedia) failed upload media to twitter with status %v.", resp.StatusCode))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "(UploadMedia) ioutil readall")
	}

	var media Media
	err = json.Unmarshal(body, &media)
	if err != nil {
		return nil, errors.Wrap(err, "(UploadMedia) unmarshal response")
	}

	return &media, nil
}
