package queries

// StatsOutput represents log statistics.
type StatsOutput struct {
	Total       int            `json:"total"`
	Last24Hours int            `json:"last_24_hours"`
	BySeverity  map[string]int `json:"by_severity"`
	BySource    map[string]int `json:"by_source"`
}

// StatsRepository defines the interface for stats queries.
type StatsRepository interface {
	Count() (int, error)
	CountLast24Hours() (int, error)
	CountBySeverity() (map[string]int, error)
	CountBySource() (map[string]int, error)
}

// GetStatsHandler handles the get stats query.
type GetStatsHandler struct {
	repo StatsRepository
}

// NewGetStatsHandler creates a new get stats handler.
func NewGetStatsHandler(repo StatsRepository) *GetStatsHandler {
	return &GetStatsHandler{repo: repo}
}

// Handle executes the get stats query.
func (h *GetStatsHandler) Handle() (*StatsOutput, error) {
	total, err := h.repo.Count()
	if err != nil {
		return nil, err
	}

	last24h, err := h.repo.CountLast24Hours()
	if err != nil {
		return nil, err
	}

	bySeverity, err := h.repo.CountBySeverity()
	if err != nil {
		return nil, err
	}

	bySource, err := h.repo.CountBySource()
	if err != nil {
		return nil, err
	}

	return &StatsOutput{
		Total:       total,
		Last24Hours: last24h,
		BySeverity:  bySeverity,
		BySource:    bySource,
	}, nil
}
