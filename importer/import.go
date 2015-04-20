package main

import (
	"bytes"
	"encoding/json"
	"github.com/martingartonft/timemachine/api"

	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type importerContent struct {
	/*
		//ID            string   `json:"id"`
		UUID          string    `json:"uuid"`
		URI           string    `json:"id"`
		Body       string    `json:"body,omitempty"`
		Brands        map[string]interface{} `json:"brands,omitempty"`
		Byline        string    `json:"byline,omitempty"`
		PublishedDate time.Time `json:"publishedDate,omitempty"`
		//RequestUrl    string   `json:"requestUrl"`
		Title string `json:"title,omitempty"`
		//Type   string `json:"type"`
		WebUrl string `json:"webUrl,omitempty"`
	*/
	Body   string `json:"body"`
	Brands []struct {
		ID string `json:"id"`
	} `json:"brands"`
	Byline      string      `json:"byline"`
	Description interface{} `json:"description"`
	Identifiers []struct {
		Authority       string `json:"authority"`
		IdentifierValue string `json:"identifierValue"`
	} `json:"identifiers"`
	InternalBinaryUrl interface{} `json:"internalBinaryUrl"`
	MainImage         string      `json:"mainImage"`
	MediaType         interface{} `json:"mediaType"`
	Members           interface{} `json:"members"`
	PixelHeight       interface{} `json:"pixelHeight"`
	PixelWidth        interface{} `json:"pixelWidth"`
	PublishedDate     time.Time   `json:"publishedDate"`
	Title             string      `json:"title"`
	Titles            interface{} `json:"titles"`
	UUID              string      `json:"uuid"`
}

func main() {
	//	uuids := []string{"16551bd2-5960-4cf3-a8fa-db4707df470c"}
	uuids := []string{"b808e3a2-e740-11e4-8e3f-00144feab7de"}

	eomFileJsons := make(chan string, 16)
	contentJsons := make(chan []byte, 16)

	go queueEomJsons(eomFileJsons, uuids)

	go queueContentJsons(eomFileJsons, contentJsons)

	for contentJson := range contentJsons {
		//fmt.Printf("%s\n", contentJson)

		var ic importerContent
		err := json.Unmarshal(contentJson, &ic)

		cont := api.Content{
			UUID:          ic.UUID,
			URI:           fmt.Sprintf("http://api.ft.com/things/%s", ic.UUID),
			BodyXML:       ic.Body,
			Byline:        ic.Byline,
			PublishedDate: ic.PublishedDate,
			Title:         ic.Title,
		}
		for _, brand := range ic.Brands {
			cont.Brands = append(cont.Brands, brand.ID)
			fmt.Printf("brand: %v\n", brand)
		}

		apiCont, err := json.Marshal(cont)
		if err != nil {
			panic(err)
		}

		req, err := http.NewRequest("PUT", "http://localhost:8082/content/"+ic.UUID, bytes.NewReader(apiCont))
		if err != nil {
			panic(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}

		if err != nil {
			panic(err)
		}
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic("reading response body failed")
		}
		if resp.StatusCode != 200 {
			panic(fmt.Sprintf("save failed with code %d:\n%v", resp.StatusCode, string(response)))
		}
		resp.Body.Close()
		fmt.Printf("%v\n", response)
	}

}

func queueEomJsons(eomJsons chan<- string, uuids []string) {
	for _, uuid := range uuids {
		resp, err := http.Get(fmt.Sprintf("http://localhost:9080/eom-file/all/%s", uuid))
		if err != nil {
			panic(err)
		}
		var eomObjects []map[string]interface{}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&eomObjects); err != nil {
			panic(err)
		}
		resp.Body.Close()

		fmt.Printf("object count: %d\n", len(eomObjects))

		for _, eomObj := range eomObjects {
			eomJson, err := json.Marshal(eomObj)
			if err != nil {
				panic(err)
			}
			eomJsons <- string(eomJson)
		}
		close(eomJsons)
	}
}

func queueContentJsons(eomFileJsons <-chan string, contentJsons chan []byte) {
	for eomJson := range eomFileJsons {
		resp, err := http.Post("http://localhost:11070/transform/eom-file/", "application/json", bytes.NewReader([]byte(eomJson)))
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != 200 {
			panic(fmt.Sprintf("transform failed with code %d\n", resp.StatusCode))
		}
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic("reading response body failed")
		}
		contentJsons <- response
		resp.Body.Close()
	}
	close(contentJsons)
}
