package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// --- ANSI color helpers ---

func header(text string) {
	fmt.Printf("\033[1;36m╔════════════════════════════════════════╗\033[0m\n")
	fmt.Printf("\033[1;36m║  %-37s ║\033[0m\n", text)
	fmt.Printf("\033[1;36m╚════════════════════════════════════════╝\033[0m\n")
}

func info(text string) {
	fmt.Printf("\033[1;34m▸ %s\033[0m\n", text)
}

func ok(text string) {
	fmt.Printf("\033[1;32m✅ %s\033[0m\n", text)
}

func warn(text string) {
	fmt.Printf("\033[1;33m⚠️  %s\033[0m\n", text)
}

func errMsg(text string) {
	fmt.Printf("\033[1;31m❌ %s\033[0m\n", text)
}

// --- Input helpers ---

var scanner *bufio.Scanner

func promptInput(prompt, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return defaultVal
	}
	return input
}

func waitEnter() {
	fmt.Print("\nPress Enter to continue...")
	scanner.Scan()
}

func parseIntDefault(s string, def int) int {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	if err != nil {
		return def
	}
	return v
}

// --- Status & helpers ---

func envDisplayName() string {
	switch envFlag {
	case "local":
		return "Dev (local)"
	case "tunnel":
		return "Dev (tunnel)"
	case "staging":
		return "Staging"
	default:
		return envFlag
	}
}

func showEnv() {
	dcGame := gameIDFlag
	if dcGame == "" {
		dcGame = "none"
	}
	mGame := marathonFlag
	if mGame == "" {
		mGame = "none"
	}
	fmt.Printf("\033[0;90mENV: %s | Player: %s | DC Game: %s | Marathon: %s\033[0m\n\n",
		envDisplayName(), playerFlag, dcGame, mGame)
}

func extractGameID(data []byte) string {
	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return ""
	}

	d, ok := resp["data"].(map[string]interface{})
	if !ok {
		return ""
	}

	// Try data.game.id
	if game, ok := d["game"].(map[string]interface{}); ok {
		if id, ok := game["id"].(string); ok {
			return id
		}
	}

	// Try data.id
	if id, ok := d["id"].(string); ok {
		return id
	}

	return ""
}

// execAndShow prints the API response without os.Exit on error.
func execAndShow(data []byte, err error) {
	if err != nil {
		errMsg(fmt.Sprintf("Request failed: %v", err))
		return
	}
	printJSON(data)
}

// --- Environment Menu ---

func menuEnvironment() {
	header("Admin CLI")
	fmt.Println()
	fmt.Println("Select environment:")
	fmt.Println("  1) Dev (local)")
	fmt.Println("  2) Dev (tunnel)")
	fmt.Println("  3) Staging")

	choice := promptInput("Environment", "1")
	switch choice {
	case "2":
		envFlag = "tunnel"
	case "3":
		envFlag = "staging"
	default:
		envFlag = "local"
	}
	// Reset adminKey so resolveEnv() picks the correct default for the new environment
	adminKey = ""
	resolveEnv()

	// Prompt for admin key if staging and not set via env var
	if envFlag == "staging" && adminKey == "" {
		key := promptInput("Admin API key", "")
		if key != "" {
			adminKey = key
		}
	}

	fmt.Println()
	ok(fmt.Sprintf("Environment: %s", envDisplayName()))
	fmt.Println()
}

// --- Main Menu ---

func menuMain() {
	for {
		header("Main Menu")
		showEnv()
		fmt.Println("  1) Daily Challenge")
		fmt.Println("  2) Marathon")
		fmt.Println("  3) Admin")
		fmt.Println("  4) Settings")
		fmt.Println("  0) Exit")
		fmt.Println()

		choice := promptInput("Choice", "")
		fmt.Println()

		switch choice {
		case "1":
			menuDaily()
		case "2":
			menuMarathon()
		case "3":
			menuAdmin()
		case "4":
			menuSettings()
		case "0":
			fmt.Println("Bye!")
			return
		default:
			warn("Invalid choice")
		}
	}
}

// --- Daily Challenge Menu ---

func menuDaily() {
	for {
		header("Daily Challenge")
		showEnv()
		fmt.Println("  1) Start game")
		fmt.Println("  2) Answer question")
		fmt.Println("  3) Open chest")
		fmt.Println("  4) Retry")
		fmt.Println("  5) Status")
		fmt.Println("  6) Streak")
		fmt.Println("  7) Leaderboard")
		fmt.Println("  0) Back")
		fmt.Println()

		choice := promptInput("Choice", "")
		fmt.Println()

		switch choice {
		case "1":
			iDailyStart()
		case "2":
			iDailyAnswer()
		case "3":
			iDailyChest()
		case "4":
			iDailyRetry()
		case "5":
			iDailyStatus()
		case "6":
			iDailyStreak()
		case "7":
			iDailyLeaderboard()
		case "0":
			return
		default:
			warn("Invalid choice")
			continue
		}
		waitEnter()
	}
}

func iDailyStart() {
	info("Starting Daily Challenge...")
	body := map[string]interface{}{"playerId": playerFlag}
	data, err := apiPost("/daily-challenge/start", body)
	execAndShow(data, err)
	if err == nil {
		if gid := extractGameID(data); gid != "" {
			gameIDFlag = gid
			ok(fmt.Sprintf("Game started! ID: %s", gid))
		}
	}
}

func iDailyAnswer() {
	gid := gameIDFlag
	if gid == "" {
		gid = promptInput("Game ID", "")
		if gid == "" {
			errMsg("Game ID is required")
			return
		}
	}
	qid := promptInput("Question ID", "")
	if qid == "" {
		errMsg("Question ID is required")
		return
	}
	aid := promptInput("Answer ID", "")
	if aid == "" {
		errMsg("Answer ID is required")
		return
	}
	timeTaken := promptInput("Time taken (seconds)", "5")

	body := map[string]interface{}{
		"playerId":   playerFlag,
		"questionId": qid,
		"answerId":   aid,
		"timeTaken":  parseIntDefault(timeTaken, 5),
	}
	data, err := apiPost(fmt.Sprintf("/daily-challenge/%s/answer", gid), body)
	execAndShow(data, err)
}

func iDailyChest() {
	gid := gameIDFlag
	if gid == "" {
		gid = promptInput("Game ID", "")
		if gid == "" {
			errMsg("Game ID is required")
			return
		}
	}
	info("Opening chest...")
	body := map[string]interface{}{"playerId": playerFlag}
	data, err := apiPost(fmt.Sprintf("/daily-challenge/%s/chest/open", gid), body)
	execAndShow(data, err)
}

func iDailyRetry() {
	gid := gameIDFlag
	if gid == "" {
		gid = promptInput("Game ID", "")
		if gid == "" {
			errMsg("Game ID is required")
			return
		}
	}
	method := promptInput("Payment method (ad/coins)", "ad")
	body := map[string]interface{}{
		"playerId":      playerFlag,
		"paymentMethod": method,
	}
	data, err := apiPost(fmt.Sprintf("/daily-challenge/%s/retry", gid), body)
	execAndShow(data, err)
	if err == nil {
		if newID := extractGameID(data); newID != "" {
			gameIDFlag = newID
			ok(fmt.Sprintf("New game ID: %s", newID))
		}
	}
}

func iDailyStatus() {
	info("Getting daily challenge status...")
	data, err := apiGet(fmt.Sprintf("/daily-challenge/status?playerId=%s", playerFlag))
	execAndShow(data, err)
}

func iDailyStreak() {
	info("Getting streak info...")
	data, err := apiGet(fmt.Sprintf("/daily-challenge/streak?playerId=%s", playerFlag))
	execAndShow(data, err)
}

func iDailyLeaderboard() {
	limit := promptInput("Limit", "10")
	data, err := apiGet(fmt.Sprintf("/daily-challenge/leaderboard?limit=%s", limit))
	execAndShow(data, err)
}

// --- Marathon Menu ---

func menuMarathon() {
	for {
		header("Marathon")
		showEnv()
		fmt.Println("  1) Start game")
		fmt.Println("  2) Answer question")
		fmt.Println("  3) Use bonus")
		fmt.Println("  4) Continue (after game over)")
		fmt.Println("  5) Abandon")
		fmt.Println("  6) Status")
		fmt.Println("  7) Personal bests")
		fmt.Println("  8) Leaderboard")
		fmt.Println("  0) Back")
		fmt.Println()

		choice := promptInput("Choice", "")
		fmt.Println()

		switch choice {
		case "1":
			iMarathonStart()
		case "2":
			iMarathonAnswer()
		case "3":
			iMarathonBonus()
		case "4":
			iMarathonContinue()
		case "5":
			iMarathonAbandon()
		case "6":
			iMarathonStatus()
		case "7":
			iMarathonPB()
		case "8":
			iMarathonLeaderboard()
		case "0":
			return
		default:
			warn("Invalid choice")
			continue
		}
		waitEnter()
	}
}

func getMarathonID() string {
	mid := marathonFlag
	if mid == "" {
		mid = promptInput("Marathon ID", "")
	}
	return mid
}

func iMarathonStart() {
	categoryID := promptInput("Category ID (empty for all)", "")
	body := map[string]interface{}{"playerId": playerFlag}
	if categoryID != "" {
		body["categoryId"] = categoryID
	}
	info("Starting Marathon...")
	data, err := apiPost("/marathon/start", body)
	execAndShow(data, err)
	if err == nil {
		if gid := extractGameID(data); gid != "" {
			marathonFlag = gid
			ok(fmt.Sprintf("Marathon started! ID: %s", gid))
		}
	}
}

func iMarathonAnswer() {
	mid := getMarathonID()
	if mid == "" {
		errMsg("Marathon ID is required")
		return
	}
	qid := promptInput("Question ID", "")
	if qid == "" {
		errMsg("Question ID is required")
		return
	}
	aid := promptInput("Answer ID", "")
	if aid == "" {
		errMsg("Answer ID is required")
		return
	}
	timeTaken := promptInput("Time taken (seconds)", "5")

	body := map[string]interface{}{
		"playerId":   playerFlag,
		"questionId": qid,
		"answerId":   aid,
		"timeTaken":  parseIntDefault(timeTaken, 5),
	}
	data, err := apiPost(fmt.Sprintf("/marathon/%s/answer", mid), body)
	execAndShow(data, err)
}

func iMarathonBonus() {
	mid := getMarathonID()
	if mid == "" {
		errMsg("Marathon ID is required")
		return
	}
	qid := promptInput("Question ID", "")
	if qid == "" {
		errMsg("Question ID is required")
		return
	}
	fmt.Println("  Bonus types: shield, fifty_fifty, skip, freeze")
	bonusType := promptInput("Bonus type", "")
	if bonusType == "" {
		errMsg("Bonus type is required")
		return
	}

	body := map[string]interface{}{
		"playerId":   playerFlag,
		"questionId": qid,
		"bonusType":  bonusType,
	}
	data, err := apiPost(fmt.Sprintf("/marathon/%s/bonus", mid), body)
	execAndShow(data, err)
}

func iMarathonContinue() {
	mid := getMarathonID()
	if mid == "" {
		errMsg("Marathon ID is required")
		return
	}
	method := promptInput("Payment method (ad/coins)", "ad")
	body := map[string]interface{}{
		"playerId":      playerFlag,
		"paymentMethod": method,
	}
	data, err := apiPost(fmt.Sprintf("/marathon/%s/continue", mid), body)
	execAndShow(data, err)
}

func iMarathonAbandon() {
	mid := getMarathonID()
	if mid == "" {
		errMsg("Marathon ID is required")
		return
	}
	body := map[string]interface{}{"playerId": playerFlag}
	data, err := apiDeleteWithBody(fmt.Sprintf("/marathon/%s", mid), body)
	execAndShow(data, err)
}

func iMarathonStatus() {
	info("Getting marathon status...")
	data, err := apiGet(fmt.Sprintf("/marathon/status?playerId=%s", playerFlag))
	execAndShow(data, err)
}

func iMarathonPB() {
	info("Getting personal bests...")
	data, err := apiGet(fmt.Sprintf("/marathon/personal-bests?playerId=%s", playerFlag))
	execAndShow(data, err)
}

func iMarathonLeaderboard() {
	timeframe := promptInput("Time frame (all_time/weekly/daily)", "all_time")
	limit := promptInput("Limit", "10")
	data, err := apiGet(fmt.Sprintf("/marathon/leaderboard?categoryId=all&timeFrame=%s&limit=%s", timeframe, limit))
	execAndShow(data, err)
}

// --- Admin Menu ---

func menuAdmin() {
	for {
		header("Admin")
		showEnv()
		fmt.Println("  1) Set streak")
		fmt.Println("  2) Simulate streak")
		fmt.Println("  3) Reset player (ALL data)")
		fmt.Println("  4) Reset daily games")
		fmt.Println("  5) View daily games")
		fmt.Println("  6) Marathon update (lives)")
		fmt.Println("  7) Marathon set bonuses")
		fmt.Println("  8) Marathon games")
		fmt.Println("  9) Marathon delete all")
		fmt.Println("  0) Back")
		fmt.Println()

		choice := promptInput("Choice", "")
		fmt.Println()

		switch choice {
		case "1":
			iAdminSetStreak()
		case "2":
			iAdminSimulateStreak()
		case "3":
			iAdminResetPlayer()
		case "4":
			iAdminResetDaily()
		case "5":
			iAdminViewDaily()
		case "6":
			iAdminMarathonUpdate()
		case "7":
			iAdminMarathonSetBonuses()
		case "8":
			iAdminMarathonGames()
		case "9":
			iAdminMarathonDelete()
		case "0":
			return
		default:
			warn("Invalid choice")
			continue
		}
		waitEnter()
	}
}

func iAdminSetStreak() {
	body := map[string]interface{}{"playerId": playerFlag}

	current := promptInput("Current streak (empty to skip)", "")
	if current != "" {
		body["currentStreak"] = parseIntDefault(current, 0)
	}
	best := promptInput("Best streak (empty to skip)", "")
	if best != "" {
		body["bestStreak"] = parseIntDefault(best, 0)
	}
	lastPlayed := promptInput("Last played date YYYY-MM-DD (empty to skip)", "")
	if lastPlayed != "" {
		body["lastPlayedDate"] = lastPlayed
	}

	data, err := apiPatch("/admin/daily-challenge/streak", body)
	execAndShow(data, err)
}

func iAdminSimulateStreak() {
	days := promptInput("Number of days", "7")
	score := promptInput("Base score per game", "50")

	body := map[string]interface{}{
		"playerId":  playerFlag,
		"days":      parseIntDefault(days, 7),
		"baseScore": parseIntDefault(score, 50),
	}
	data, err := apiPost("/admin/daily-challenge/simulate-streak", body)
	execAndShow(data, err)
}

func iAdminResetPlayer() {
	warn(fmt.Sprintf("This will delete ALL data for player %s!", playerFlag))
	confirm := promptInput("Type 'yes' to confirm", "")
	if confirm != "yes" {
		info("Cancelled")
		return
	}
	data, err := apiDelete(fmt.Sprintf("/admin/player/reset?playerId=%s", playerFlag))
	execAndShow(data, err)
}

func iAdminResetDaily() {
	date := promptInput("Date YYYY-MM-DD (empty for all dates)", "")
	url := fmt.Sprintf("/admin/daily-challenge/games?playerId=%s", playerFlag)
	if date != "" {
		url += "&date=" + date
	}
	data, err := apiDelete(url)
	execAndShow(data, err)
}

func iAdminViewDaily() {
	limit := promptInput("Limit", "20")
	data, err := apiGet(fmt.Sprintf("/admin/daily-challenge/games?playerId=%s&limit=%s", playerFlag, limit))
	execAndShow(data, err)
}

func iAdminMarathonUpdate() {
	body := map[string]interface{}{"playerId": playerFlag}
	if marathonFlag != "" {
		body["gameId"] = marathonFlag
	}

	lives := promptInput("Set lives (empty to skip)", "")
	if lives != "" {
		body["lives"] = parseIntDefault(lives, 0)
	}
	addLives := promptInput("Add lives (empty to skip)", "")
	if addLives != "" {
		body["addLives"] = parseIntDefault(addLives, 0)
	}

	data, err := apiPatch("/admin/marathon/game", body)
	execAndShow(data, err)
}

func iAdminMarathonSetBonuses() {
	body := map[string]interface{}{"playerId": playerFlag}
	if marathonFlag != "" {
		body["gameId"] = marathonFlag
	}

	shield := promptInput("Shield count (empty to skip)", "")
	if shield != "" {
		body["shield"] = parseIntDefault(shield, 0)
	}
	fifty := promptInput("50/50 count (empty to skip)", "")
	if fifty != "" {
		body["fiftyFifty"] = parseIntDefault(fifty, 0)
	}
	skip := promptInput("Skip count (empty to skip)", "")
	if skip != "" {
		body["skip"] = parseIntDefault(skip, 0)
	}
	freeze := promptInput("Freeze count (empty to skip)", "")
	if freeze != "" {
		body["freeze"] = parseIntDefault(freeze, 0)
	}

	data, err := apiPatch("/admin/marathon/game", body)
	execAndShow(data, err)
}

func iAdminMarathonGames() {
	limit := promptInput("Limit", "20")
	data, err := apiGet(fmt.Sprintf("/admin/marathon/games?playerId=%s&limit=%s", playerFlag, limit))
	execAndShow(data, err)
}

func iAdminMarathonDelete() {
	warn(fmt.Sprintf("This will delete ALL marathon games for player %s!", playerFlag))
	confirm := promptInput("Type 'yes' to confirm", "")
	if confirm != "yes" {
		info("Cancelled")
		return
	}
	data, err := apiDelete(fmt.Sprintf("/admin/marathon/games?playerId=%s", playerFlag))
	execAndShow(data, err)
}

// --- Settings Menu ---

func menuSettings() {
	for {
		header("Settings")
		showEnv()
		fmt.Println("  1) Change environment")
		fmt.Println("  2) Change player ID")
		fmt.Println("  3) Set DC game ID")
		fmt.Println("  4) Set Marathon game ID")
		fmt.Println("  0) Back")
		fmt.Println()

		choice := promptInput("Choice", "")
		fmt.Println()

		switch choice {
		case "1":
			menuEnvironment()
		case "2":
			playerFlag = promptInput("Player ID", playerFlag)
			ok(fmt.Sprintf("Player ID: %s", playerFlag))
		case "3":
			gameIDFlag = promptInput("DC Game ID", gameIDFlag)
			ok(fmt.Sprintf("DC Game ID: %s", gameIDFlag))
		case "4":
			marathonFlag = promptInput("Marathon Game ID", marathonFlag)
			ok(fmt.Sprintf("Marathon Game ID: %s", marathonFlag))
		case "0":
			return
		default:
			warn("Invalid choice")
		}
	}
}

// --- Entry point ---

// runInteractive starts the interactive menu mode.
func runInteractive() {
	scanner = bufio.NewScanner(os.Stdin)
	menuEnvironment()
	menuMain()
}
