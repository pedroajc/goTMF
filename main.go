package main

import (
	"encoding/json"
	"fmt"
)

func checkNode(nodeName string, latencyMs int) (string, error) {
	if latencyMs < 0 {
		return "", fmt.Errorf("invalid latency for node %s", nodeName)
	}
	if latencyMs > 500 {
		return "", fmt.Errorf("node %s is too slow: %dms", nodeName, latencyMs)
	}
	return fmt.Sprintf("node %s is healthy (%dms)", nodeName, latencyMs), nil
}

func printResult(message string, err error) {
	if err != nil {
		fmt.Printf("WARN: %v\n", err)
		return
	}
	fmt.Println(message)
}

func main() {
	/*
	   // Lesson 1
	   printResult(checkNode("catalogue-eu", 120))
	   printResult(checkNode("catalogue-us", 750))
	   printResult(checkNode("catalogue-apac", -5))
	*/

	// Lesson 2
	cat := Catalog{
		ID:              "cat-001",
		Name:            "B2B Catalogue",
		LifecycleStatus: "Active",
		AtType:          "Catalog",
		ValidFor:        &TimePeriod{StartDateTime: "2025-01-01T00:00:00Z"},
	}
	catBytes, catErr := json.MarshalIndent(cat, "", "  ")
	printResult(string(catBytes), catErr)
	f := false
	offer := ProductOffering{
		ID:              "po-001",
		Name:            "Fibre 1Gbps",
		LifecycleStatus: "Active",
		AtType:          "ProductOffering",
		IsBundle:        &f}

	offerBytes, offerErr := json.MarshalIndent(offer, "", "  ")
	printResult(string(offerBytes), offerErr)

	var incoming Catalog
	err := json.Unmarshal(catBytes, &incoming)
	if err != nil {
		fmt.Printf("ERROR: unmarshalling catalog: %v\n", err)
		return
	}

	fmt.Printf("Unmarshalled catalog name: %s\n", incoming.Name)
	fmt.Printf("Unmarshalled catalog status: %s\n", incoming.LifecycleStatus)

}
