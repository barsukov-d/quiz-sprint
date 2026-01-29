package quiz

import (
	"errors"
	"regexp"
	"strings"
)

// Tag Aggregate - Represents a tag for quiz classification
// Tags follow the format: {category}:{value}
// Examples: language:go, difficulty:easy, topic:concurrency
type Tag struct {
	id   TagID
	name TagName
}

// Value Objects

// TagID - Unique identifier for a tag (derived from name)
type TagID struct {
	value string
}

// TagName - Tag name with validation
type TagName struct {
	value string
}

// Tag validation rules
var (
	// Pattern: lowercase alphanumeric, hyphens, and colon separator
	// Format: {category}:{value}
	// Examples: language:go, difficulty:easy, topic:web-development
	tagNamePattern = regexp.MustCompile(`^[a-z0-9-]+:[a-z0-9-]+$`)

	// Maximum length for tag names
	maxTagNameLength = 100

	// Valid tag categories
	validTagCategories = map[string]bool{
		"language":   true, // Programming languages (go, python, javascript)
		"difficulty": true, // Quiz difficulty (easy, medium, hard, expert)
		"topic":      true, // Specific topics (variables, functions, concurrency)
		"domain":     true, // General domains (programming, history, science)
		"format":     true, // Question format (multiple-choice, true-false)
	}
)

// Domain Errors
var (
	ErrEmptyTagName        = errors.New("tag name cannot be empty")
	ErrTagNameTooLong      = errors.New("tag name exceeds maximum length of 100 characters")
	ErrInvalidTagFormat    = errors.New("tag name must follow format {category}:{value} with lowercase alphanumeric and hyphens only")
	ErrInvalidTagCategory  = errors.New("tag category must be one of: language, difficulty, topic, domain, format")
	ErrTagNameHasSpaces    = errors.New("tag name cannot contain spaces (use hyphens instead)")
	ErrTagNameHasUppercase = errors.New("tag name must be lowercase")
	ErrTagMissingColon     = errors.New("tag name must contain a colon separator (:)")
)

// Constructors

// NewTag creates a new Tag with validation
func NewTag(name string) (*Tag, error) {
	tagName, err := NewTagName(name)
	if err != nil {
		return nil, err
	}

	tagID := NewTagID(name)

	return &Tag{
		id:   tagID,
		name: tagName,
	}, nil
}

// ReconstructTag reconstructs a Tag from database without validation
// Used by repository when loading from database
func ReconstructTag(id, name string) *Tag {
	return &Tag{
		id:   TagID{value: id},
		name: TagName{value: name},
	}
}

// NewTagID creates a TagID from tag name
// TagID is derived from the tag name for simplicity
func NewTagID(tagName string) TagID {
	// Convert name to ID format (lowercase, replace special chars)
	id := strings.ToLower(tagName)
	id = strings.ReplaceAll(id, ":", "-")
	return TagID{value: id}
}

// NewTagName creates a TagName with validation
func NewTagName(value string) (TagName, error) {
	// 1. Empty check
	if value == "" {
		return TagName{}, ErrEmptyTagName
	}

	// 2. Length check
	if len(value) > maxTagNameLength {
		return TagName{}, ErrTagNameTooLong
	}

	// 3. Check for spaces
	if strings.Contains(value, " ") {
		return TagName{}, ErrTagNameHasSpaces
	}

	// 4. Check for uppercase
	if value != strings.ToLower(value) {
		return TagName{}, ErrTagNameHasUppercase
	}

	// 5. Check for colon separator
	if !strings.Contains(value, ":") {
		return TagName{}, ErrTagMissingColon
	}

	// 6. Pattern validation
	if !tagNamePattern.MatchString(value) {
		return TagName{}, ErrInvalidTagFormat
	}

	// 7. Validate category
	parts := strings.SplitN(value, ":", 2)
	category := parts[0]
	if !validTagCategories[category] {
		return TagName{}, ErrInvalidTagCategory
	}

	return TagName{value: value}, nil
}

// Getters

// ID returns the tag ID
func (t *Tag) ID() TagID {
	return t.id
}

// Name returns the tag name
func (t *Tag) Name() TagName {
	return t.name
}

// String returns the tag ID as string
func (id TagID) String() string {
	return id.value
}

// String returns the tag name as string
func (n TagName) String() string {
	return n.value
}

// Business Methods

// Category extracts the category part from tag name
// Example: "language:go" → "language"
func (n TagName) Category() string {
	parts := strings.SplitN(n.value, ":", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// Value extracts the value part from tag name
// Example: "language:go" → "go"
func (n TagName) Value() string {
	parts := strings.SplitN(n.value, ":", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// Equals checks if two tags are equal
func (t *Tag) Equals(other *Tag) bool {
	if other == nil {
		return false
	}
	return t.name.value == other.name.value
}

// Helper Functions

// IsValidTagCategory checks if a category is valid
func IsValidTagCategory(category string) bool {
	return validTagCategories[category]
}

// GetValidTagCategories returns list of valid tag categories
func GetValidTagCategories() []string {
	categories := make([]string, 0, len(validTagCategories))
	for category := range validTagCategories {
		categories = append(categories, category)
	}
	return categories
}

// InferCategoryFromTags infers quiz category from tags
// Priority: language:* > domain:* > fallback to "general"
func InferCategoryFromTags(tags []string) string {
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
			return domain // history, science, movies, etc.
		}
	}

	// Fallback
	return "general"
}
