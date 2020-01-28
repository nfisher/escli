package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {

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
	command := os.Args[1]
	switch command {
	case "ls":
		getIndicies(client, esHost)
	case "search":
		searchIndex(client, esHost, os.Args[2])
	}
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

	for _, v := range entries {
		fmt.Println(v.Index)
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
	resp, err := client.Post(fmt.Sprintf("%s/%s/%s", esHost, indexName, "_search?format=json"), "application/json", r)
	if err != nil {
		fmt.Printf("Error searching index %s: %v\n", indexName, err)
		return err
	}

	var searchResponse SearchResponse
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()

	err = dec.Decode(&searchResponse)
	if err != nil {
		fmt.Printf("Error parsing search index %s: %v\n", indexName, err)
		return err
	}

	fmt.Printf("%v, %v", resp.StatusCode, len(searchResponse.Hits.Hits))
	for _, v := range searchResponse.Hits.Hits {
		b, _ := json.MarshalIndent(v, "", "  ")
		fmt.Printf("%s\n", b)
	}

	return nil
}
