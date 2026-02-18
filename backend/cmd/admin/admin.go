package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin/debug commands (requires admin key)",
}

// --- Daily Challenge Admin ---

var adminSetStreakCmd = &cobra.Command{
	Use:   "set-streak",
	Short: "Set player streak values",
	Run: func(cmd *cobra.Command, args []string) {
		body := map[string]interface{}{
			"playerId": playerFlag,
		}

		current, _ := cmd.Flags().GetInt("current")
		if cmd.Flags().Changed("current") {
			body["currentStreak"] = current
		}
		best, _ := cmd.Flags().GetInt("best")
		if cmd.Flags().Changed("best") {
			body["bestStreak"] = best
		}
		lastPlayed, _ := cmd.Flags().GetString("last-played-date")
		if lastPlayed != "" {
			body["lastPlayedDate"] = lastPlayed
		}

		data, err := apiPatch("/admin/daily-challenge/streak", body)
		printResult(data, err)
	},
}

var adminSimulateStreakCmd = &cobra.Command{
	Use:   "simulate-streak",
	Short: "Simulate N consecutive days of play",
	Run: func(cmd *cobra.Command, args []string) {
		days, _ := cmd.Flags().GetInt("days")
		score, _ := cmd.Flags().GetInt("score")

		body := map[string]interface{}{
			"playerId":  playerFlag,
			"days":      days,
			"baseScore": score,
		}
		data, err := apiPost("/admin/daily-challenge/simulate-streak", body)
		printResult(data, err)
	},
}

var adminResetPlayerCmd = &cobra.Command{
	Use:   "reset-player",
	Short: "Reset ALL data for a player (daily, marathon, quiz sessions, stats)",
	Run: func(cmd *cobra.Command, args []string) {
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("This will delete ALL data for player %s. Use --yes to confirm.\n", playerFlag)
			return
		}

		data, err := apiDelete(fmt.Sprintf("/admin/player/reset?playerId=%s", playerFlag))
		printResult(data, err)
	},
}

var adminResetDailyCmd = &cobra.Command{
	Use:   "reset-daily",
	Short: "Delete daily challenge games for a player",
	Run: func(cmd *cobra.Command, args []string) {
		date, _ := cmd.Flags().GetString("date")

		url := fmt.Sprintf("/admin/daily-challenge/games?playerId=%s", playerFlag)
		if date != "" {
			url += "&date=" + date
		}

		data, err := apiDelete(url)
		printResult(data, err)
	},
}

var adminViewDailyCmd = &cobra.Command{
	Use:   "view-daily",
	Short: "List daily challenge games for a player",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")

		data, err := apiGet(fmt.Sprintf("/admin/daily-challenge/games?playerId=%s&limit=%d", playerFlag, limit))
		printResult(data, err)
	},
}

// --- Marathon Admin ---

var adminMarathonUpdateCmd = &cobra.Command{
	Use:   "marathon-update",
	Short: "Update marathon game state (lives, bonuses)",
	Run: func(cmd *cobra.Command, args []string) {
		body := map[string]interface{}{
			"playerId": playerFlag,
		}

		if marathonFlag != "" {
			body["gameId"] = marathonFlag
		}

		// Direct set fields
		if cmd.Flags().Changed("lives") {
			v, _ := cmd.Flags().GetInt("lives")
			body["lives"] = v
		}
		if cmd.Flags().Changed("shield") {
			v, _ := cmd.Flags().GetInt("shield")
			body["shield"] = v
		}
		if cmd.Flags().Changed("fifty") {
			v, _ := cmd.Flags().GetInt("fifty")
			body["fiftyFifty"] = v
		}
		if cmd.Flags().Changed("skip") {
			v, _ := cmd.Flags().GetInt("skip")
			body["skip"] = v
		}
		if cmd.Flags().Changed("freeze") {
			v, _ := cmd.Flags().GetInt("freeze")
			body["freeze"] = v
		}

		// Additive fields
		if cmd.Flags().Changed("add-lives") {
			v, _ := cmd.Flags().GetInt("add-lives")
			body["addLives"] = v
		}
		if cmd.Flags().Changed("add-shield") {
			v, _ := cmd.Flags().GetInt("add-shield")
			body["addShield"] = v
		}
		if cmd.Flags().Changed("add-fifty") {
			v, _ := cmd.Flags().GetInt("add-fifty")
			body["addFiftyFifty"] = v
		}
		if cmd.Flags().Changed("add-skip") {
			v, _ := cmd.Flags().GetInt("add-skip")
			body["addSkip"] = v
		}
		if cmd.Flags().Changed("add-freeze") {
			v, _ := cmd.Flags().GetInt("add-freeze")
			body["addFreeze"] = v
		}

		data, err := apiPatch("/admin/marathon/game", body)
		printResult(data, err)
	},
}

var adminMarathonGamesCmd = &cobra.Command{
	Use:   "marathon-games",
	Short: "List marathon games for a player",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")

		data, err := apiGet(fmt.Sprintf("/admin/marathon/games?playerId=%s&limit=%d", playerFlag, limit))
		printResult(data, err)
	},
}

var adminMarathonDeleteCmd = &cobra.Command{
	Use:   "marathon-delete",
	Short: "Delete ALL marathon games for a player",
	Run: func(cmd *cobra.Command, args []string) {
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("This will delete ALL marathon games for player %s. Use --yes to confirm.\n", playerFlag)
			return
		}

		data, err := apiDelete(fmt.Sprintf("/admin/marathon/games?playerId=%s", playerFlag))
		printResult(data, err)
	},
}

func init() {
	// set-streak flags
	adminSetStreakCmd.Flags().Int("current", 0, "Current streak value")
	adminSetStreakCmd.Flags().Int("best", 0, "Best streak value")
	adminSetStreakCmd.Flags().String("last-played-date", "", "Last played date (YYYY-MM-DD)")

	// simulate-streak flags
	adminSimulateStreakCmd.Flags().Int("days", 7, "Number of days to simulate")
	adminSimulateStreakCmd.Flags().Int("score", 50, "Base score per game")
	adminSimulateStreakCmd.MarkFlagRequired("days")

	// reset-player flags
	adminResetPlayerCmd.Flags().Bool("yes", false, "Confirm destructive operation")

	// reset-daily flags
	adminResetDailyCmd.Flags().String("date", "", "Date (YYYY-MM-DD), empty for all dates")

	// view-daily flags
	adminViewDailyCmd.Flags().Int("limit", 20, "Number of results")

	// marathon-update flags
	adminMarathonUpdateCmd.Flags().Int("lives", 0, "Set lives to exact value")
	adminMarathonUpdateCmd.Flags().Int("add-lives", 0, "Add/subtract lives")
	adminMarathonUpdateCmd.Flags().Int("shield", 0, "Set shield bonus count")
	adminMarathonUpdateCmd.Flags().Int("fifty", 0, "Set 50/50 bonus count")
	adminMarathonUpdateCmd.Flags().Int("skip", 0, "Set skip bonus count")
	adminMarathonUpdateCmd.Flags().Int("freeze", 0, "Set freeze bonus count")
	adminMarathonUpdateCmd.Flags().Int("add-shield", 0, "Add shield bonuses")
	adminMarathonUpdateCmd.Flags().Int("add-fifty", 0, "Add 50/50 bonuses")
	adminMarathonUpdateCmd.Flags().Int("add-skip", 0, "Add skip bonuses")
	adminMarathonUpdateCmd.Flags().Int("add-freeze", 0, "Add freeze bonuses")

	// marathon-games flags
	adminMarathonGamesCmd.Flags().Int("limit", 20, "Number of results")

	// marathon-delete flags
	adminMarathonDeleteCmd.Flags().Bool("yes", false, "Confirm destructive operation")

	adminCmd.AddCommand(
		adminSetStreakCmd,
		adminSimulateStreakCmd,
		adminResetPlayerCmd,
		adminResetDailyCmd,
		adminViewDailyCmd,
		adminMarathonUpdateCmd,
		adminMarathonGamesCmd,
		adminMarathonDeleteCmd,
	)
	rootCmd.AddCommand(adminCmd)
}
