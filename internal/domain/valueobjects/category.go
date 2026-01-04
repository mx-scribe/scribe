package valueobjects

// Category represents the category of a log entry.
type Category string

const (
	CategoryHTTP        Category = "http"
	CategoryDatabase    Category = "database"
	CategorySecurity    Category = "security"
	CategoryPerformance Category = "performance"
	CategoryBusiness    Category = "business"
	CategorySystem      Category = "system"
	CategoryGeneral     Category = "general"
)

// validCategories for validation.
var validCategories = map[Category]bool{
	CategoryHTTP:        true,
	CategoryDatabase:    true,
	CategorySecurity:    true,
	CategoryPerformance: true,
	CategoryBusiness:    true,
	CategorySystem:      true,
	CategoryGeneral:     true,
}

// IsValid checks if the category is valid.
func (c Category) IsValid() bool {
	return validCategories[c]
}

// String returns the string representation of the category.
func (c Category) String() string {
	return string(c)
}

// DefaultCategory returns the default category when none is specified.
func DefaultCategory() Category {
	return CategoryGeneral
}

// CategoryFromString creates a Category from a string, returns default if invalid.
func CategoryFromString(s string) Category {
	category := Category(s)
	if category.IsValid() {
		return category
	}
	return DefaultCategory()
}
