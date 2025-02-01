package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use: "log",
	Short: "Show local commit history",
	RunE: func(cmd *cobra.Command, args []string) error {
		commitsFile := ".groovepush/commits.json"
        data, err := os.ReadFile(commitsFile)
        if err != nil {
            return err
        }

        commitsData, err := parseCommitsJson(data)
        if err != nil {
            return err
        }
        commits := commitsData["commits"].([]interface{})

		fmt.Println("-------------------------------------")
		fmt.Println("Commit History")
		fmt.Println("-------------------------------------")
        for _, c := range commits {
            commitObj := c.(map[string]interface{})
            fmt.Printf("Commit ID: %v\n", commitObj["id"])
            fmt.Printf("Message:   %v\n", commitObj["message"])
            fmt.Printf("Created:   %v\n", commitObj["createdAt"])
            fmt.Println("-------------------------------------")
        }

        return nil
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}