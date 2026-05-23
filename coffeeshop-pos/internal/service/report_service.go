package service

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

// ProfitReport contains profit/loss data for a date range.
type ProfitReport struct {
	TotalSales      int64              `json:"total_sales"`
	TotalRecipeCost int64              `json:"total_recipe_cost"`
	GrossProfit     int64              `json:"gross_profit"`
	ProfitMargin    float64            `json:"profit_margin"`
	OrderCount      int                `json:"order_count"`
	VoidedCount     int                `json:"voided_count"`
	TopItems        []TopSellingItem   `json:"top_items"`
	SalesBySource   map[string]int64   `json:"sales_by_source"`
	DailyBreakdown  []DailyReportEntry `json:"daily_breakdown"`
}

// TopSellingItem represents a top-selling menu item.
type TopSellingItem struct {
	NameAr       string `db:"name_ar_snapshot" json:"name_ar"`
	TotalQty     int    `db:"total_qty"        json:"total_qty"`
	TotalRevenue int64  `db:"total_revenue"    json:"total_revenue"`
}

// DailyReportEntry is a single-day aggregation within a range report.
type DailyReportEntry struct {
	Date       string `db:"day"          json:"date"`
	OrderCount int    `db:"order_count"  json:"order_count"`
	TotalSales int64  `db:"total_sales"  json:"total_sales"`
}

// ReportService is a Wails-bound service for profit reports.
type ReportService struct {
	db *sqlx.DB
}

// NewReportService creates a new ReportService.
func NewReportService(db *sqlx.DB) *ReportService {
	return &ReportService{db: db}
}

// GetProfitReport generates a profit report for a date range.
// from and to should be in "YYYY-MM-DD" format.
func (s *ReportService) GetProfitReport(from, to string) (*ProfitReport, error) {
	if from == "" || to == "" {
		return nil, fmt.Errorf("from and to dates are required")
	}

	report := &ProfitReport{
		SalesBySource: make(map[string]int64),
	}

	// 1. Total sales + order count (non-voided)
	err := s.db.Get(&report.TotalSales,
		`SELECT COALESCE(SUM(total), 0) FROM orders
		 WHERE status != 'voided' AND created_at >= ? AND created_at < date(?, '+1 day')`,
		from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total sales: %w", err)
	}

	err = s.db.Get(&report.OrderCount,
		`SELECT COUNT(*) FROM orders
		 WHERE status != 'voided' AND created_at >= ? AND created_at < date(?, '+1 day')`,
		from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	// 2. Voided count
	err = s.db.Get(&report.VoidedCount,
		`SELECT COUNT(*) FROM orders
		 WHERE status = 'voided' AND created_at >= ? AND created_at < date(?, '+1 day')`,
		from, to)
	if err != nil {
		slog.Warn("report: failed to count voided", "error", err)
	}

	// 3. Total recipe cost
	report.TotalRecipeCost = s.calculateTotalCost(from, to)

	// 4. Gross profit + margin
	report.GrossProfit = report.TotalSales - report.TotalRecipeCost
	if report.TotalSales > 0 {
		report.ProfitMargin = float64(report.GrossProfit) / float64(report.TotalSales) * 100
	}

	// 5. Top selling items
	err = s.db.Select(&report.TopItems,
		`SELECT oi.name_ar_snapshot, SUM(oi.quantity) as total_qty, SUM(oi.line_total) as total_revenue
		 FROM order_items oi
		 JOIN orders o ON o.id = oi.order_id
		 WHERE o.status != 'voided' AND o.created_at >= ? AND o.created_at < date(?, '+1 day')
		 GROUP BY oi.name_ar_snapshot
		 ORDER BY total_qty DESC
		 LIMIT 10`,
		from, to)
	if err != nil {
		slog.Warn("report: failed to get top items", "error", err)
		report.TopItems = []TopSellingItem{}
	}

	// 6. Sales by source
	type sourceRow struct {
		Source string `db:"source"`
		Total  int64  `db:"total"`
	}
	var sources []sourceRow
	err = s.db.Select(&sources,
		`SELECT source, COALESCE(SUM(total), 0) as total FROM orders
		 WHERE status != 'voided' AND created_at >= ? AND created_at < date(?, '+1 day')
		 GROUP BY source`,
		from, to)
	if err != nil {
		slog.Warn("report: failed to get sales by source", "error", err)
	}
	for _, sr := range sources {
		report.SalesBySource[sr.Source] = sr.Total
	}

	// 7. Daily breakdown
	err = s.db.Select(&report.DailyBreakdown,
		`SELECT date(created_at) as day, COUNT(*) as order_count, COALESCE(SUM(total), 0) as total_sales
		 FROM orders
		 WHERE status != 'voided' AND created_at >= ? AND created_at < date(?, '+1 day')
		 GROUP BY date(created_at)
		 ORDER BY day ASC`,
		from, to)
	if err != nil {
		slog.Warn("report: failed to get daily breakdown", "error", err)
		report.DailyBreakdown = []DailyReportEntry{}
	}

	return report, nil
}

// calculateTotalCost computes the recipe-based cost for all non-voided orders in a date range.
func (s *ReportService) calculateTotalCost(from, to string) int64 {
	// For each order item, look up its recipe → inventory_items.unit_cost × recipe qty × order qty
	// Falls back to menu_items.cached_auto_cost or manual_cost_price if available
	type costRow struct {
		MenuItemID string `db:"menu_item_id"`
		OrderQty   int    `db:"order_qty"`
	}

	var items []costRow
	err := s.db.Select(&items,
		`SELECT oi.menu_item_id, SUM(oi.quantity) as order_qty
		 FROM order_items oi
		 JOIN orders o ON o.id = oi.order_id
		 WHERE o.status != 'voided'
		   AND oi.menu_item_id IS NOT NULL AND oi.menu_item_id != ''
		   AND o.created_at >= ? AND o.created_at < date(?, '+1 day')
		 GROUP BY oi.menu_item_id`,
		from, to)
	if err != nil {
		slog.Warn("report: failed to query order items for cost", "error", err)
		return 0
	}

	var totalCost int64
	for _, item := range items {
		cost := s.getCostPerUnit(item.MenuItemID)
		totalCost += cost * int64(item.OrderQty)
	}

	return totalCost
}

// getCostPerUnit computes the recipe cost for a single unit of a menu item.
func (s *ReportService) getCostPerUnit(menuItemID string) int64 {
	type recipeRow struct {
		Quantity int   `db:"quantity"`
		UnitCost int64 `db:"unit_cost"`
	}

	var ingredients []recipeRow
	err := s.db.Select(&ingredients,
		`SELECT ri.quantity, COALESCE(ii.unit_cost, 0) as unit_cost
		 FROM recipe_ingredients ri
		 JOIN inventory_items ii ON ii.id = ri.inventory_item_id
		 WHERE ri.menu_item_id = ?`,
		menuItemID)
	if err != nil || len(ingredients) == 0 {
		// Fallback to cached cost or manual cost
		var costCalcMethod string
		var manualCost, cachedCost int64
		err = s.db.QueryRow(
			`SELECT cost_calc_method, manual_cost_price, cached_auto_cost FROM menu_items WHERE id = ?`,
			menuItemID).Scan(&costCalcMethod, &manualCost, &cachedCost)
		if err != nil {
			return 0
		}
		if costCalcMethod == "manual" {
			return manualCost
		}
		return cachedCost
	}

	var cost int64
	for _, ing := range ingredients {
		cost += int64(ing.Quantity) * ing.UnitCost
	}
	return cost
}
