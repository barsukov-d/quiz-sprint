package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/persistence/postgres"
	"github.com/barsukov/quiz-sprint/backend/pkg/database"
)

// QuizImportData represents the JSON structure for importing a quiz (verbose format)
type QuizImportData struct {
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	CategoryID   *string          `json:"categoryId,omitempty"` // Optional category ID
	TimeLimit    int              `json:"timeLimit"`            // seconds
	PassingScore int              `json:"passingScore"`         // percentage (0-100)
	Questions    []QuestionImport `json:"questions"`
	Tags         []string         `json:"tags,omitempty"` // Optional tags
}

// QuestionImport represents a question in the import file (verbose format)
type QuestionImport struct {
	Text    string         `json:"text"`
	Points  int            `json:"points"`
	Answers []AnswerImport `json:"answers"`
}

// AnswerImport represents an answer in the import file (verbose format)
type AnswerImport struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
}

// CompactQuiz represents the compact JSON format for LLM generation
type CompactQuiz struct {
	V    *int              `json:"v,omitempty"`    // version (omit if 1)
	T    string            `json:"t"`              // title
	D    string            `json:"d"`              // description
	Cat  string            `json:"cat,omitempty"`  // category (can be inferred from tags)
	Tags []string          `json:"tags,omitempty"` // optional tags
	L    *int              `json:"l,omitempty"`    // timeLimit seconds (omit if 60)
	P    *int              `json:"p,omitempty"`    // passingScore % (omit if 70)
	Q    []CompactQuestion `json:"q"`              // questions
}

// CompactQuestion represents a question in compact format
type CompactQuestion struct {
	T string   `json:"t"`          // question text
	A []string `json:"a"`          // answers (array of strings)
	C int      `json:"c"`          // correctIndex (0-based)
	P *int     `json:"p,omitempty"` // points (omit if 10)
}

// BatchImport represents a batch of quizzes with shared metadata
type BatchImport struct {
	Batch struct {
		Version   int      `json:"version"`
		Generated string   `json:"generated,omitempty"`
		Cat       string   `json:"cat,omitempty"`  // default category
		Tags      []string `json:"tags,omitempty"` // shared tags
	} `json:"batch"`
	Quizzes []CompactQuiz `json:"quizzes"`
}

// detectFormat determines the format of the JSON file
func detectFormat(data []byte) (string, error) {
	// Try to parse as a generic map to inspect structure
	var generic map[string]interface{}
	if err := json.Unmarshal(data, &generic); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	// Check for batch format (has "batch" and "quizzes" fields)
	if _, hasBatch := generic["batch"]; hasBatch {
		if _, hasQuizzes := generic["quizzes"]; hasQuizzes {
			return "batch", nil
		}
	}

	// Check for compact format (has "t" field for title)
	if _, hasT := generic["t"]; hasT {
		return "compact", nil
	}

	// Check for verbose format (has "title" field)
	if _, hasTitle := generic["title"]; hasTitle {
		return "verbose", nil
	}

	return "", fmt.Errorf("unknown format: unable to detect quiz structure")
}

// deduplicateTags removes duplicate tags from a slice
func deduplicateTags(tags []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}
	return result
}

// inferCategoryFromTags infers a category name from tags
func inferCategoryFromTags(tags []string) string {
	// Priority 1: language tags → programming
	for _, tag := range tags {
		if strings.HasPrefix(tag, "language:") {
			return "programming"
		}
	}

	// Priority 2: domain tags
	for _, tag := range tags {
		if strings.HasPrefix(tag, "domain:") {
			domain := strings.TrimPrefix(tag, "domain:")
			return domain
		}
	}

	// Fallback
	return "general"
}

// inferCategoryIDFromName looks up category UUID by name
func inferCategoryIDFromName(db *sql.DB, categoryName string) (*quiz.CategoryID, error) {
	if categoryName == "" {
		return nil, nil
	}

	categoryRepo := postgres.NewCategoryRepository(db)

	// Find all categories and match by name (case-insensitive)
	categories, err := categoryRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	for _, cat := range categories {
		if strings.EqualFold(cat.Name().String(), categoryName) {
			id := cat.ID()
			return &id, nil
		}
	}

	return nil, fmt.Errorf("category not found: %s", categoryName)
}

// convertCompactToVerbose converts compact format to verbose format
func convertCompactToVerbose(compact CompactQuiz, batchTags []string) QuizImportData {
	// Merge tags (batch + quiz, deduplicated)
	allTags := append([]string{}, batchTags...)
	allTags = append(allTags, compact.Tags...)
	allTags = deduplicateTags(allTags)

	// Apply defaults
	timeLimit := 60
	if compact.L != nil {
		timeLimit = *compact.L
	}

	passingScore := 70
	if compact.P != nil {
		passingScore = *compact.P
	}

	// Convert questions
	questions := make([]QuestionImport, len(compact.Q))
	for i, cq := range compact.Q {
		// Default points
		points := 10
		if cq.P != nil {
			points = *cq.P
		}

		// Convert answers from index-based to boolean array
		answers := make([]AnswerImport, len(cq.A))
		for j, answerText := range cq.A {
			answers[j] = AnswerImport{
				Text:      answerText,
				IsCorrect: j == cq.C, // True if this is the correct index
			}
		}

		questions[i] = QuestionImport{
			Text:    cq.T,
			Points:  points,
			Answers: answers,
		}
	}

	return QuizImportData{
		Title:        compact.T,
		Description:  compact.D,
		CategoryID:   nil, // Will be inferred later
		TimeLimit:    timeLimit,
		PassingScore: passingScore,
		Questions:    questions,
		Tags:         allTags,
	}
}

func main() {
	// Parse command-line flags
	filePath := flag.String("file", "", "Path to JSON file to import")
	dirPath := flag.String("dir", "", "Path to directory with JSON files to import")
	dryRun := flag.Bool("dry-run", false, "Validate without importing")
	flag.Parse()

	if *filePath == "" && *dirPath == "" {
		log.Fatal("Error: You must specify either -file or -dir")
	}

	// Connect to database (skip if dry-run)
	var db *sql.DB
	var err error
	if !*dryRun {
		dbConfig := database.LoadConfigFromEnv()
		db, err = database.Connect(dbConfig)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		log.Println("✓ Connected to database")
	}

	// Determine files to import
	var files []string
	if *filePath != "" {
		files = []string{*filePath}
	} else {
		files, err = filepath.Glob(filepath.Join(*dirPath, "*.json"))
		if err != nil {
			log.Fatalf("Failed to read directory: %v", err)
		}
	}

	if len(files) == 0 {
		log.Fatal("No JSON files found to import")
	}

	log.Printf("Found %d file(s) to import\n", len(files))

	// Import each file
	successCount := 0
	errorCount := 0

	for _, file := range files {
		log.Printf("\n--- Processing: %s ---", filepath.Base(file))

		if err := importQuizFromFile(file, db, *dryRun); err != nil {
			log.Printf("✗ Error: %v", err)
			errorCount++
		} else {
			log.Printf("✓ Success")
			successCount++
		}
	}

	// Summary
	log.Printf("\n=== Import Summary ===")
	log.Printf("Total: %d", len(files))
	log.Printf("Success: %d", successCount)
	log.Printf("Errors: %d", errorCount)

	if *dryRun {
		log.Println("\n(Dry run - no data was imported)")
	}
}

// importQuizFromFile reads and imports a quiz from a JSON file
func importQuizFromFile(filePath string, db *sql.DB, dryRun bool) error {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Detect format
	format, err := detectFormat(data)
	if err != nil {
		return fmt.Errorf("format detection failed: %w", err)
	}

	log.Printf("  Format: %s", format)

	// Parse and convert based on format
	var quizzesToImport []QuizImportData
	var batchID *string
	var generatedAt *int64

	switch format {
	case "batch":
		var batch BatchImport
		if err := json.Unmarshal(data, &batch); err != nil {
			return fmt.Errorf("failed to parse batch JSON: %w", err)
		}

		// Generate batch ID from filename and timestamp
		batchIDStr := fmt.Sprintf("%s-%d", filepath.Base(filePath), time.Now().Unix())
		batchID = &batchIDStr

		// Parse generated timestamp if provided
		if batch.Batch.Generated != "" {
			t, err := time.Parse(time.RFC3339, batch.Batch.Generated)
			if err == nil {
				ts := t.Unix()
				generatedAt = &ts
			}
		}

		// Convert each quiz
		for _, compactQuiz := range batch.Quizzes {
			// Merge batch category if quiz doesn't have one
			if compactQuiz.Cat == "" {
				compactQuiz.Cat = batch.Batch.Cat
			}

			verbose := convertCompactToVerbose(compactQuiz, batch.Batch.Tags)
			quizzesToImport = append(quizzesToImport, verbose)
		}

		log.Printf("  Batch ID: %s", *batchID)
		log.Printf("  Quizzes in batch: %d", len(quizzesToImport))

	case "compact":
		var compactQuiz CompactQuiz
		if err := json.Unmarshal(data, &compactQuiz); err != nil {
			return fmt.Errorf("failed to parse compact JSON: %w", err)
		}

		verbose := convertCompactToVerbose(compactQuiz, nil)
		quizzesToImport = append(quizzesToImport, verbose)

	case "verbose":
		var importData QuizImportData
		if err := json.Unmarshal(data, &importData); err != nil {
			return fmt.Errorf("failed to parse verbose JSON: %w", err)
		}

		quizzesToImport = append(quizzesToImport, importData)

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Import each quiz
	for i, importData := range quizzesToImport {
		if len(quizzesToImport) > 1 {
			log.Printf("\n  --- Quiz %d/%d ---", i+1, len(quizzesToImport))
		}

		// Validate
		if err := validateQuizData(&importData); err != nil {
			return fmt.Errorf("validation failed for quiz %d: %w", i+1, err)
		}

		log.Printf("  Title: %s", importData.Title)
		log.Printf("  Questions: %d", len(importData.Questions))
		log.Printf("  Time Limit: %d seconds", importData.TimeLimit)
		log.Printf("  Passing Score: %d%%", importData.PassingScore)
		if len(importData.Tags) > 0 {
			log.Printf("  Tags: %s", strings.Join(importData.Tags, ", "))
		}

		// If dry-run, stop here
		if dryRun {
			continue
		}

		// Convert to domain model and save
		if err := saveQuizToDB(db, &importData, batchID, generatedAt); err != nil {
			return fmt.Errorf("failed to save quiz %d to database: %w", i+1, err)
		}
	}

	return nil
}

// validateQuizData validates the import data structure
func validateQuizData(data *QuizImportData) error {
	if data.Title == "" {
		return fmt.Errorf("title is required")
	}

	if data.TimeLimit <= 0 {
		return fmt.Errorf("timeLimit must be positive")
	}

	if data.PassingScore < 0 || data.PassingScore > 100 {
		return fmt.Errorf("passingScore must be between 0 and 100")
	}

	if len(data.Questions) == 0 {
		return fmt.Errorf("at least one question is required")
	}

	// Validate each question
	for i, q := range data.Questions {
		if q.Text == "" {
			return fmt.Errorf("question %d: text is required", i+1)
		}

		if q.Points <= 0 {
			return fmt.Errorf("question %d: points must be positive", i+1)
		}

		if len(q.Answers) < 2 {
			return fmt.Errorf("question %d: at least 2 answers required", i+1)
		}

		// Check that there's exactly one correct answer
		correctCount := 0
		for _, a := range q.Answers {
			if a.IsCorrect {
				correctCount++
			}
		}

		if correctCount != 1 {
			return fmt.Errorf("question %d: exactly one answer must be correct (found %d)", i+1, correctCount)
		}
	}

	return nil
}

// saveQuizToDB converts import data to domain model and saves to database
func saveQuizToDB(db *sql.DB, data *QuizImportData, batchID *string, generatedAt *int64) error {
	// Create repositories
	quizRepo := postgres.NewQuizRepository(db)
	tagRepo := postgres.NewTagRepository(db)

	// Convert to domain types
	title, err := quiz.NewQuizTitle(data.Title)
	if err != nil {
		return fmt.Errorf("invalid title: %w", err)
	}

	timeLimit, err := quiz.NewTimeLimit(data.TimeLimit)
	if err != nil {
		return fmt.Errorf("invalid time limit: %w", err)
	}

	passingScore, err := quiz.NewPassingScore(data.PassingScore)
	if err != nil {
		return fmt.Errorf("invalid passing score: %w", err)
	}

	// Convert category ID if provided, otherwise infer from tags
	var categoryID quiz.CategoryID
	if data.CategoryID != nil && *data.CategoryID != "" {
		cid, err := quiz.NewCategoryIDFromString(*data.CategoryID)
		if err != nil {
			return fmt.Errorf("invalid categoryId: %w", err)
		}
		categoryID = cid
	} else if len(data.Tags) > 0 {
		// Infer category from tags
		categoryName := inferCategoryFromTags(data.Tags)
		log.Printf("  Inferred category: %s (from tags)", categoryName)

		cid, err := inferCategoryIDFromName(db, categoryName)
		if err != nil {
			log.Printf("  Warning: %v (quiz will have no category)", err)
			// Continue without category - categoryID remains zero value
		} else if cid != nil {
			categoryID = *cid
		}
	}

	// Create Quiz aggregate
	createdAt := int64(0) // Will be set by database
	quizAggregate, err := quiz.NewQuiz(
		quiz.NewQuizID(), // Generate new ID
		title,
		data.Description,
		categoryID,
		timeLimit,
		passingScore,
		createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create quiz: %w", err)
	}

	// Add tags to quiz
	if len(data.Tags) > 0 {
		tags := make([]*quiz.Tag, 0, len(data.Tags))
		for _, tagName := range data.Tags {
			tag, err := quiz.NewTag(tagName)
			if err != nil {
				return fmt.Errorf("invalid tag '%s': %w", tagName, err)
			}
			tags = append(tags, tag)

			// Add tag to quiz aggregate
			if err := quizAggregate.AddTag(*tag); err != nil {
				return fmt.Errorf("failed to add tag '%s': %w", tagName, err)
			}
		}

		// Save tags to database (creates if not exists)
		if err := tagRepo.SaveAll(tags); err != nil {
			return fmt.Errorf("failed to save tags: %w", err)
		}

		log.Printf("  Created/assigned %d tags", len(tags))
	}

	// Set import metadata if provided
	if batchID != nil && generatedAt != nil {
		quizAggregate.SetImportMetadata(*batchID, *generatedAt)
	} else if batchID != nil {
		// Only batch ID, use current timestamp
		quizAggregate.SetImportMetadata(*batchID, time.Now().Unix())
	}

	// Convert questions and add to quiz
	for questionIndex, qData := range data.Questions {
		questionText, err := quiz.NewQuestionText(qData.Text)
		if err != nil {
			return fmt.Errorf("invalid question text: %w", err)
		}

		points, err := quiz.NewPoints(qData.Points)
		if err != nil {
			return fmt.Errorf("invalid points: %w", err)
		}

		// Create question with position
		question, err := quiz.NewQuestion(
			quiz.NewQuestionID(),
			questionText,
			points,
			questionIndex, // position
		)
		if err != nil {
			return fmt.Errorf("failed to create question: %w", err)
		}

		// Convert answers and add to question
		for answerIndex, aData := range qData.Answers {
			answerText, err := quiz.NewAnswerText(aData.Text)
			if err != nil {
				return fmt.Errorf("invalid answer text: %w", err)
			}

			answer, err := quiz.NewAnswer(
				quiz.NewAnswerID(),
				answerText,
				aData.IsCorrect,
				answerIndex, // position
			)
			if err != nil {
				return fmt.Errorf("failed to create answer: %w", err)
			}

			// Add answer to question
			if err := question.AddAnswer(*answer); err != nil {
				return fmt.Errorf("failed to add answer to question: %w", err)
			}
		}

		// Add question to quiz
		if err := quizAggregate.AddQuestion(*question); err != nil {
			return fmt.Errorf("failed to add question to quiz: %w", err)
		}
	}

	// Save to database
	if err := quizRepo.Save(quizAggregate); err != nil {
		return fmt.Errorf("failed to save quiz: %w", err)
	}

	log.Printf("  Quiz ID: %s", quizAggregate.ID().String())

	return nil
}
