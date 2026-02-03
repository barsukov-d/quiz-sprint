package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/persistence/postgres"
	"github.com/barsukov/quiz-sprint/backend/pkg/database"
)

// CompactQuiz represents the compact JSON format (same as import)
type CompactQuiz struct {
	T    string            `json:"t"`              // title
	D    string            `json:"d,omitempty"`    // description
	Cat  string            `json:"cat,omitempty"`  // category name
	Tags []string          `json:"tags,omitempty"` // tags
	L    *int              `json:"l,omitempty"`    // timeLimit (omit if 60)
	P    *int              `json:"p,omitempty"`    // passingScore (omit if 70)
	Q    []CompactQuestion `json:"q"`              // questions
}

// CompactQuestion represents a question in compact format
type CompactQuestion struct {
	T string   `json:"t"`           // question text
	A []string `json:"a"`           // answers
	C int      `json:"c"`           // correctIndex (0-based)
	P *int     `json:"p,omitempty"` // points (omit if 0)
}

// BatchExport represents a batch of quizzes
type BatchExport struct {
	Batch BatchMeta     `json:"batch"`
	Quizzes []CompactQuiz `json:"quizzes"`
}

// BatchMeta holds batch metadata
type BatchMeta struct {
	Version   int    `json:"version"`
	Generated string `json:"generated,omitempty"`
}

func main() {
	outDir := flag.String("dir", "data/quizzes/exported", "Output directory for exported JSON files")
	quizID := flag.String("id", "", "Export a single quiz by ID")
	batch := flag.Bool("batch", false, "Export all quizzes as a single batch file")
	listOnly := flag.Bool("list", false, "List all quizzes without exporting")
	flag.Parse()

	// Connect to database
	dbConfig := database.LoadConfigFromEnv()
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✓ Connected to database")

	quizRepo := postgres.NewQuizRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)

	// Build category ID → name map
	categoryMap, err := buildCategoryMap(categoryRepo)
	if err != nil {
		log.Fatalf("Failed to load categories: %v", err)
	}

	// List mode
	if *listOnly {
		listQuizzes(quizRepo)
		return
	}

	// Load quizzes
	var quizzes []*quiz.Quiz

	if *quizID != "" {
		// Single quiz
		qid, err := quiz.NewQuizIDFromString(*quizID)
		if err != nil {
			log.Fatalf("Invalid quiz ID: %v", err)
		}
		q, err := quizRepo.FindByID(qid)
		if err != nil {
			log.Fatalf("Failed to load quiz: %v", err)
		}
		quizzes = append(quizzes, q)
	} else {
		// All quizzes - load summaries first, then full quizzes
		summaries, err := quizRepo.FindAllSummaries()
		if err != nil {
			log.Fatalf("Failed to load quiz summaries: %v", err)
		}

		log.Printf("Found %d quizzes to export", len(summaries))

		for _, s := range summaries {
			q, err := quizRepo.FindByID(s.ID())
			if err != nil {
				log.Printf("✗ Failed to load quiz %s: %v", s.ID().String(), err)
				continue
			}
			quizzes = append(quizzes, q)
		}
	}

	if len(quizzes) == 0 {
		log.Println("No quizzes found to export")
		return
	}

	// Convert to compact format
	compactQuizzes := make([]CompactQuiz, 0, len(quizzes))
	for _, q := range quizzes {
		compact := convertToCompact(q, categoryMap)
		compactQuizzes = append(compactQuizzes, compact)
	}

	// Export
	if *batch {
		exportBatch(compactQuizzes, *outDir)
	} else {
		exportIndividual(compactQuizzes, *outDir)
	}
}

func buildCategoryMap(repo *postgres.CategoryRepository) (map[string]string, error) {
	categories, err := repo.FindAll()
	if err != nil {
		return nil, err
	}

	m := make(map[string]string, len(categories))
	for _, cat := range categories {
		m[cat.ID().String()] = cat.Name().String()
	}
	return m, nil
}

func listQuizzes(repo *postgres.QuizRepository) {
	summaries, err := repo.FindAllSummaries()
	if err != nil {
		log.Fatalf("Failed to load quizzes: %v", err)
	}

	log.Printf("\n=== Quizzes (%d) ===\n", len(summaries))
	for _, s := range summaries {
		log.Printf("  %s  %s  (%d questions)", s.ID().String(), s.Title().String(), s.QuestionCount())
	}
}

func convertToCompact(q *quiz.Quiz, categoryMap map[string]string) CompactQuiz {
	compact := CompactQuiz{
		T: q.Title().String(),
		D: q.Description(),
	}

	// Category
	if !q.CategoryID().IsZero() {
		if name, ok := categoryMap[q.CategoryID().String()]; ok {
			compact.Cat = name
		}
	}

	// Tags
	tagNames := q.TagNames()
	if len(tagNames) > 0 {
		compact.Tags = tagNames
	}

	// TimeLimit (omit if 60)
	tl := q.TimeLimit().Seconds()
	if tl != 60 {
		compact.L = &tl
	}

	// PassingScore (omit if 70)
	ps := q.PassingScore().Percentage()
	if ps != 70 {
		compact.P = &ps
	}

	// Questions
	questions := q.Questions()
	compact.Q = make([]CompactQuestion, 0, len(questions))
	for _, question := range questions {
		cq := CompactQuestion{
			T: question.Text().String(),
		}

		// Answers and correct index
		answers := question.Answers()
		cq.A = make([]string, 0, len(answers))
		for i, a := range answers {
			cq.A = append(cq.A, a.Text().String())
			if a.IsCorrect() {
				cq.C = i
			}
		}

		// Points (omit if 0)
		pts := question.Points().Value()
		if pts != 0 {
			cq.P = &pts
		}

		compact.Q = append(compact.Q, cq)
	}

	return compact
}

func exportBatch(quizzes []CompactQuiz, outDir string) {
	batch := BatchExport{
		Batch: BatchMeta{
			Version: 1,
		},
		Quizzes: quizzes,
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	outPath := filepath.Join(outDir, "all-quizzes.json")
	writeJSON(outPath, batch)

	log.Printf("✓ Exported %d quizzes to %s", len(quizzes), outPath)
}

func exportIndividual(quizzes []CompactQuiz, outDir string) {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	for i, q := range quizzes {
		// Generate filename from title
		filename := sanitizeFilename(q.T) + ".json"
		outPath := filepath.Join(outDir, filename)

		writeJSON(outPath, q)
		log.Printf("  [%d/%d] ✓ %s", i+1, len(quizzes), filename)
	}

	log.Printf("\n✓ Exported %d quizzes to %s/", len(quizzes), outDir)
}

func writeJSON(path string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Ensure trailing newline
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Fatalf("Failed to write file %s: %v", path, err)
	}
}

func sanitizeFilename(title string) string {
	// Lowercase, replace spaces with hyphens, remove special chars
	name := strings.ToLower(title)
	name = strings.ReplaceAll(name, " ", "-")

	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	s := result.String()
	// Collapse multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")

	if s == "" {
		s = "quiz"
	}

	return s
}
