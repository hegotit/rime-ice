package rime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Entry represents the JSON structure where each key is a potential entry.
type Entry map[string]interface{}

var (
	inputFilePath  = "D:\\SoftwareInstalled\\MDictPC\\doc\\out.json"
	outputFilePath = "D:\\Projects\\Others\\rime-ice\\en_dicts\\output_words.txt"
)

func ExtractWords() {
	// Specify your JSON file path and the output file path.

	// Open the JSON file.
	file, err := os.Open(inputFilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create an output file.
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	// Create a buffered scanner to read the file line by line.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Each line is assumed to be a complete JSON object.
		var entry Entry
		line := scanner.Text()

		// Skip comments and empty lines.
		if strings.HasPrefix(line, "##") || strings.TrimSpace(line) == "" {
			continue
		}

		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			panic(err) // Handle errors that might occur during JSON parsing.
		}

		// Process each key in the map (each entry).
		for key := range entry {
			// Split the key on '|' and write each part to the output file.
			words := strings.Split(key, "|")
			for _, word := range words {
				trimmedWord := strings.TrimSpace(word)
				if trimmedWord != "" {
					outputFile.WriteString(trimmedWord + "\n")
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func ExtractWords2() {
	// 打开JSON文件
	file, err := os.Open(inputFilePath)
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return
	}
	defer file.Close()

	// 创建输出文件
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	// 创建JSON解码器
	decoder := json.NewDecoder(file)

	// 读取并解析JSON
	for {
		var obj map[string]interface{}
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break // 文件结束
			}
			fmt.Println("Error decoding JSON:", err)
			return
		}

		// 处理解析后的JSON对象
		for key, _ := range obj {
			if strings.HasPrefix(key, "#") || strings.TrimSpace(key) == "" {
				continue
			}
			words := strings.Split(key, "|")
			for _, word := range words {
				trimmedWord := strings.TrimSpace(word)
				if trimmedWord != "" {
					outputFile.WriteString(trimmedWord + "\n") // 写入到文件
				}
			}
		}
	}
}
