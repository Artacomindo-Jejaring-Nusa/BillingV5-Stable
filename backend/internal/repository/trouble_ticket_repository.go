package repository

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type troubleTicketRepository struct {
	db *gorm.DB
}

func NewTroubleTicketRepository(db *gorm.DB) domain.TroubleTicketRepository {
	return &troubleTicketRepository{db: db}
}

func (r *troubleTicketRepository) GetAll(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]domain.TroubleTicket, int64, error) {
	var tickets []domain.TroubleTicket
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.TroubleTicket{})

	// Apply Filters
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if priority, ok := filters["priority"]; ok && priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if category, ok := filters["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if assignedTo, ok := filters["assigned_to"]; ok && assignedTo != nil {
		query = query.Where("assigned_to = ?", assignedTo)
	}
	if pelangganID, ok := filters["pelanggan_id"]; ok && pelangganID != nil {
		query = query.Where("pelanggan_id = ?", pelangganID)
	}
	if search, ok := filters["search"]; ok && search != "" {
		searchTerm := "%" + search.(string) + "%"
		query = query.Where("title LIKE ? OR description LIKE ? OR ticket_number LIKE ?", searchTerm, searchTerm, searchTerm)
	}
	if idBrand, ok := filters["id_brand"]; ok && idBrand != "" {
		query = query.Joins("JOIN pelanggan ON pelanggan.id = trouble_ticket.pelanggan_id").Where("pelanggan.id_brand = ?", idBrand)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Pelanggan.HargaLayanan").
		Preload("DataTeknis").
		Preload("AssignedUser").
		Limit(limit).
		Offset(offset).
		Order("created_at desc").
		Find(&tickets).Error

	return tickets, total, err
}

func (r *troubleTicketRepository) GetByID(ctx context.Context, id uint64) (*domain.TroubleTicket, error) {
	var ticket domain.TroubleTicket
	err := r.db.WithContext(ctx).
		Preload("Pelanggan.HargaLayanan").
		Preload("DataTeknis").
		Preload("AssignedUser").
		First(&ticket, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("trouble ticket not found")
		}
		return nil, err
	}
	return &ticket, nil
}

func (r *troubleTicketRepository) Create(ctx context.Context, ticket *domain.TroubleTicket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *troubleTicketRepository) Update(ctx context.Context, ticket *domain.TroubleTicket) error {
	return r.db.WithContext(ctx).Omit("Pelanggan", "DataTeknis", "AssignedUser").Save(ticket).Error
}

func (r *troubleTicketRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.TroubleTicket{}, id).Error
}

func (r *troubleTicketRepository) GetLastTicket(ctx context.Context) (*domain.TroubleTicket, error) {
	var ticket domain.TroubleTicket
	err := r.db.WithContext(ctx).Order("ticket_number desc").First(&ticket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ticket, nil
}

func (r *troubleTicketRepository) CreateHistory(ctx context.Context, history *domain.TicketHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *troubleTicketRepository) CreateActionTaken(ctx context.Context, action *domain.ActionTaken) error {
	return r.db.WithContext(ctx).Create(action).Error
}

func (r *troubleTicketRepository) GetHistory(ctx context.Context, ticketID uint64) ([]domain.TicketHistory, error) {
	var history []domain.TicketHistory
	err := r.db.WithContext(ctx).
		Preload("ChangedUser").
		Where("ticket_id = ?", ticketID).
		Order("created_at desc").
		Find(&history).Error
	return history, err
}

func (r *troubleTicketRepository) GetActionsTaken(ctx context.Context, ticketID uint64) ([]domain.ActionTaken, error) {
	var actions []domain.ActionTaken
	err := r.db.WithContext(ctx).
		Preload("TakenUser").
		Where("ticket_id = ?", ticketID).
		Order("created_at desc").
		Find(&actions).Error
	return actions, err
}

func (r *troubleTicketRepository) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	var totalTickets int64
	var openTickets int64
	var inProgressTickets int64
	var resolvedTickets int64
	var closedTickets int64
	var highPriorityTickets int64
	var criticalPriorityTickets int64
	var ticketsThisMonth int64
	var unresolvedOver24h int64

	db := r.db.WithContext(ctx).Model(&domain.TroubleTicket{})

	db.Count(&totalTickets)
	db.Where("status = ?", domain.TicketStatusOpen).Count(&openTickets)
	db.Where("status = ?", domain.TicketStatusInProgress).Count(&inProgressTickets)
	db.Where("status = ?", domain.TicketStatusResolved).Count(&resolvedTickets)
	db.Where("status = ?", domain.TicketStatusClosed).Count(&closedTickets)
	db.Where("priority = ?", domain.TicketPriorityHigh).Count(&highPriorityTickets)
	db.Where("priority = ?", domain.TicketPriorityCritical).Count(&criticalPriorityTickets)

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	db.Where("created_at >= ?", monthStart).Count(&ticketsThisMonth)

	yesterday := now.Add(-24 * time.Hour)
	db.Where("created_at <= ? AND status IN ?", yesterday, []string{"open", "in_progress", "pending_customer", "pending_vendor"}).Count(&unresolvedOver24h)

	var avgHours float64
	// Average resolution time query in MySQL
	r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
		Select("COALESCE(AVG(TIMESTAMPDIFF(HOUR, created_at, resolved_at)), 0)").
		Where("resolved_at IS NOT NULL").
		Row().
		Scan(&avgHours)

	return map[string]interface{}{
		"total_tickets":              totalTickets,
		"open_tickets":               openTickets,
		"in_progress_tickets":        inProgressTickets,
		"resolved_tickets":           resolvedTickets,
		"closed_tickets":             closedTickets,
		"high_priority_tickets":      highPriorityTickets,
		"critical_priority_tickets":  criticalPriorityTickets,
		"tickets_this_month":         ticketsThisMonth,
		"unresolved_over_24h":        unresolvedOver24h,
		"avg_resolution_time_hours": mathRound(avgHours, 2),
	}, nil
}

func mathRound(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func (r *troubleTicketRepository) GetMonthlyTrends(ctx context.Context, months int) ([]map[string]interface{}, error) {
	var trends []map[string]interface{}
	now := time.Now()

	for i := 0; i < months; i++ {
		monthStart := now.AddDate(0, -i, 0)
		monthStart = time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, now.Location())
		nextMonth := monthStart.AddDate(0, 1, 0)

		var total int64
		if err := r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
			Where("created_at >= ? AND created_at < ?", monthStart, nextMonth).
			Count(&total).Error; err != nil {
			return nil, err
		}

		var resolved int64
		if err := r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
			Where("resolved_at >= ? AND resolved_at < ?", monthStart, nextMonth).
			Count(&resolved).Error; err != nil {
			return nil, err
		}

		var avgHours float64
		r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
			Select("COALESCE(AVG(TIMESTAMPDIFF(HOUR, created_at, resolved_at)), 0)").
			Where("resolved_at >= ? AND resolved_at < ? AND created_at IS NOT NULL", monthStart, nextMonth).
			Row().
			Scan(&avgHours)

		type CategoryCount struct {
			Category string
			Count    int64
		}
		var catCounts []CategoryCount
		if err := r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
			Select("category, COUNT(id) as count").
			Where("created_at >= ? AND created_at < ?", monthStart, nextMonth).
			Group("category").
			Scan(&catCounts).Error; err != nil {
			return nil, err
		}

		byCategory := make(map[string]int64)
		for _, cc := range catCounts {
			byCategory[cc.Category] = cc.Count
		}

		type PriorityCount struct {
			Priority string
			Count    int64
		}
		var priCounts []PriorityCount
		if err := r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
			Select("priority, COUNT(id) as count").
			Where("created_at >= ? AND created_at < ?", monthStart, nextMonth).
			Group("priority").
			Scan(&priCounts).Error; err != nil {
			return nil, err
		}

		byPriority := make(map[string]int64)
		for _, pc := range priCounts {
			byPriority[pc.Priority] = pc.Count
		}

		trends = append(trends, map[string]interface{}{
			"month":      monthStart.Format("2006-01"),
			"month_name": monthStart.Format("January 2006"),
			"statistics": map[string]interface{}{
				"total":                total,
				"resolved":             resolved,
				"avg_resolution_hours": mathRound(avgHours, 2),
				"by_category":          byCategory,
				"by_priority":          byPriority,
			},
		})
	}

	// Reverse to show oldest to newest
	for i, j := 0, len(trends)-1; i < j; i, j = i+1, j-1 {
		trends[i], trends[j] = trends[j], trends[i]
	}

	return trends, nil
}

func (r *troubleTicketRepository) GetCategoryPerformance(ctx context.Context, dateFrom, dateTo *time.Time) ([]map[string]interface{}, error) {
	type CatPerfRow struct {
		Category             string
		TotalTickets         int64
		TotalDowntimeMinutes int64
	}

	query := r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
		Select("category, COUNT(id) as total_tickets, SUM(COALESCE(total_downtime_minutes, 0)) as total_downtime_minutes")

	if dateFrom != nil {
		query = query.Where("created_at >= ?", *dateFrom)
	}
	if dateTo != nil {
		query = query.Where("created_at <= ?", *dateTo)
	}

	var rows []CatPerfRow
	if err := query.Group("category").Order("total_tickets DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	var performance []map[string]interface{}
	for _, row := range rows {
		resolvedQuery := r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
			Where("category = ? AND status = ?", row.Category, domain.TicketStatusResolved)
		if dateFrom != nil {
			resolvedQuery = resolvedQuery.Where("created_at >= ?", *dateFrom)
		}
		if dateTo != nil {
			resolvedQuery = resolvedQuery.Where("created_at <= ?", *dateTo)
		}

		var resolved int64
		if err := resolvedQuery.Count(&resolved).Error; err != nil {
			return nil, err
		}

		avgResolution := 0.0
		if resolved > 0 {
			type TimeRow struct {
				CreatedAt  *time.Time
				ResolvedAt *time.Time
			}
			var timeRows []TimeRow
			tQuery := r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
				Select("created_at, resolved_at").
				Where("category = ? AND status = ? AND resolved_at IS NOT NULL AND created_at IS NOT NULL", row.Category, domain.TicketStatusResolved)
			if dateFrom != nil {
				tQuery = tQuery.Where("created_at >= ?", *dateFrom)
			}
			if dateTo != nil {
				tQuery = tQuery.Where("created_at <= ?", *dateTo)
			}

			if err := tQuery.Scan(&timeRows).Error; err == nil && len(timeRows) > 0 {
				var totalHours float64
				validCount := 0
				for _, tr := range timeRows {
					if tr.CreatedAt != nil && tr.ResolvedAt != nil {
						hours := tr.ResolvedAt.Sub(*tr.CreatedAt).Hours()
						totalHours += hours
						validCount++
					}
				}
				if validCount > 0 {
					avgResolution = mathRound(totalHours/float64(validCount), 2)
				}
			}
		}

		resolutionRate := 0.0
		if row.TotalTickets > 0 {
			resolutionRate = mathRound(float64(resolved)/float64(row.TotalTickets)*100, 2)
		}

		avgDowntime := 0.0
		if resolved > 0 {
			avgDowntime = mathRound(float64(row.TotalDowntimeMinutes)/float64(resolved), 2)
		}

		catDisplay := strings.ReplaceAll(row.Category, "_", " ")
		catDisplay = strings.Title(catDisplay)

		performance = append(performance, map[string]interface{}{
			"category":                row.Category,
			"category_display":        catDisplay,
			"total_tickets":           row.TotalTickets,
			"resolved_tickets":        resolved,
			"resolution_rate_percent": resolutionRate,
			"avg_resolution_hours":    avgResolution,
			"avg_downtime_minutes":    avgDowntime,
			"total_downtime_minutes":  row.TotalDowntimeMinutes,
		})
	}

	return performance, nil
}

func (r *troubleTicketRepository) GetUserPerformance(ctx context.Context, dateFrom, dateTo *time.Time) ([]map[string]interface{}, error) {
	type UserPerfRow struct {
		UserID        uint64
		UserName      string
		TotalAssigned int64
		FirstTicket   *time.Time
		LastTicket    *time.Time
	}

	query := r.db.WithContext(ctx).Table("trouble_ticket").
		Select("users.id as user_id, users.name as user_name, COUNT(trouble_ticket.id) as total_assigned, MIN(trouble_ticket.created_at) as first_ticket, MAX(trouble_ticket.created_at) as last_ticket").
		Joins("JOIN users ON trouble_ticket.assigned_to = users.id").
		Where("trouble_ticket.assigned_to IS NOT NULL")

	if dateFrom != nil {
		query = query.Where("trouble_ticket.created_at >= ?", *dateFrom)
	}
	if dateTo != nil {
		query = query.Where("trouble_ticket.created_at <= ?", *dateTo)
	}

	var rows []UserPerfRow
	if err := query.Group("users.id, users.name").Order("total_assigned DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	var performance []map[string]interface{}
	for _, row := range rows {
		resolvedQuery := r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
			Where("assigned_to = ? AND status = ?", row.UserID, domain.TicketStatusResolved)
		if dateFrom != nil {
			resolvedQuery = resolvedQuery.Where("created_at >= ?", *dateFrom)
		}
		if dateTo != nil {
			resolvedQuery = resolvedQuery.Where("created_at <= ?", *dateTo)
		}

		var resolved int64
		if err := resolvedQuery.Count(&resolved).Error; err != nil {
			return nil, err
		}

		avgResolution := 0.0
		if resolved > 0 {
			type TimeRow struct {
				CreatedAt  *time.Time
				ResolvedAt *time.Time
			}
			var timeRows []TimeRow
			tQuery := r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
				Select("created_at, resolved_at").
				Where("assigned_to = ? AND status = ? AND resolved_at IS NOT NULL AND created_at IS NOT NULL", row.UserID, domain.TicketStatusResolved)
			if dateFrom != nil {
				tQuery = tQuery.Where("created_at >= ?", *dateFrom)
			}
			if dateTo != nil {
				tQuery = tQuery.Where("created_at <= ?", *dateTo)
			}

			if err := tQuery.Scan(&timeRows).Error; err == nil && len(timeRows) > 0 {
				var totalHours float64
				validCount := 0
				for _, tr := range timeRows {
					if tr.CreatedAt != nil && tr.ResolvedAt != nil {
						hours := tr.ResolvedAt.Sub(*tr.CreatedAt).Hours()
						totalHours += hours
						validCount++
					}
				}
				if validCount > 0 {
					avgResolution = mathRound(totalHours/float64(validCount), 2)
				}
			}
		}

		resolutionRate := 0.0
		if row.TotalAssigned > 0 {
			resolutionRate = mathRound(float64(resolved)/float64(row.TotalAssigned)*100, 2)
		}

		var firstDateStr, lastDateStr *string
		if row.FirstTicket != nil {
			s := row.FirstTicket.Format(time.RFC3339)
			firstDateStr = &s
		}
		if row.LastTicket != nil {
			s := row.LastTicket.Format(time.RFC3339)
			lastDateStr = &s
		}

		performance = append(performance, map[string]interface{}{
			"user_id":                 row.UserID,
			"user_name":               row.UserName,
			"total_assigned":          row.TotalAssigned,
			"resolved_tickets":        resolved,
			"resolution_rate_percent": resolutionRate,
			"avg_resolution_hours":    avgResolution,
			"first_ticket_date":       firstDateStr,
			"last_ticket_date":        lastDateStr,
		})
	}

	return performance, nil
}

func (r *troubleTicketRepository) GetDowntimeAnalysis(ctx context.Context, dateFrom, dateTo *time.Time) (map[string]interface{}, error) {
	type OverallStats struct {
		TotalDowntime            int64   `gorm:"column:total_downtime"`
		AvgDowntime              float64 `gorm:"column:avg_downtime"`
		MaxDowntime              int64   `gorm:"column:max_downtime"`
		TicketsWithDowntimeCount int64   `gorm:"column:tickets_with_downtime_count"`
	}

	var overall OverallStats
	overallQuery := r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
		Select("SUM(COALESCE(total_downtime_minutes, 0)) as total_downtime, AVG(COALESCE(total_downtime_minutes, 0)) as avg_downtime, MAX(COALESCE(total_downtime_minutes, 0)) as max_downtime, COUNT(id) as tickets_with_downtime_count").
		Where("total_downtime_minutes IS NOT NULL AND total_downtime_minutes > 0")

	if dateFrom != nil {
		overallQuery = overallQuery.Where("created_at >= ?", *dateFrom)
	}
	if dateTo != nil {
		overallQuery = overallQuery.Where("created_at <= ?", *dateTo)
	}

	_ = overallQuery.Scan(&overall)

	type CatDowntimeRow struct {
		Category      string
		TotalDowntime int64
		AvgDowntime   float64
		TicketCount   int64
	}
	catQuery := r.db.WithContext(ctx).Model(&domain.TroubleTicket{}).
		Select("category, SUM(COALESCE(total_downtime_minutes, 0)) as total_downtime, AVG(COALESCE(total_downtime_minutes, 0)) as avg_downtime, COUNT(id) as ticket_count").
		Where("total_downtime_minutes IS NOT NULL AND total_downtime_minutes > 0")

	if dateFrom != nil {
		catQuery = catQuery.Where("created_at >= ?", *dateFrom)
	}
	if dateTo != nil {
		catQuery = catQuery.Where("created_at <= ?", *dateTo)
	}

	var catRows []CatDowntimeRow
	_ = catQuery.Group("category").Order("total_downtime DESC").Scan(&catRows)

	var categoryDowntime []map[string]interface{}
	for _, row := range catRows {
		catDisplay := strings.ReplaceAll(row.Category, "_", " ")
		catDisplay = strings.Title(catDisplay)

		categoryDowntime = append(categoryDowntime, map[string]interface{}{
			"category":               row.Category,
			"category_display":       catDisplay,
			"total_downtime_minutes": row.TotalDowntime,
			"avg_downtime_minutes":   mathRound(row.AvgDowntime, 2),
			"ticket_count":           row.TicketCount,
			"total_downtime_hours":   mathRound(float64(row.TotalDowntime)/60.0, 2),
			"avg_downtime_hours":     mathRound(row.AvgDowntime/60.0, 2),
		})
	}

	type CustDowntimeRow struct {
		CustomerID    uint64
		CustomerName  string
		TotalDowntime int64
		TicketCount   int64
		AvgDowntime   float64
	}
	custQuery := r.db.WithContext(ctx).Table("trouble_ticket").
		Select("pelanggan.id as customer_id, pelanggan.nama as customer_name, SUM(COALESCE(trouble_ticket.total_downtime_minutes, 0)) as total_downtime, COUNT(trouble_ticket.id) as ticket_count, AVG(COALESCE(trouble_ticket.total_downtime_minutes, 0)) as avg_downtime").
		Joins("JOIN pelanggan ON trouble_ticket.pelanggan_id = pelanggan.id").
		Where("trouble_ticket.total_downtime_minutes IS NOT NULL AND trouble_ticket.total_downtime_minutes > 0")

	if dateFrom != nil {
		custQuery = custQuery.Where("trouble_ticket.created_at >= ?", *dateFrom)
	}
	if dateTo != nil {
		custQuery = custQuery.Where("trouble_ticket.created_at <= ?", *dateTo)
	}

	var custRows []CustDowntimeRow
	_ = custQuery.Group("pelanggan.id, pelanggan.nama").Order("total_downtime DESC").Limit(20).Scan(&custRows)

	var customerDowntime []map[string]interface{}
	for _, row := range custRows {
		customerDowntime = append(customerDowntime, map[string]interface{}{
			"customer_id":            row.CustomerID,
			"customer_name":          row.CustomerName,
			"total_downtime_minutes": row.TotalDowntime,
			"ticket_count":           row.TicketCount,
			"avg_downtime_minutes":   mathRound(row.AvgDowntime, 2),
			"total_downtime_hours":   mathRound(float64(row.TotalDowntime)/60.0, 2),
			"avg_downtime_hours":     mathRound(row.AvgDowntime/60.0, 2),
		})
	}

	return map[string]interface{}{
		"overall_statistics": map[string]interface{}{
			"total_downtime_minutes": overall.TotalDowntime,
			"total_downtime_hours":   mathRound(float64(overall.TotalDowntime)/60.0, 2),
			"total_downtime_days":    mathRound(float64(overall.TotalDowntime)/(60.0*24.0), 2),
			"avg_downtime_minutes":   mathRound(overall.AvgDowntime, 2),
			"avg_downtime_hours":     mathRound(overall.AvgDowntime/60.0, 2),
			"max_downtime_minutes":   overall.MaxDowntime,
			"max_downtime_hours":     mathRound(float64(overall.MaxDowntime)/60.0, 2),
			"tickets_with_downtime":  overall.TicketsWithDowntimeCount,
		},
		"by_category":   categoryDowntime,
		"top_customers": customerDowntime,
	}, nil
}
