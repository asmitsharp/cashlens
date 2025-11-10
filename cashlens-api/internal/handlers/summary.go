package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ashmitsharp/cashlens-api/internal/database/db"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
)

type SummaryHandler struct {
	queries *db.Queries
}

func NewSummaryHandler(queries *db.Queries) *SummaryHandler {
	return &SummaryHandler{
		queries: queries,
	}
}

type KPIsResponse struct {
	TotalInflow      float64 `json:"total_inflow"`
	TotalOutflow     float64 `json:"total_outflow"`
	NetCashFlow      float64 `json:"net_cash_flow"`
	TransactionCount int64   `json:"transaction_count"`
}

type NetFlowTrendPoint struct {
	Period  string  `json:"period"`
	Inflow  float64 `json:"inflow"`
	Outflow float64 `json:"outflow"`
	NetFlow float64 `json:"net_flow"`
}

type SummaryResponse struct {
	KPIs          KPIsResponse        `json:"kpis"`
	NetFlowTrend  []NetFlowTrendPoint `json:"net_flow_trend"`
	FromDate      string              `json:"from_date"`
	ToDate        string              `json:"to_date"`
	GroupBy       string              `json:"group_by"`
}

// GetSummary handles GET /v1/summary
// Query params: from (date), to (date), group_by (day|week|month|year)
func (h *SummaryHandler) GetSummary(c fiber.Ctx) error {
	// Extract user ID from Clerk auth middleware
	clerkUserID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get user from database using Clerk ID
	user, err := h.queries.GetUserByClerkID(c.Context(), clerkUserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Parse query parameters
	fromStr := c.Query("from")
	toStr := c.Query("to")
	groupBy := c.Query("group_by", "month")

	// Default to last 12 months if not provided
	var fromDate, toDate time.Time
	if fromStr == "" || toStr == "" {
		toDate = time.Now()
		fromDate = toDate.AddDate(-1, 0, 0) // 1 year ago
	} else {
		fromDate, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid from date format: %s", err.Error()),
			})
		}
		toDate, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid to date format: %s", err.Error()),
			})
		}
	}

	// Validate groupBy parameter
	validGroupBy := map[string]bool{
		"day":   true,
		"week":  true,
		"month": true,
		"year":  true,
	}
	if !validGroupBy[groupBy] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid group_by parameter. Must be one of: day, week, month, year",
		})
	}

	// Convert to pgtype.Date
	fromPgDate := pgtype.Date{
		Time:  fromDate,
		Valid: true,
	}
	toPgDate := pgtype.Date{
		Time:  toDate,
		Valid: true,
	}

	// Get KPIs
	kpisRow, err := h.queries.GetKPIs(c.Context(), db.GetKPIsParams{
		UserID:    user.ID,
		TxnDate:   fromPgDate,
		TxnDate_2: toPgDate,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch KPIs: %s", err.Error()),
		})
	}

	// Convert KPIs to response format
	kpis := KPIsResponse{
		TotalInflow:      convertToFloat64(kpisRow.TotalInflow),
		TotalOutflow:     convertToFloat64(kpisRow.TotalOutflow),
		NetCashFlow:      convertToFloat64(kpisRow.NetCashFlow),
		TransactionCount: kpisRow.TransactionCount,
	}

	// Get cash flow trend (inflow, outflow, net)
	trendRows, err := h.queries.GetCashFlowTrend(c.Context(), db.GetCashFlowTrendParams{
		UserID:    user.ID,
		TxnDate:   fromPgDate,
		TxnDate_2: toPgDate,
		DateTrunc: groupBy,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch cash flow trend: %s", err.Error()),
		})
	}

	// Convert trend to response format
	trend := make([]NetFlowTrendPoint, 0, len(trendRows))
	for _, row := range trendRows {
		periodStr := formatPeriodTimestamp(row.Period, groupBy)

		trend = append(trend, NetFlowTrendPoint{
			Period:  periodStr,
			Inflow:  convertToFloat64(row.Inflow),
			Outflow: convertToFloat64(row.Outflow),
			NetFlow: convertToFloat64(row.NetFlow),
		})
	}

	response := SummaryResponse{
		KPIs:         kpis,
		NetFlowTrend: trend,
		FromDate:     fromDate.Format("2006-01-02"),
		ToDate:       toDate.Format("2006-01-02"),
		GroupBy:      groupBy,
	}

	return c.JSON(response)
}

// convertToFloat64 converts pgtype.Numeric or interface{} to float64
func convertToFloat64(val interface{}) float64 {
	if val == nil {
		return 0.0
	}

	switch v := val.(type) {
	case pgtype.Numeric:
		if !v.Valid {
			return 0.0
		}
		f, _ := v.Float64Value()
		return f.Float64
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case string:
		// Try to parse string as float
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return 0.0
	default:
		// Log unexpected type for debugging
		fmt.Printf("Warning: unexpected type %T for value %v\n", v, v)
		return 0.0
	}
}

// formatPeriodTimestamp formats the period timestamp based on groupBy parameter
func formatPeriodTimestamp(ts pgtype.Timestamp, groupBy string) string {
	if !ts.Valid {
		return ""
	}

	// Format based on groupBy
	switch groupBy {
	case "day":
		return ts.Time.Format("2006-01-02")
	case "week":
		// Return the start of the week
		return ts.Time.Format("2006-01-02")
	case "month":
		// Return YYYY-MM format
		return ts.Time.Format("2006-01")
	case "year":
		// Return YYYY format
		return ts.Time.Format("2006")
	default:
		return ts.Time.Format("2006-01-02")
	}
}
