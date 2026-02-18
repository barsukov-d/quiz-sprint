package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var marathonCmd = &cobra.Command{
	Use:   "marathon",
	Short: "Marathon commands",
}

var marathonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a marathon game",
	Run: func(cmd *cobra.Command, args []string) {
		categoryID, _ := cmd.Flags().GetString("category")

		body := map[string]interface{}{
			"playerId": playerFlag,
		}
		if categoryID != "" {
			body["categoryId"] = categoryID
		}
		data, err := apiPost("/marathon/start", body)
		printResult(data, err)
	},
}

var marathonAnswerCmd = &cobra.Command{
	Use:   "answer",
	Short: "Submit an answer to a marathon question",
	Run: func(cmd *cobra.Command, args []string) {
		mid := requireMarathonID()
		qid, _ := cmd.Flags().GetString("question-id")
		aid, _ := cmd.Flags().GetString("answer-id")
		timeTaken, _ := cmd.Flags().GetInt("time")

		body := map[string]interface{}{
			"playerId":   playerFlag,
			"questionId": qid,
			"answerId":   aid,
			"timeTaken":  timeTaken,
		}
		data, err := apiPost(fmt.Sprintf("/marathon/%s/answer", mid), body)
		printResult(data, err)
	},
}

var marathonBonusCmd = &cobra.Command{
	Use:   "bonus",
	Short: "Use a bonus in a marathon game",
	Run: func(cmd *cobra.Command, args []string) {
		mid := requireMarathonID()
		qid, _ := cmd.Flags().GetString("question-id")
		bonusType, _ := cmd.Flags().GetString("type")

		body := map[string]interface{}{
			"playerId":   playerFlag,
			"questionId": qid,
			"bonusType":  bonusType,
		}
		data, err := apiPost(fmt.Sprintf("/marathon/%s/bonus", mid), body)
		printResult(data, err)
	},
}

var marathonContinueCmd = &cobra.Command{
	Use:   "continue",
	Short: "Continue a marathon game after game over",
	Run: func(cmd *cobra.Command, args []string) {
		mid := requireMarathonID()
		method, _ := cmd.Flags().GetString("method")

		body := map[string]interface{}{
			"playerId":      playerFlag,
			"paymentMethod": method,
		}
		data, err := apiPost(fmt.Sprintf("/marathon/%s/continue", mid), body)
		printResult(data, err)
	},
}

var marathonAbandonCmd = &cobra.Command{
	Use:   "abandon",
	Short: "Abandon a marathon game",
	Run: func(cmd *cobra.Command, args []string) {
		mid := requireMarathonID()
		body := map[string]interface{}{
			"playerId": playerFlag,
		}
		data, err := apiDeleteWithBody(fmt.Sprintf("/marathon/%s", mid), body)
		printResult(data, err)
	},
}

var marathonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get marathon status for a player",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := apiGet(fmt.Sprintf("/marathon/status?playerId=%s", playerFlag))
		printResult(data, err)
	},
}

var marathonPBCmd = &cobra.Command{
	Use:   "pb",
	Short: "Get marathon personal bests",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := apiGet(fmt.Sprintf("/marathon/personal-bests?playerId=%s", playerFlag))
		printResult(data, err)
	},
}

var marathonLeaderboardCmd = &cobra.Command{
	Use:   "leaderboard",
	Short: "Get marathon leaderboard",
	Run: func(cmd *cobra.Command, args []string) {
		timeframe, _ := cmd.Flags().GetString("timeframe")
		limit, _ := cmd.Flags().GetInt("limit")

		data, err := apiGet(fmt.Sprintf("/marathon/leaderboard?categoryId=all&timeFrame=%s&limit=%d", timeframe, limit))
		printResult(data, err)
	},
}

// requireMarathonID returns the marathon ID from flag or exits.
func requireMarathonID() string {
	if marathonFlag != "" {
		return marathonFlag
	}
	fmt.Fprintln(os.Stderr, "Error: --marathon-id (-m) is required")
	fmt.Fprintln(os.Stderr, "Tip: start a marathon first, then pass the marathon ID")
	os.Exit(1)
	return "" // unreachable
}

func init() {
	// marathon start flags
	marathonStartCmd.Flags().String("category", "", "Category ID (empty for all)")

	// marathon answer flags
	marathonAnswerCmd.Flags().String("question-id", "", "Question ID (required)")
	marathonAnswerCmd.Flags().String("answer-id", "", "Answer ID (required)")
	marathonAnswerCmd.Flags().Int("time", 5, "Time taken in seconds")
	marathonAnswerCmd.MarkFlagRequired("question-id")
	marathonAnswerCmd.MarkFlagRequired("answer-id")

	// marathon bonus flags
	marathonBonusCmd.Flags().String("question-id", "", "Question ID (required)")
	marathonBonusCmd.Flags().String("type", "", "Bonus type: shield, fifty_fifty, skip, freeze (required)")
	marathonBonusCmd.MarkFlagRequired("question-id")
	marathonBonusCmd.MarkFlagRequired("type")

	// marathon continue flags
	marathonContinueCmd.Flags().String("method", "ad", "Payment method: ad or coins")

	// marathon leaderboard flags
	marathonLeaderboardCmd.Flags().String("timeframe", "all_time", "Time frame: all_time, weekly, daily")
	marathonLeaderboardCmd.Flags().Int("limit", 10, "Number of results")

	marathonCmd.AddCommand(
		marathonStartCmd,
		marathonAnswerCmd,
		marathonBonusCmd,
		marathonContinueCmd,
		marathonAbandonCmd,
		marathonStatusCmd,
		marathonPBCmd,
		marathonLeaderboardCmd,
	)
	rootCmd.AddCommand(marathonCmd)
}
