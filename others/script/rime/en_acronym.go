package rime

import (
	"bufio"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReadAcronym 从 others/en_acronym.txt 更新英文缩拼词
func ReadAcronym() map[string][]string {
	defer printlnTimeCost("更新英文缩拼词 ", time.Now())
	file, err := os.Open(filepath.Join(RimeDir, "others/en_acronym.txt"))
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	uniq := mapset.NewSet[string]()
	resultMap := make(map[string][]string)
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if uniq.Contains(line) {
			fmt.Println("❌ 重复", line)
			continue
		}
		uniq.Add(line)

		parts := strings.Split(line, "\t")
		if len(parts) > 1 {
			var result []string
			result = append(result, fmt.Sprintf("%s\t%s", parts[0], parts[0]))
			for _, part := range parts[1:] {
				compactPart := strings.ReplaceAll(part, " ", "")
				result = append(result, fmt.Sprintf("%s\t%s", part, parts[0]))
				result = append(result, fmt.Sprintf("%s\t%s", part, compactPart))
			}
			resultMap[parts[0]] = result
		}
	}

	return resultMap
}
