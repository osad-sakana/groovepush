package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use: "init",
	Short: "Initialize groovepush in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		gpDir := ".groovepush"

		// すでに.groovepushディレクトリが存在する場合はエラーにする
		if _, err := os.Stat(gpDir); !os.IsNotExist(err) {
			return fmt.Errorf("groovepush is already initialized in this directory")
		}

		// .groovepushディレクトリを作成する
		if err := os.Mkdir(gpDir, 0755); err != nil {
			return err
		}

		// config.jsonとcommits.jsonを初期化する
		configPath := filepath.Join(gpDir, "config.json")
		commitPath := filepath.Join(gpDir, "commits.json")

		configData := fmt.Sprintf(`{
			"projectName": "my-project",
			"initializedAt": "%s"
		}`, time.Now().Format(time.RFC3339))

		// config.json
		if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
			return err
		}

		// commits.json
		emptyCommits := `{"commits": []}`
		if err := os.WriteFile(commitPath, []byte(emptyCommits), 0644); err != nil {
			return err
		}

		fmt.Println("groovepush initialized in this directory")
		return nil
	},
}

func init(){
	rootCmd.AddCommand(initCmd)
}