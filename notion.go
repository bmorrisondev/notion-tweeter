package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

type NotionClient struct {
	Token string
}

type Property struct {
	Id       *string      `json:"id,omitempty"`
	Type     *string      `json:"type,omitempty"`
	Name     *string      `json:"name,omitempty"`
	Formula  *FormulaProp `json:"formula,omitempty"`
	Title    *[]TitleProp `json:"title,omitempty"`
	Date     *DateProp    `json:"date,omitempty"`
	Number   *int         `json:"number,omitempty"`
	Checkbox *bool        `json:"checkbox,omitempty"`
}

type DateProp struct {
	Start *string `json:"start,omitempty"`
	End   *string `json:"end,omitempty"`
}
type TitleProp struct {
	Type      *string   `json:"type,omitempty"`
	Text      *TextProp `json:"text,omitempty"`
	PlainText *string   `json:"plain_text,omitempty"`
	Href      *string   `json:"href,omitempty"`
}

type TextProp struct {
	Content *string `json:"content,omitempty"`
	Link    *Link   `json:"link,omitempty"`
}

type Link struct {
	Url *string `json:"url,omitempty"`
}

type FormulaProp struct {
	Type   *string `json:"type,omitempty"`
	String *string `json:"string,omitempty"`
	Number *int    `json:"number,omitempty"`
}

type FileObject struct {
	Type *string         `json:"type,omitempty"`
	File *FileUrlWrapper `json:"file,omitempty"`
}

type FileUrlWrapper struct {
	Url        *string `json:"url,omitempty"`
	ExpiryTime *string `json:"expiry_time,omitempty"`
}

type QueryResult struct {
	Object  *string `json:"object,omitempty"`
	Results *[]Page `json:"results,omitempty"`
	Code    string  `json:"code"`
	Message string  `json:"message"`
}

func (c *NotionClient) callApi(path string, method string, headers *map[string]string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("https://api.notion.com/v1%v", path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "(callApi) create request")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.Token))
	req.Header.Set("Notion-Version", "2021-05-13")
	if headers != nil {
		for idx, el := range *headers {
			req.Header.Set(idx, el)
		}
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "(callApi) exec request")
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		out, _ := ioutil.ReadAll(res.Body)
		return nil, errors.New(fmt.Sprintf("Failed to call %v with status %v, response: %v", path, res.StatusCode, string(out)))
	}

	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "(callApi) ioutil.readall")
	}
	return out, nil
}

func (c *NotionClient) QueryDatabase(databaseId string, queryFilter *QueryFilter) (*QueryResult, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var qr QueryResult
	var body *bytes.Buffer

	if queryFilter != nil {
		jsonData, err := json.Marshal(queryFilter)
		if err != nil {
			return nil, errors.Wrap(err, "(QueryDatabase) marshal json")
		}
		body = bytes.NewBuffer(jsonData)
	}

	path := fmt.Sprintf("/databases/%v/query", databaseId)
	bytes, err := c.callApi(path, "POST", &headers, body)
	if err != nil {
		return nil, errors.Wrap(err, "(QueryDatabase) callApi")
	}

	err = json.Unmarshal(bytes, &qr)
	if err != nil {
		return nil, errors.Wrap(err, "(QueryDatabase) unmarshal response")
	}

	return &qr, nil
}

func (c *NotionClient) UpdatePage(pageId string, updates Page) error {
	jsonData, err := json.Marshal(updates)
	if err != nil {
		return errors.Wrap(err, "(UpdatePage) marshal json")
	}
	body := bytes.NewBuffer(jsonData)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	path := fmt.Sprintf("/pages/%v", pageId)
	_, err = c.callApi(path, "PATCH", &headers, body)
	return err
}

func (c *NotionClient) FetchPage(pageId string) (*Page, error) {
	path := fmt.Sprintf("/pages/%v", pageId)
	bytes, err := c.callApi(path, "GET", nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "(FetchPage) callApi")
	}

	var page Page
	err = json.Unmarshal(bytes, &page)
	if err != nil {
		return nil, errors.Wrap(err, "(FetchPage) unmarshal json")
	}

	return &page, nil
}

func (c *NotionClient) CreatePage(page Page) error {
	jsonData, err := json.Marshal(page)
	if err != nil {
		return errors.Wrap(err, "(UpdatePage) marshal json")
	}
	body := bytes.NewBuffer(jsonData)

	log.Println(string(jsonData))

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	_, err = c.callApi("/pages", "POST", &headers, body)
	return err
}
