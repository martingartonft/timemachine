package main

import (
	"bytes"
	"encoding/json"
	"github.com/martingartonft/timemachine/api"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type importerContent struct {
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
	PublishedDate     time.Time   `json:"publishedDate"`
	Title             string      `json:"title"`
	Titles            interface{} `json:"titles"`
	UUID              string      `json:"uuid"`
}

func main() {
	//	uuids := []string{"16551bd2-5960-4cf3-a8fa-db4707df470c"}
	uuids := []string{
		"f00dfbc8-e4d6-11e4-8b61-00144feab7de",
		"a4013768-e7e7-11e4-894a-00144feab7de",
		"fb1469fc-e7f7-11e4-894a-00144feab7de",
		"2e61fc44-e761-11e4-a01c-00144feab7de",
		"872fd0cc-e7f6-11e4-9960-00144feab7de",
		"e5cf8b74-e7f9-11e4-894a-00144feab7de",
		"a88667a8-e7f7-11e4-894a-00144feab7de",
		"75ef1d94-e784-11e4-8ebb-00144feab7de",
		"a7c21a68-e515-11e4-bb4b-00144feab7de",
		"658529e4-e7f2-11e4-9960-00144feab7de",
		"726977dc-e76b-11e4-a01c-00144feab7de",
		"d65bbf32-e7b3-11e4-8e3f-00144feab7de",
		"9fe1a040-debf-11e4-8a01-00144feab7de",
		"c1823f5a-e489-11e4-9039-00144feab7de",
		"e09b49d6-e1fa-11e4-bb7f-00144feab7de",
	}

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
	}
	close(eomJsons)
}

func queueContentJsons(eomFileJsons <-chan string, contentJsons chan []byte) {
	for eomJson := range eomFileJsons {
		resp, err := http.Post("http://localhost:11070/transform/eom-file/", "application/json", bytes.NewReader([]byte(eomJson)))
		if err != nil {
			panic(err)
		}
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic("reading response body failed")
		}
		if resp.StatusCode != 200 {

			log.Printf("transform failed for %v with code %d, skipping:\n%v\n", eomJson, resp.StatusCode, string(response))
			continue
		}
		contentJsons <- response
		resp.Body.Close()
	}
	close(contentJsons)
}
