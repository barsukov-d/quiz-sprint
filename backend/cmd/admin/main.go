package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	envFlag      string
	playerFlag   string
	adminKeyFlag string
	gameIDFlag   string
	marathonFlag string

	baseURL  string
	adminKey string
)

var rootCmd = &cobra.Command{
	Use:   "quiz-admin",
	Short: "Quiz Sprint Admin CLI",
	Long:  "Admin CLI for testing Quiz Sprint API endpoints (daily challenge, marathon, admin).\nRun without subcommands for interactive menu mode.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		resolveEnv()
	},
	Run: func(cmd *cobra.Command, args []string) {
		runInteractive()
	},
}

func resolveEnv() {
	switch envFlag {
	case "local":
		baseURL = "http://localhost:3000/api/v1"
		if adminKey == "" {
			adminKey = "dev-admin-key-2026"
		}
	case "tunnel":
		baseURL = "https://dev.quiz-sprint-tma.online/api/v1"
		if adminKey == "" {
			adminKey = "dev-admin-key-2026"
		}
	case "staging":
		baseURL = "https://staging.quiz-sprint-tma.online/api/v1"
		if adminKey == "" {
			adminKey = os.Getenv("STAGING_ADMIN_API_KEY")
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown env: %s (use local, tunnel, staging)\n", envFlag)
		os.Exit(1)
	}

	if adminKeyFlag != "" {
		adminKey = adminKeyFlag
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&envFlag, "env", "e", "local", "Environment: local, tunnel, staging")
	rootCmd.PersistentFlags().StringVarP(&playerFlag, "player", "p", "1121083057", "Player ID")
	rootCmd.PersistentFlags().StringVar(&adminKeyFlag, "admin-key", "", "Admin API key (overrides env default)")
	rootCmd.PersistentFlags().StringVarP(&gameIDFlag, "game-id", "g", "", "Daily challenge game ID")
	rootCmd.PersistentFlags().StringVarP(&marathonFlag, "marathon-id", "m", "", "Marathon game ID")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
