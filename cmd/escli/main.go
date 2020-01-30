package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
)

func main() {
	var isHelp bool
	var command string
	var indexName string
	var docID string

	flag.BoolVar(&isHelp, "h", false, "prints the usage of this executable")
	flag.StringVar(&command, "q", "", "command or query to apply to cluster")
	flag.StringVar(&indexName, "index", "", "index name to apply command or query to")
	flag.StringVar(&docID, "id", "", "document id to get")
	flag.Parse()

	if isHelp {
		flag.Usage()
		return
	}

	esHost := os.Getenv("ESHOST")
	if esHost == "" {
		fmt.Println("ESHOST environment must be set to the URL base")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("No command provided, at least one argument required")
		os.Exit(1)
	}

	client := &http.Client{}

	switch command {
	case "ls":
		getIndicies(client, esHost)
	case "search":
		searchIndex(client, esHost, indexName)
	case "doc":
		docIndex(client, esHost, indexName, docID)
	default:
		flag.Usage()
	}
	return
}

type IndexEntries []IndexEntry

type IndexEntry struct {
	Health string
	Status string
	Index  string
}

func getIndicies(client *http.Client, esHost string) error {
	resp, err := client.Get(fmt.Sprintf("%s/_cat/indices?format=json", esHost))
	if err != nil {
		fmt.Printf("Error retrieving indices: %v\n", err)
		return err
	}
	var entries IndexEntries
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	err = dec.Decode(&entries)
	if err != nil {
		fmt.Printf("Error parsing indices: %v\n", err)
		return nil
	}

	var names []string
	for _, v := range entries {
		names = append(names, v.Index)
	}
	sort.Strings(names)

	for _, v := range names {
		fmt.Println(v)
	}

	return nil
}

type SearchResponse struct {
	Hits struct {
		Hits []map[string]interface{}
	}
}

func searchIndex(client *http.Client, esHost string, indexName string) error {
	r := strings.NewReader(`{"size": 15}`)
	resp, err := client.Post(fmt.Sprintf("%s/%s/_search?format=json", esHost, indexName), "application/json", r)
	if err != nil {
		fmt.Printf("Error searching index %s: %v\n", indexName, err)
		return err
	}

	var searchResponse SearchResponse
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()

	err = dec.Decode(&searchResponse)
	if err != nil {
		fmt.Printf("Error parsing search index <%s>: %v\n", indexName, err)
		return err
	}

	fmt.Printf("%v, %v", resp.StatusCode, len(searchResponse.Hits.Hits))
	for _, v := range searchResponse.Hits.Hits {
		b, _ := json.MarshalIndent(v, "", "  ")
		fmt.Printf("%s\n", b)
	}

	return nil
}

func docIndex(client *http.Client, esHost string, indexName string, docID string) error {
	// hacky way to work around routing :D
	r := strings.NewReader(fmt.Sprintf(`{
		"query": {
		  "match": {
			"_id": "%s"
		  }
		}
	  }`, docID))
	resp, err := client.Post(fmt.Sprintf("%s/%s/_search?format=json", esHost, indexName), "application/json", r)
	if err != nil {
		fmt.Printf("Error retrieving document <%s> from <%s>: %v\n", docID, indexName, err)
		return err
	}

	var searchResponse SearchResponse
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()

	err = dec.Decode(&searchResponse)
	if err != nil {
		fmt.Printf("Error parsing search index <%s>: %v\n", indexName, err)
		return err
	}

	for _, v := range searchResponse.Hits.Hits {
		b, _ := json.MarshalIndent(v, "", "  ")
		fmt.Printf("%s\n", b)
	}

	return nil
}
