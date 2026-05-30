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
	// 1. コマンドライン引数の数を検証
	if len(os.Args) < 2 {
		fmt.Println("使用方法:")
		fmt.Println("  go run main.go <入力ファイルパス> [出力ファイルパス]")
		fmt.Println("\n実行例:")
		fmt.Println("  go run main.go assets/homego.txt")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	var outputFile string

	// 出力ファイルパスの決定
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	} else {
		ext := filepath.Ext(inputFile)
		outputFile = strings.TrimSuffix(inputFile, ext) + ".html"
	}

	// 2. 変換元ファイルの読み込み
	codeBytes, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("エラー: 入力ファイル '%s' の読み込みに失敗しました: %v\n", inputFile, err)
		os.Exit(1)
	}

	// 3. Chroma構文解析ツールのセットアップ
	lexer := lexers.Get("go")
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get("monokai")

	formatter := html.New(
		html.WithClasses(true),
		html.TabWidth(4),
	)

	// 4. トークン化とHTMLへのパースを実行
	iterator, err := lexer.Tokenise(nil, string(codeBytes))
	if err != nil {
		fmt.Printf("エラー: コードの解析に失敗しました: %v\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		fmt.Printf("エラー: HTMLへの変換処理に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 💡 5. 生成されたHTML内の「{」と「}」を安全な文字実体参照に置換
	htmlStr := buf.String()
	htmlStr = strings.ReplaceAll(htmlStr, "{", "&#123;")
	htmlStr = strings.ReplaceAll(htmlStr, "}", "&#125;")

	// 6. 出力先フォルダーの自動作成とファイルの書き出し
	err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
	if err != nil {
		fmt.Printf("エラー: 出力先ディレクトリの作成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(outputFile, []byte(htmlStr), 0644)
	if err != nil {
		fmt.Printf("エラー: HTMLファイルの書き出しに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✨ 変換完了 (波括弧エスケープ済): %s -> %s\n", inputFile, outputFile)
}
