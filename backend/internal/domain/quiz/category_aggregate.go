package quiz

// Category represents a quiz category. It's an aggregate root.
type Category struct {
	id   CategoryID
	name CategoryName
}

// NewCategory creates a new Category aggregate.
func NewCategory(id CategoryID, name CategoryName) (*Category, error) {
	if id.IsZero() {
		return nil, ErrInvalidCategoryID
	}

	return &Category{
		id:   id,
		name: name,
	}, nil
}

// ReconstructCategory reconstructs a Category from persistence.
func ReconstructCategory(id CategoryID, name CategoryName) *Category {
	return &Category{id: id, name: name}
}

// ID returns the category's ID.
func (c *Category) ID() CategoryID {
	return c.id
}

// Name returns the category's name.
func (c *Category) Name() CategoryName {
	return c.name
}
