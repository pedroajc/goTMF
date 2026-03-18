package main

import "fmt"

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
	printResult(checkNode("catalogue-eu", 120))
	printResult(checkNode("catalogue-us", 750))
	printResult(checkNode("catalogue-apac", -5))
}
