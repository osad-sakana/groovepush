package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use: "restore",
	Short: "Restore the project files from a specific commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("commit ID is required")
		}
		commitID := args[0]

		// commits.jsonから該当コミットを探す
		archivePath, err := findArchiveByCommitID(commitID)
		if err != nil {
			return err
		}
		if archivePath == "" {
			return fmt.Errorf("commit not found: %s", commitID)
		}

		// zipファイルを解凍して現在のフォルダに上書き
		err = unzipToCurrent(archivePath)
		if err != nil {
			return err
		}

		fmt.Printf("Restored to commit: %s\n", commitID)
		return nil
	},
}

func init(){
	rootCmd.AddCommand(restoreCmd)
}


func findArchiveByCommitID(commitID string) (string, error) {
	commitsFile := ".groovepush/commits.json"
	data, err := os.ReadFile(commitsFile)
	if err != nil {
		return "", err
	}

	commitsData, err := parseCommitsJson(data)
	if err != nil {
		return "", err
	}
	commits := commitsData["commits"].([]interface{})

	for _, c := range commits {
		commitObj := c.(map[string]interface{})
		if commitObj["id"] == commitID {
			return commitObj["archivePath"].(string), nil
		}
	}

	return "", nil
}

func unzipToCurrent(archivePath string) error {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		// ディレクトリかどうかチェック
        fPath := filepath.Join(".", f.Name)

        // セキュリティ対策: zip のエントリに .. などが含まれていてもルートを飛び出さないようにする
        if !strings.HasPrefix(filepath.Clean(fPath), ".") {
            return fmt.Errorf("illegal file path: %s", fPath)
        }

        if f.FileInfo().IsDir() {
            // ディレクトリの場合
            if err := os.MkdirAll(fPath, f.Mode()); err != nil {
                return err
            }
            continue
        }

        // ファイルの場合
        if err := os.MkdirAll(filepath.Dir(fPath), 0755); err != nil {
            return err
        }
        dstFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            return err
        }

        srcFile, err := f.Open()
        if err != nil {
            dstFile.Close()
            return err
        }

        _, err = io.Copy(dstFile, srcFile)

        // クローズ
        dstFile.Close()
        srcFile.Close()

        if err != nil {
            return err
        }
	}

	return nil
}