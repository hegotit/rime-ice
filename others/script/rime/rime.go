package rime

import (
	"bufio"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// 一个词的组成部分
type lemma struct {
	text   string // 汉字
	code   string // 编码
	weight int    // 权重
}

// rule4EngEntry 定义了英文词条处理规则，包括词条过滤和字符剪切：
// filterType: 词条过滤选项
//  1. 仅匹配原本为全小写的词条
//  2. 仅首字母转为小写（例如：'Windows XP' 转为 'windows xP'）
//  3. 将所有字母转为小写
//
// cutOffType: 字符剪切选项
//  1. 去除所有空格
//  2. 去除所有连字符"-"
//  3. 同时去除空格和连字符"-"
type rule4EngEntry struct {
	filterType int // 词条筛选规则
	cutOffType int // 字符剪切规则
}

var (
	mark    = "# +_+"      // 词库中的标记符号，表示从这行开始进行检查或排序
	RimeDir = getRimeDir() // Rime 配置目录

	EmojiMapPath = filepath.Join(RimeDir, "others/emoji-map.txt")
	EmojiPath    = filepath.Join(RimeDir, "opencc/emoji.txt")

	HanziPath    = filepath.Join(RimeDir, "cn_dicts/8105.dict.yaml")
	BasePath     = filepath.Join(RimeDir, "cn_dicts/base.dict.yaml")
	ExtPath      = filepath.Join(RimeDir, "cn_dicts/ext.dict.yaml")
	TencentPath  = filepath.Join(RimeDir, "cn_dicts/tencent.dict.yaml")
	EnPath       = filepath.Join(RimeDir, "en_dicts/en.dict.yaml")
	EnExtPath    = filepath.Join(RimeDir, "en_dicts/en_ext.dict.yaml")
	EnProperPath = filepath.Join(RimeDir, "en_dicts/en_proper.dict.yaml")
	AHDPath      = filepath.Join(RimeDir, "en_dicts/AHD5_mix_split.yaml")

	rule4En    = rule4EngEntry{3, 3}
	rule4EnExt = rule4EngEntry{0, 3}
	rule4AHD   = rule4EngEntry{4, 3}

	HanziSet   = readToSet(HanziPath)
	BaseSet    = readToSet(BasePath)
	ExtSet     = readToSet(ExtPath)
	TencentSet = readToSet(TencentPath)
	EnSet      = readToSet4Eng(EnPath, rule4En)
	EnExtSet   = readToSet4Eng(EnExtPath, rule4EnExt)
	AHDMap     = readToMap4Eng(AHDPath, rule4AHD)
	AHDMap2    = readToMap4Eng2(AHDPath, rule4AHD)
	AHDSet     = readToSet4Eng(AHDPath, rule4AHD)

	需要注音TXT   = filepath.Join(RimeDir, "others/script/rime/需要注音.txt")
	错别字TXT    = filepath.Join(RimeDir, "others/script/rime/错别字.txt")
	汉字拼音映射TXT = filepath.Join(RimeDir, "others/script/rime/汉字拼音映射.txt")
)

// 将所有词库读入 set，供检查或排序使用
func readToSet(dictPath string) mapset.Set[string] {
	set := mapset.NewSet[string]()

	file, err := os.Open(dictPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	isMark := false
	for sc.Scan() {
		line := sc.Text()
		if !isMark {
			if strings.HasPrefix(line, mark) {
				isMark = true
			}
			continue
		}
		parts := strings.Split(line, "\t")
		set.Add(parts[0])
	}

	return set
}

// 将所有词库读入 set，供检查或排序使用
func readToSet4Eng(dictPath string, rule rule4EngEntry) mapset.Set[string] {
	set := mapset.NewSet[string]()
	set2 := mapset.NewSet[string]()

	file, err := os.Open(dictPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	isMark := false
	for sc.Scan() {
		line := sc.Text()
		if !isMark {
			if strings.HasPrefix(line, mark) {
				isMark = true
			}
			continue
		}
		if rule.filterType == 1 && !isLowercase(line) {
			continue
		}

		parts := strings.Split(line, "\t")

		text, code := parts[0], strings.ToLower(parts[1])

		if rule.filterType == 4 {
			if !containsCapital(line) {
				set.Add(getKey4EngSet(text, code, rule))
			} else {
				set2.Add(getKey4EngSet(text, code, rule))
			}
			continue
		}

		set.Add(getKey4EngSet(text, code, rule))
	}

	if rule.filterType == 4 {
		return set.Difference(set2)
	}

	return set
}

func readToMap4Eng2(dictPath string, rule rule4EngEntry) map[string]interface{} {
	set := mapset.NewSet[string]()
	set2 := mapset.NewSet[string]()
	entries := make(map[string]interface{})

	file, err := os.Open(dictPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	isMark := false
	for sc.Scan() {
		line := sc.Text()
		if !isMark {
			if strings.HasPrefix(line, mark) {
				isMark = true
			}
			continue
		}
		if rule.filterType == 1 && !isLowercase(line) {
			continue
		}

		parts := strings.Split(line, "\t")

		text, code := parts[0], strings.ToLower(parts[1])
		key := getKey4EngSet(text, code, rule)

		if rule.filterType == 4 {
			if !containsCapital(line) {
				set.Add(key)
				entries[key] = lemma{text: parts[0], code: parts[1]}
			} else {
				set2.Add(key)
			}
			continue
		}

		set.Add(key)
	}

	if rule.filterType == 4 {
		return entries
	}

	return nil
}

func readToMap4Eng(dictPath string, rule rule4EngEntry) map[string]mapset.Set[interface{}] {
	entries := make(map[string]mapset.Set[interface{}])

	file, err := os.Open(dictPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	isMark := false
	for sc.Scan() {
		line := sc.Text()
		if !isMark {
			if strings.HasPrefix(line, mark) {
				isMark = true
			}
			continue
		}
		if rule.filterType == 1 && !isLowercase(line) {
			continue
		}
		parts := strings.Split(line, "\t")

		text, code := parts[0], strings.ToLower(parts[1])
		key := getKey4EngSet(text, code, rule)
		if _, exists := entries[key]; !exists {
			entries[key] = mapset.NewSet[interface{}]()
		}
		entries[key].Add(lemma{text: parts[0], code: parts[1]})
	}
	filtered := make(map[string]mapset.Set[interface{}])
	for key, set := range entries {
		if set.Cardinality() >= 2 || hasCapitalStart(set) {
			filtered[key] = set
		}
	}
	return filtered
}

func hasCapitalStart(set mapset.Set[interface{}]) bool {
	for elem := range set.Iter() {
		if l, ok := elem.(lemma); ok {
			if len(l.text) > 0 && unicode.IsUpper(rune(l.text[0])) {
				return true
			}
		}
	}
	return false
}

func getKey4EngSet(text string, code string, rule rule4EngEntry) string {
	if rule.filterType == 2 {
		text = convertFirstLetterToLower(text)
	} else if rule.filterType == 3 {
		text = strings.ToLower(text)
	} else if rule.filterType == 4 {
		text = strings.ToLower(text)
	}

	code = strings.ToLower(code)

	return applyCutOff(text+code, rule.cutOffType)
}

func applyCutOff(text string, cutOffType int) string {
	switch cutOffType {
	case 1:
		return strings.ReplaceAll(text, " ", "")
	case 2:
		return strings.ReplaceAll(text, "-", "")
	case 3:
		text = strings.ReplaceAll(text, " ", "")
		return strings.ReplaceAll(text, "-", "")
	default:
		return text
	}
}

// 首字母大写的单词，仅将首字母转为小写
func convertFirstLetterToLower(s string) string {
	var builder strings.Builder
	words := strings.Fields(s)
	for _, word := range words {
		if len(word) > 0 && unicode.IsUpper(rune(word[0])) {
			builder.WriteRune(unicode.ToLower(rune(word[0])))
			if len(word) > 1 {
				builder.WriteString(word[1:])
			}
		} else {
			builder.WriteString(word)
		}
		builder.WriteRune(' ')
	}
	result := builder.String()
	return strings.TrimSpace(result)
}

// 检查是否全为小写
func isLowercase(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

func containsCapital(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

// 打印耗时时间
func printlnTimeCost(content string, start time.Time) {
	// fmt.Printf("%s：\t%.2fs\n", content, time.Since(start).Seconds())
	printfTimeCost(content, start)
	fmt.Println()
}

// 打印耗时时间
func printfTimeCost(content string, start time.Time) {
	fmt.Printf("%s：\t%.2fs", content, time.Since(start).Seconds())
}

// slice 是否包含 item
func contains(arr []string, item string) bool {
	for _, x := range arr {
		if item == x {
			return true
		}
	}
	return false
}

// AddWeight  为 ext、tencent 没权重的词条加上权重，有权重的改为 weight
func AddWeight(dictPath string, weight int) {
	// 控制台输出
	printlnTimeCost("加权重\t"+path.Base(dictPath), time.Now())

	// 读取到 lines 数组
	file, err := os.ReadFile(dictPath)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(file), "\n")

	isMark := false
	for i, line := range lines {
		if !isMark {
			if strings.HasPrefix(line, mark) {
				isMark = true
			}
			continue
		}
		// 过滤空行
		if line == "" {
			continue
		}
		// 修改权重为传入的 weight，没有就加上
		parts := strings.Split(line, "\t")
		_, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			lines[i] = line + "\t" + strconv.Itoa(weight)
		} else {
			lines[i] = strings.Join(parts[:len(parts)-1], "\t") + "\t" + strconv.Itoa(weight)
		}
	}

	// 写入
	resultString := strings.Join(lines, "\n")
	err = os.WriteFile(dictPath, []byte(resultString), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
