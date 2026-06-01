package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

var wsSpanRe = regexp.MustCompile(`<span class="w">([^<]*)</span>`)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run main.go <入力ファイルパス> [出力ファイルパス]")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	var outputFile string

	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	} else {
		baseName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
		outputFile = filepath.Join("../docs/code", baseName+".templ")
	}

	if _, err := os.Stat(outputFile); err == nil {
		fmt.Printf("ファイルが既に存在します: %s\n上書きしますか？ [y/n]: ", outputFile)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" {
			fmt.Println("キャンセルしました。")
			os.Exit(0)
		}
	}

	codeBytes, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}

	lexer := lexers.Get("go")
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get("monokai")

	formatter := html.New(
		html.WithClasses(true),
		html.PreventSurroundingPre(true),
		html.TabWidth(4),
	)

	iterator, err := lexer.Tokenise(nil, string(codeBytes))
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}

	htmlStr := buf.String()

	// strip whitespace span wrappers to keep each source line on one line
	htmlStr = wsSpanRe.ReplaceAllString(htmlStr, "$1")

	// escape for templ
	htmlStr = strings.ReplaceAll(htmlStr, "{", "&#123;")
	htmlStr = strings.ReplaceAll(htmlStr, "}", "&#125;")
	htmlStr = strings.TrimRight(htmlStr, "\n")

	packageName := filepath.Base(filepath.Dir(outputFile))
	baseName := strings.TrimSuffix(filepath.Base(outputFile), filepath.Ext(outputFile))
	funcName := strings.ToUpper(baseName[:1]) + baseName[1:]

	finalContent := fmt.Sprintf(
		"package %s\n\ntempl %s() {\n\t@templ.Raw(\n\t\t`\n\t\t\t<pre class=\"chroma\"><code>%s</code></pre>\n\t\t`,\n\t)\n}\n",
		packageName,
		funcName,
		htmlStr,
	)

	err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(outputFile, []byte(finalContent), 0644)
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✨ 変換完了: %s -> %s\n", inputFile, outputFile)
}
