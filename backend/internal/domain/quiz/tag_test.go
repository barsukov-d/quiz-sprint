package quiz

import (
	"testing"
)

func TestNewTagName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError error
	}{
		{
			name:      "Valid language tag",
			input:     "language:go",
			wantError: nil,
		},
		{
			name:      "Valid difficulty tag",
			input:     "difficulty:easy",
			wantError: nil,
		},
		{
			name:      "Valid topic tag with hyphen",
			input:     "topic:web-development",
			wantError: nil,
		},
		{
			name:      "Valid domain tag",
			input:     "domain:programming",
			wantError: nil,
		},
		{
			name:      "Valid format tag",
			input:     "format:multiple-choice",
			wantError: nil,
		},
		{
			name:      "Empty string",
			input:     "",
			wantError: ErrEmptyTagName,
		},
		{
			name:      "Too long",
			input:     "language:" + string(make([]byte, 100)),
			wantError: ErrTagNameTooLong,
		},
		{
			name:      "Contains space",
			input:     "language:go programming",
			wantError: ErrTagNameHasSpaces,
		},
		{
			name:      "Contains uppercase",
			input:     "Language:Go",
			wantError: ErrTagNameHasUppercase,
		},
		{
			name:      "Missing colon",
			input:     "languagego",
			wantError: ErrTagMissingColon,
		},
		{
			name:      "Invalid category",
			input:     "invalid:go",
			wantError: ErrInvalidTagCategory,
		},
		{
			name:      "Invalid characters",
			input:     "language:go@programming",
			wantError: ErrInvalidTagFormat,
		},
		{
			name:      "Only colon",
			input:     ":",
			wantError: ErrInvalidTagFormat,
		},
		{
			name:      "Multiple hyphens",
			input:     "topic:web-api-development",
			wantError: nil,
		},
		{
			name:      "Numbers in value",
			input:     "topic:es6-features",
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTagName(tt.input)
			if err != tt.wantError {
				t.Errorf("NewTagName(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
		})
	}
}

func TestNewTag(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError error
	}{
		{
			name:      "Valid tag",
			input:     "language:go",
			wantError: nil,
		},
		{
			name:      "Invalid tag name",
			input:     "invalid format",
			wantError: ErrTagNameHasSpaces,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := NewTag(tt.input)
			if err != tt.wantError {
				t.Errorf("NewTag(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
			if err == nil && tag == nil {
				t.Error("NewTag returned nil tag with nil error")
			}
		})
	}
}

func TestTagName_Category(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		expected string
	}{
		{
			name:     "Language tag",
			tagName:  "language:go",
			expected: "language",
		},
		{
			name:     "Difficulty tag",
			tagName:  "difficulty:easy",
			expected: "difficulty",
		},
		{
			name:     "Topic tag",
			tagName:  "topic:concurrency",
			expected: "topic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagName, err := NewTagName(tt.tagName)
			if err != nil {
				t.Fatalf("NewTagName(%q) failed: %v", tt.tagName, err)
			}

			category := tagName.Category()
			if category != tt.expected {
				t.Errorf("Category() = %q, want %q", category, tt.expected)
			}
		})
	}
}

func TestTagName_Value(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		expected string
	}{
		{
			name:     "Language tag",
			tagName:  "language:go",
			expected: "go",
		},
		{
			name:     "Difficulty tag",
			tagName:  "difficulty:easy",
			expected: "easy",
		},
		{
			name:     "Topic with hyphen",
			tagName:  "topic:web-development",
			expected: "web-development",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagName, err := NewTagName(tt.tagName)
			if err != nil {
				t.Fatalf("NewTagName(%q) failed: %v", tt.tagName, err)
			}

			value := tagName.Value()
			if value != tt.expected {
				t.Errorf("Value() = %q, want %q", value, tt.expected)
			}
		})
	}
}

func TestTag_Equals(t *testing.T) {
	tag1, _ := NewTag("language:go")
	tag2, _ := NewTag("language:go")
	tag3, _ := NewTag("language:python")

	if !tag1.Equals(tag2) {
		t.Error("Expected tag1 to equal tag2")
	}

	if tag1.Equals(tag3) {
		t.Error("Expected tag1 to not equal tag3")
	}

	if tag1.Equals(nil) {
		t.Error("Expected tag1 to not equal nil")
	}
}

func TestInferCategoryFromTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{
			name:     "Language tag → programming",
			tags:     []string{"language:go", "difficulty:easy"},
			expected: "programming",
		},
		{
			name:     "Domain history tag → history",
			tags:     []string{"domain:history", "difficulty:medium"},
			expected: "history",
		},
		{
			name:     "Domain science tag → science",
			tags:     []string{"domain:science", "topic:astronomy"},
			expected: "science",
		},
		{
			name:     "No matching tags → general",
			tags:     []string{"difficulty:easy", "topic:basics"},
			expected: "general",
		},
		{
			name:     "Language has priority over domain",
			tags:     []string{"language:go", "domain:history"},
			expected: "programming",
		},
		{
			name:     "Empty tags → general",
			tags:     []string{},
			expected: "general",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferCategoryFromTags(tt.tags)
			if result != tt.expected {
				t.Errorf("InferCategoryFromTags(%v) = %q, want %q", tt.tags, result, tt.expected)
			}
		})
	}
}

func TestNewTagID(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		expected string
	}{
		{
			name:     "Language tag",
			tagName:  "language:go",
			expected: "language-go",
		},
		{
			name:     "Difficulty tag",
			tagName:  "difficulty:easy",
			expected: "difficulty-easy",
		},
		{
			name:     "Multi-word with hyphens",
			tagName:  "topic:web-development",
			expected: "topic-web-development",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagID := NewTagID(tt.tagName)
			if tagID.String() != tt.expected {
				t.Errorf("NewTagID(%q).String() = %q, want %q", tt.tagName, tagID.String(), tt.expected)
			}
		})
	}
}

func TestIsValidTagCategory(t *testing.T) {
	tests := []struct {
		category string
		expected bool
	}{
		{"language", true},
		{"difficulty", true},
		{"topic", true},
		{"domain", true},
		{"format", true},
		{"invalid", false},
		{"Language", false}, // case-sensitive
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			result := IsValidTagCategory(tt.category)
			if result != tt.expected {
				t.Errorf("IsValidTagCategory(%q) = %v, want %v", tt.category, result, tt.expected)
			}
		})
	}
}

func TestGetValidTagCategories(t *testing.T) {
	categories := GetValidTagCategories()

	if len(categories) != 5 {
		t.Errorf("Expected 5 valid categories, got %d", len(categories))
	}

	expectedCategories := map[string]bool{
		"language":   true,
		"difficulty": true,
		"topic":      true,
		"domain":     true,
		"format":     true,
	}

	for _, category := range categories {
		if !expectedCategories[category] {
			t.Errorf("Unexpected category: %s", category)
		}
	}
}

func TestReconstructTag(t *testing.T) {
	tag := ReconstructTag("language-go", "language:go")

	if tag == nil {
		t.Fatal("ReconstructTag returned nil")
	}

	if tag.ID().String() != "language-go" {
		t.Errorf("Expected ID 'language-go', got %q", tag.ID().String())
	}

	if tag.Name().String() != "language:go" {
		t.Errorf("Expected name 'language:go', got %q", tag.Name().String())
	}
}

// Test Quiz methods with tags

func TestQuiz_AddTag(t *testing.T) {
	quiz, err := NewQuiz(
		NewQuizID(),
		MustNewQuizTitle("Test Quiz"),
		"Description",
		CategoryID{},
		MustNewTimeLimit(60),
		MustNewPassingScore(70),
		0,
	)
	if err != nil {
		t.Fatalf("Failed to create quiz: %v", err)
	}

	tag1, _ := NewTag("language:go")
	tag2, _ := NewTag("difficulty:easy")
	tag3, _ := NewTag("language:go") // Duplicate

	// Add first tag
	err = quiz.AddTag(*tag1)
	if err != nil {
		t.Errorf("AddTag failed: %v", err)
	}

	if len(quiz.Tags()) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(quiz.Tags()))
	}

	// Add second tag
	err = quiz.AddTag(*tag2)
	if err != nil {
		t.Errorf("AddTag failed: %v", err)
	}

	if len(quiz.Tags()) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(quiz.Tags()))
	}

	// Add duplicate tag (should be idempotent)
	err = quiz.AddTag(*tag3)
	if err != nil {
		t.Errorf("AddTag with duplicate should not error: %v", err)
	}

	if len(quiz.Tags()) != 2 {
		t.Errorf("Expected 2 tags (duplicate should be ignored), got %d", len(quiz.Tags()))
	}

	// Test maximum tags limit
	for i := 0; i < 10; i++ {
		tag, _ := NewTag("topic:test-" + string(rune('a'+i)))
		quiz.AddTag(*tag)
	}

	tagOver, _ := NewTag("topic:over-limit")
	err = quiz.AddTag(*tagOver)
	if err != ErrTooManyTags {
		t.Errorf("Expected ErrTooManyTags, got %v", err)
	}
}

func TestQuiz_HasTag(t *testing.T) {
	quiz, _ := NewQuiz(
		NewQuizID(),
		MustNewQuizTitle("Test Quiz"),
		"Description",
		CategoryID{},
		MustNewTimeLimit(60),
		MustNewPassingScore(70),
		0,
	)

	tag, _ := NewTag("language:go")
	quiz.AddTag(*tag)

	if !quiz.HasTag("language:go") {
		t.Error("Expected quiz to have tag 'language:go'")
	}

	if quiz.HasTag("language:python") {
		t.Error("Expected quiz to not have tag 'language:python'")
	}
}

func TestQuiz_HasTagCategory(t *testing.T) {
	quiz, _ := NewQuiz(
		NewQuizID(),
		MustNewQuizTitle("Test Quiz"),
		"Description",
		CategoryID{},
		MustNewTimeLimit(60),
		MustNewPassingScore(70),
		0,
	)

	tag1, _ := NewTag("language:go")
	tag2, _ := NewTag("topic:concurrency")
	quiz.AddTag(*tag1)
	quiz.AddTag(*tag2)

	if !quiz.HasTagCategory("language") {
		t.Error("Expected quiz to have language category")
	}

	if !quiz.HasTagCategory("topic") {
		t.Error("Expected quiz to have topic category")
	}

	if quiz.HasTagCategory("difficulty") {
		t.Error("Expected quiz to not have difficulty category")
	}
}

func TestQuiz_TagNames(t *testing.T) {
	quiz, _ := NewQuiz(
		NewQuizID(),
		MustNewQuizTitle("Test Quiz"),
		"Description",
		CategoryID{},
		MustNewTimeLimit(60),
		MustNewPassingScore(70),
		0,
	)

	tag1, _ := NewTag("language:go")
	tag2, _ := NewTag("difficulty:easy")
	quiz.AddTag(*tag1)
	quiz.AddTag(*tag2)

	tagNames := quiz.TagNames()
	if len(tagNames) != 2 {
		t.Errorf("Expected 2 tag names, got %d", len(tagNames))
	}

	expectedNames := map[string]bool{
		"language:go":     true,
		"difficulty:easy": true,
	}

	for _, name := range tagNames {
		if !expectedNames[name] {
			t.Errorf("Unexpected tag name: %s", name)
		}
	}
}

func TestQuiz_SetImportMetadata(t *testing.T) {
	quiz, _ := NewQuiz(
		NewQuizID(),
		MustNewQuizTitle("Test Quiz"),
		"Description",
		CategoryID{},
		MustNewTimeLimit(60),
		MustNewPassingScore(70),
		0,
	)

	batchID := "batch-2024-01-01"
	generatedAt := int64(1704067200)

	quiz.SetImportMetadata(batchID, generatedAt)

	if quiz.ImportBatchID() == nil || *quiz.ImportBatchID() != batchID {
		t.Errorf("Expected import batch ID %q, got %v", batchID, quiz.ImportBatchID())
	}

	if quiz.GeneratedAt() == nil || *quiz.GeneratedAt() != generatedAt {
		t.Errorf("Expected generated at %d, got %v", generatedAt, quiz.GeneratedAt())
	}
}

// Helper function for tests
func MustNewQuizTitle(title string) QuizTitle {
	qt, err := NewQuizTitle(title)
	if err != nil {
		panic(err)
	}
	return qt
}

func MustNewTimeLimit(seconds int) TimeLimit {
	tl, err := NewTimeLimit(seconds)
	if err != nil {
		panic(err)
	}
	return tl
}

func MustNewPassingScore(percentage int) PassingScore {
	ps, err := NewPassingScore(percentage)
	if err != nil {
		panic(err)
	}
	return ps
}
