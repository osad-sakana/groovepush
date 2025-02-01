package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	// "golang.org/x/tools/go/analysis/passes/defers"
)

// 指定フォルダを再起的にzip化する(.groovepushディレクトリは除外)
func zipFolder(srcDir, dstZip string) error {
	zipFile, err := os.Create(dstZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	// walk
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// .groovepushディレクトリは除外
		if strings.Contains(path, ".groovepush") {
			return nil
		}

		// ディレクトリはスキップ
		if info.IsDir() {
			return nil
		}

		// 相対パスを取得
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// zip内に書き込むファイルを作成
		zipEntry, err := writer.Create(relPath)
		if err != nil {
			return err
		}

		// 実ファイルを読み込み
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		// zipEntryにコピー
		_, err = io.Copy(zipEntry, f)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

var commitCmd = &cobra.Command{
	Use: "commit",
	Short: "Create a snapshot (zip) of the current directory and save it as a commit",
	RunE: func(cmd *cobra.Command, args []string) error{
		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			return fmt.Errorf("commit message is requried")
		}

		commitID := uuid.New().String()
		commitsDir := ".groovepush/commits"
		if _, err := os.Stat(commitsDir); os.IsNotExist(err) {
			os.Mkdir(commitsDir, 0755)
		}

		archivePath := filepath.Join(commitsDir, commitID+".zip")

		err := zipFolder(".", archivePath)
		if err != nil {
			return err
		}

		// commit情報を保存
		err = appendCommitInfo(commitID, message, archivePath)
		if err != nil {
			return err
		}

		fmt.Println("commit created:", commitID)
		return nil
	},
}

func init(){
	commitCmd.Flags().StringP("message", "m", "", "commit message")
	rootCmd.AddCommand(commitCmd)
}

func appendCommitInfo(commitID, message, archivePath string) error {
	commitsFile := ".groovepush/commits.json"

	// ファイル読み込み
	data, err := os.ReadFile(commitsFile)
	if err != nil {
		return err
	}

	// JSONパース
	commitsData, err := parseCommitsJson(data)
	if err != nil{
		return err
	}

	// commits配列に新しい要素を追加
	newCommit := map[string]interface{}{
		"id": commitID,
		"message": message,
		"archive": archivePath,
		"createdAt": time.Now().Format(time.RFC3339),
	}
	commits := commitsData["commits"].([]interface{})
	commits = append(commits, newCommit)
	commitsData["commits"] = commits

	// JSONエンコード
	newJson, err := toJsonBytes(commitsData)
	if err != nil {
		return err
	}
	err = os.WriteFile(commitsFile, newJson, 0644)
	if err != nil {
		return err
	}

	return nil
}

func parseCommitsJson(data []byte) (map[string]interface{}, error) {
    var result map[string]interface{}
    err := json.Unmarshal(data, &result)
    return result, err
}

func toJsonBytes(v interface{}) ([]byte, error) {
    return json.MarshalIndent(v, "", "  ")
}