package valueobjects

import "testing"

func TestCategory_IsValid(t *testing.T) {
	tests := []struct {
		category Category
		want     bool
	}{
		{CategoryHTTP, true},
		{CategoryDatabase, true},
		{CategorySecurity, true},
		{CategoryPerformance, true},
		{CategoryBusiness, true},
		{CategorySystem, true},
		{CategoryGeneral, true},
		{Category("invalid"), false},
		{Category(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			if got := tt.category.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCategory_String(t *testing.T) {
	tests := []struct {
		category Category
		want     string
	}{
		{CategoryHTTP, "http"},
		{CategoryDatabase, "database"},
		{CategorySecurity, "security"},
		{CategoryPerformance, "performance"},
		{CategoryBusiness, "business"},
		{CategorySystem, "system"},
		{CategoryGeneral, "general"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.category.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultCategory(t *testing.T) {
	if got := DefaultCategory(); got != CategoryGeneral {
		t.Errorf("DefaultCategory() = %v, want %v", got, CategoryGeneral)
	}
}

func TestCategoryFromString(t *testing.T) {
	tests := []struct {
		input string
		want  Category
	}{
		{"http", CategoryHTTP},
		{"database", CategoryDatabase},
		{"security", CategorySecurity},
		{"performance", CategoryPerformance},
		{"business", CategoryBusiness},
		{"system", CategorySystem},
		{"general", CategoryGeneral},
		{"invalid", CategoryGeneral}, // defaults to general
		{"", CategoryGeneral},        // defaults to general
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := CategoryFromString(tt.input); got != tt.want {
				t.Errorf("CategoryFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
