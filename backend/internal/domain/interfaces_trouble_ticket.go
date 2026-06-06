package domain

import (
	"context"
	"time"
)

type TroubleTicketRepository interface {
	GetAll(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]TroubleTicket, int64, error)
	GetByID(ctx context.Context, id uint64) (*TroubleTicket, error)
	Create(ctx context.Context, ticket *TroubleTicket) error
	Update(ctx context.Context, ticket *TroubleTicket) error
	Delete(ctx context.Context, id uint64) error
	GetLastTicket(ctx context.Context) (*TroubleTicket, error)
	
	// History and Actions Taken
	CreateHistory(ctx context.Context, history *TicketHistory) error
	CreateActionTaken(ctx context.Context, action *ActionTaken) error
	GetHistory(ctx context.Context, ticketID uint64) ([]TicketHistory, error)
	GetActionsTaken(ctx context.Context, ticketID uint64) ([]ActionTaken, error)
	GetStatistics(ctx context.Context) (map[string]interface{}, error)
	GetMonthlyTrends(ctx context.Context, months int) ([]map[string]interface{}, error)
	GetCategoryPerformance(ctx context.Context, dateFrom, dateTo *time.Time) ([]map[string]interface{}, error)
	GetUserPerformance(ctx context.Context, dateFrom, dateTo *time.Time) ([]map[string]interface{}, error)
	GetDowntimeAnalysis(ctx context.Context, dateFrom, dateTo *time.Time) (map[string]interface{}, error)
}

type TroubleTicketUsecase interface {
	FetchAll(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]TroubleTicket, int64, error)
	GetByID(ctx context.Context, id uint64) (*TroubleTicket, error)
	Create(ctx context.Context, ticket *TroubleTicket, userID uint64) error
	Update(ctx context.Context, id uint64, ticket *TroubleTicket, userID uint64) error
	UpdateStatus(ctx context.Context, id uint64, status TicketStatus, notes string, actionDesc string, problemSummary string, actionSummary string, evidence string, userID uint64) error
	UpdateDowntime(ctx context.Context, id uint64, update map[string]interface{}, userID uint64) error
	AssignTicket(ctx context.Context, id uint64, assignedTo uint64, notes string, userID uint64) error
	AddAction(ctx context.Context, id uint64, action *ActionTaken, userID uint64) error
	GetHistory(ctx context.Context, ticketID uint64) ([]TicketHistory, error)
	GetActions(ctx context.Context, ticketID uint64) ([]ActionTaken, error)
	Delete(ctx context.Context, id uint64, userID uint64) error
	GetStatistics(ctx context.Context) (map[string]interface{}, error)
	GetMonthlyTrends(ctx context.Context, months int) ([]map[string]interface{}, error)
	GetCategoryPerformance(ctx context.Context, dateFrom, dateTo *time.Time) ([]map[string]interface{}, error)
	GetUserPerformance(ctx context.Context, dateFrom, dateTo *time.Time) ([]map[string]interface{}, error)
	GetDowntimeAnalysis(ctx context.Context, dateFrom, dateTo *time.Time) (map[string]interface{}, error)
}
