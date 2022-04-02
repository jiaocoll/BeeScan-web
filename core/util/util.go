package util

import (
	"fmt"
	"github.com/karrick/godirwalk"
	"log"
	"math/rand"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

/*
创建人员：云深不知处
创建时间：2022/1/3
程序功能：工具包
*/

const numbers string = "0123456789"
const letters string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const specials = "~!@#$%^*()_+-=[]{}|;:,./<>?"
const alphanumberic string = letters + numbers
const ascii string = alphanumberic + specials

func Random(n int, chars string) string {
	if n <= 0 {
		return ""
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, n, n)
	l := len(chars)
	for i := 0; i < n; i++ {
		bytes[i] = chars[r.Intn(l)]
	}
	return string(bytes)
}

func RandomAlphanumeric(n int) string {
	return Random(n, alphanumberic)
}

func RandomAlphabetic(n int) string {
	return Random(n, letters)
}

func RandomNumeric(n int) string {
	return Random(n, numbers)
}

func RandomAscii(n int) string {
	return Random(n, ascii)
}

// TrimProtocol removes the HTTP scheme from an URI
func TrimProtocol(targetURL string) string {
	URL := strings.TrimSpace(targetURL)
	if strings.HasPrefix(strings.ToLower(URL), "http://") || strings.HasPrefix(strings.ToLower(URL), "https://") {
		URL = URL[strings.Index(URL, "//")+2:]
	}
	URL = strings.TrimRight(URL, "/")
	return URL
}

// Removesamesip 去重函数
func Removesamesip(ips []string) (result []string) {
	result = make([]string, 0)
	tempMap := make(map[string]bool, len(ips))
	for _, e := range ips {
		if tempMap[e] == false {
			tempMap[e] = true
			result = append(result, e)
		}
	}
	return result
}

func In(target string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	if index < len(strArray) && strArray[index] == target {
		return true
	}
	return false
}

func RandomStr(randSource *rand.Rand, letterBytes string, n int) string {
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
		//letterBytes   = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	)
	randBytes := make([]byte, n)
	for i, cache, remain := n-1, randSource.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randSource.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			randBytes[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(randBytes)
}

func DirectoryWalker(fsPath string, callback func(fsPath string, d *godirwalk.Dirent) error) error {
	err := godirwalk.Walk(fsPath, &godirwalk.Options{
		Callback: callback,
		ErrorCallback: func(fsPath string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
		Unsorted: true,
	})

	// directory couldn't be walked
	if err != nil {
		return err
	}

	return nil
}

func IsFilePath(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	return info.Mode().IsRegular(), nil
}

func ResolvePath(templateName string, TemplatesDirectory string) (string, error) {
	curDirectory, err := os.Getwd()
	if err != nil {
		return "", err
	}

	templatePath := path.Join(curDirectory, templateName)
	if _, err := os.Stat(templatePath); !os.IsNotExist(err) {
		log.Printf("Found template in current directory: %s\n", templatePath)

		return templatePath, nil
	}

	if TemplatesDirectory != "" {
		templatePath := path.Join(TemplatesDirectory, templateName)
		if _, err := os.Stat(templatePath); !os.IsNotExist(err) {
			log.Printf("Found template in nuclei-templates directory: %s\n", templatePath)

			return templatePath, nil
		}
	}

	return "", fmt.Errorf("no such path found: %s", templateName)
}

func IsRelative(filePath string) bool {
	if strings.HasPrefix(filePath, "/") || strings.Contains(filePath, ":\\") {
		return false
	}

	return true
}

func ResolvePathIfRelative(f string) (string, error) {
	var absPath string
	var err error
	if IsRelative(f) {
		absPath, err = ResolvePath(f, "")
		if err != nil {
			return "", err
		}
	} else {
		absPath = f
	}
	return absPath, nil
}
