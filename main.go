package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorCyan   = "\033[36m"
)

type DiffResult struct {
	Type  string `json:"type"`
	Field string `json:"field"`
}

var (
	jsonDiff []DiffResult
	silent   bool
)

func fetchJSON(url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 status: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}
	return data, nil
}

func readJSONFromFile(path string) (map[string]interface{}, error) {
	dataBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON in file: %v", err)
	}
	return data, nil
}

func loadJSON(input string) (map[string]interface{}, error) {
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return fetchJSON(input)
	}
	return readJSONFromFile(input)
}

func diffValues(key string, v1, v2 interface{}, prefix string) {
	fullKey := prefix + key
	switch v1Typed := v1.(type) {
	case map[string]interface{}:
		v2Typed, ok := v2.(map[string]interface{})
		if ok {
			diffMaps(v1Typed, v2Typed, fullKey+".")
		} else {
			fmt.Printf("%s🟡 Changed (type): %s%s%s\n", colorYellow, prefix, key, colorReset)
			jsonDiff = append(jsonDiff, DiffResult{Type: "changed", Field: fullKey})
		}
	default:
		if !reflect.DeepEqual(v1, v2) {
			fmt.Printf("%s🟡 Changed: %s%s%s\n", colorYellow, prefix, key, colorReset)
			jsonDiff = append(jsonDiff, DiffResult{Type: "changed", Field: fullKey})
		} else if !silent {
			fmt.Printf("%s✅ Same: %s%s%s\n", colorGreen, prefix, key, colorReset)
			jsonDiff = append(jsonDiff, DiffResult{Type: "same", Field: fullKey})
		}
	}
}

func diffMaps(m1, m2 map[string]interface{}, prefix string) {
	visited := make(map[string]bool)

	for k, v1 := range m1 {
		visited[k] = true
		if v2, ok := m2[k]; ok {
			diffValues(k, v1, v2, prefix)
		} else {
			fmt.Printf("%s🔻 Removed: %s%s%s\n", colorRed, prefix, k, colorReset)
			jsonDiff = append(jsonDiff, DiffResult{Type: "removed", Field: prefix + k})
		}
	}

	for k := range m2 {
		if !visited[k] {
			fmt.Printf("%s🔺 Added: %s%s%s\n", colorCyan, prefix, k, colorReset)
			jsonDiff = append(jsonDiff, DiffResult{Type: "added", Field: prefix + k})
		}
	}
}

func saveDiffToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(jsonDiff, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

func main() {
	var jsonOutFile string

	var rootCmd = &cobra.Command{
		Use:   "json-diff <input1> <input2>",
		Short: "Compare two JSON inputs (URLs or file paths) and show the diff",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			input1 := args[0]
			input2 := args[1]

			fmt.Println(colorCyan + "📥 Loading JSON from: " + input1 + colorReset)
			json1, err := loadJSON(input1)
			if err != nil {
				fmt.Println("❌ Error:", err)
				return
			}

			fmt.Println(colorCyan + "📥 Loading JSON from: " + input2 + colorReset)
			json2, err := loadJSON(input2)
			if err != nil {
				fmt.Println("❌ Error:", err)
				return
			}

			fmt.Println(colorCyan + "\n🔍 Comparing Responses...\n" + colorReset)
			diffMaps(json1, json2, "")

			if jsonOutFile != "" {
				if err := saveDiffToFile(jsonOutFile); err != nil {
					fmt.Println("❌ Failed to save diff:", err)
				} else {
					fmt.Println(colorGreen + "✅ JSON diff saved to: " + jsonOutFile + colorReset)
				}
			}
		},
	}

	rootCmd.Flags().StringVarP(&jsonOutFile, "json-out", "j", "", "Write diff output to JSON file")
	rootCmd.Flags().BoolVarP(&silent, "silent", "s", false, "Suppress output for unchanged  fields")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("❌ Error:", err)
		os.Exit(1)
	}
}
