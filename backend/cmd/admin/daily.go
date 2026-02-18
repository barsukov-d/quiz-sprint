package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var dailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Daily Challenge commands",
}

var dailyStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a daily challenge game",
	Run: func(cmd *cobra.Command, args []string) {
		body := map[string]interface{}{
			"playerId": playerFlag,
		}
		data, err := apiPost("/daily-challenge/start", body)
		printResult(data, err)
	},
}

var dailyAnswerCmd = &cobra.Command{
	Use:   "answer",
	Short: "Submit an answer to a daily challenge question",
	Run: func(cmd *cobra.Command, args []string) {
		gid := requireGameID()
		qid, _ := cmd.Flags().GetString("question-id")
		aid, _ := cmd.Flags().GetString("answer-id")
		timeTaken, _ := cmd.Flags().GetInt("time")

		body := map[string]interface{}{
			"playerId":   playerFlag,
			"questionId": qid,
			"answerId":   aid,
			"timeTaken":  timeTaken,
		}
		data, err := apiPost(fmt.Sprintf("/daily-challenge/%s/answer", gid), body)
		printResult(data, err)
	},
}

var dailyChestCmd = &cobra.Command{
	Use:   "chest",
	Short: "Open chest after completing daily challenge",
	Run: func(cmd *cobra.Command, args []string) {
		gid := requireGameID()
		body := map[string]interface{}{
			"playerId": playerFlag,
		}
		data, err := apiPost(fmt.Sprintf("/daily-challenge/%s/chest/open", gid), body)
		printResult(data, err)
	},
}

var dailyRetryCmd = &cobra.Command{
	Use:   "retry",
	Short: "Retry a daily challenge game",
	Run: func(cmd *cobra.Command, args []string) {
		gid := requireGameID()
		method, _ := cmd.Flags().GetString("method")

		body := map[string]interface{}{
			"playerId":      playerFlag,
			"paymentMethod": method,
		}
		data, err := apiPost(fmt.Sprintf("/daily-challenge/%s/retry", gid), body)
		printResult(data, err)
	},
}

var dailyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get daily challenge status for a player",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := apiGet(fmt.Sprintf("/daily-challenge/status?playerId=%s", playerFlag))
		printResult(data, err)
	},
}

var dailyStreakCmd = &cobra.Command{
	Use:   "streak",
	Short: "Get player streak info",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := apiGet(fmt.Sprintf("/daily-challenge/streak?playerId=%s", playerFlag))
		printResult(data, err)
	},
}

var dailyLeaderboardCmd = &cobra.Command{
	Use:   "leaderboard",
	Short: "Get daily challenge leaderboard",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		data, err := apiGet(fmt.Sprintf("/daily-challenge/leaderboard?limit=%d", limit))
		printResult(data, err)
	},
}

// requireGameID returns the game ID from flag or exits.
func requireGameID() string {
	if gameIDFlag != "" {
		return gameIDFlag
	}
	fmt.Fprintln(os.Stderr, "Error: --game-id (-g) is required")
	fmt.Fprintln(os.Stderr, "Tip: start a game first, then pass the game ID")
	os.Exit(1)
	return "" // unreachable
}

func init() {
	// daily answer flags
	dailyAnswerCmd.Flags().String("question-id", "", "Question ID (required)")
	dailyAnswerCmd.Flags().String("answer-id", "", "Answer ID (required)")
	dailyAnswerCmd.Flags().Int("time", 5, "Time taken in seconds")
	dailyAnswerCmd.MarkFlagRequired("question-id")
	dailyAnswerCmd.MarkFlagRequired("answer-id")

	// daily retry flags
	dailyRetryCmd.Flags().String("method", "ad", "Payment method: ad or coins")

	// daily leaderboard flags
	dailyLeaderboardCmd.Flags().Int("limit", 10, "Number of results")

	dailyCmd.AddCommand(
		dailyStartCmd,
		dailyAnswerCmd,
		dailyChestCmd,
		dailyRetryCmd,
		dailyStatusCmd,
		dailyStreakCmd,
		dailyLeaderboardCmd,
	)
	rootCmd.AddCommand(dailyCmd)
}
