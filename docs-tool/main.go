package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func main() {
	// 1. 引数の数を検証
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run main.go <入力ファイルパス> [出力ファイルパス]")
		os.Exit(1)
	}

	// 💡 os.Args のインデックス 1 と 2 を、通常の半角角カッコ [ ] で正しく取得します
	inputFile := os.Args[1]
	var outputFile string

	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	} else {
		ext := filepath.Ext(inputFile)
		outputFile = strings.TrimSuffix(inputFile, ext) + ".html"
	}

	// 2. 変換元ファイルの読み込み
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

	// 💡 インラインの background を消去し、Chromaの pre 生成をスキップする正しいオプション
	formatter := html.New(
		html.WithClasses(true),
		html.PreventSurroundingPre(true),
		html.TabWidth(4),
	)

	// 3. 構文解析の実行
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

	// 4. templのエラーを防ぐエスケープ処理
	htmlStr = strings.ReplaceAll(htmlStr, "{", "&#123;")
	htmlStr = strings.ReplaceAll(htmlStr, "}", "&#125;")

	// 5. 外枠の pre.chroma で包んで結合
	finalHTML := fmt.Sprintf("<pre class=\"chroma\"><code>%s</code></pre>", htmlStr)

	err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(outputFile, []byte(finalHTML), 0644)
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✨ 変換完了: %s -> %s\n", inputFile, outputFile)
}
