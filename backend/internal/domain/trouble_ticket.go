package domain

import (
	"time"
)

type TicketStatus string

const (
	TicketStatusOpen            TicketStatus = "open"
	TicketStatusInProgress      TicketStatus = "in_progress"
	TicketStatusPendingCustomer TicketStatus = "pending_customer"
	TicketStatusPendingVendor   TicketStatus = "pending_vendor"
	TicketStatusResolved        TicketStatus = "resolved"
	TicketStatusClosed          TicketStatus = "closed"
	TicketStatusCancelled       TicketStatus = "cancelled"
)

type TicketPriority string

const (
	TicketPriorityLow      TicketPriority = "low"
	TicketPriorityMedium   TicketPriority = "medium"
	TicketPriorityHigh     TicketPriority = "high"
	TicketPriorityCritical TicketPriority = "critical"
)

type TicketCategory string

const (
	TicketCategoryNoConnection   TicketCategory = "no_connection"
	TicketCategorySlowConnection TicketCategory = "slow_connection"
	TicketCategoryIntermittent   TicketCategory = "intermittent"
	TicketCategoryHardwareIssue  TicketCategory = "hardware_issue"
	TicketCategoryCableIssue     TicketCategory = "cable_issue"
	TicketCategoryOnuIssue       TicketCategory = "onu_issue"
	TicketCategoryOltIssue       TicketCategory = "olt_issue"
	TicketCategoryMikrotikIssue  TicketCategory = "mikrotik_issue"
	TicketCategoryOther          TicketCategory = "other"
)

// TroubleTicket represents customer support tickets.
type TroubleTicket struct {
	ID                   uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PelangganID          uint64         `gorm:"index;not null" json:"pelanggan_id"`
	DataTeknisID         *uint64        `gorm:"index" json:"data_teknis_id"`
	TicketNumber         string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"ticket_number"`
	Title                string         `gorm:"type:varchar(200);index;not null" json:"title"`
	Description          string         `gorm:"type:text;not null" json:"description"`
	Category             TicketCategory `gorm:"type:varchar(50);index;not null" json:"category"`
	Priority             TicketPriority `gorm:"type:varchar(50);index;not null" json:"priority"`
	Status               TicketStatus   `gorm:"type:varchar(50);default:'open';index;not null" json:"status"`
	DowntimeStart        *time.Time     `gorm:"type:datetime;index" json:"downtime_start"`
	DowntimeEnd          *time.Time     `gorm:"type:datetime;index" json:"downtime_end"`
	TotalDowntimeMinutes *int           `gorm:"index" json:"total_downtime_minutes"`
	PendingStart         *time.Time     `gorm:"type:datetime" json:"pending_start"`
	TotalPendingMinutes  int            `gorm:"default:0;not null" json:"total_pending_minutes"`
	AssignedTo           *uint64        `gorm:"index" json:"assigned_to"`
	ResolvedAt           *time.Time     `gorm:"type:datetime;index" json:"resolved_at"`
	ResolutionNotes      *string        `gorm:"type:text" json:"resolution_notes"`
	CustomerNotified     bool           `gorm:"default:false;not null" json:"customer_notified"`
	LastCustomerContact  *time.Time     `gorm:"type:datetime" json:"last_customer_contact"`
	Evidence             *string        `gorm:"type:text" json:"evidence"`
	CreatedAt            time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index;not null" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime;index;not null" json:"updated_at"`

	// Relationships
	Pelanggan    *Pelanggan      `gorm:"foreignKey:PelangganID" json:"pelanggan"`
	DataTeknis   *DataTeknis     `gorm:"foreignKey:DataTeknisID" json:"data_teknis"`
	AssignedUser *User           `gorm:"foreignKey:AssignedTo" json:"assigned_user"`
	History      []TicketHistory `gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE" json:"history"`
	ActionsTaken []ActionTaken   `gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE" json:"actions_taken"`
}

// TableName overrides the default table name for TroubleTicket
func (TroubleTicket) TableName() string {
	return "trouble_ticket"
}

// TicketHistory tracks status changes for a TroubleTicket.
type TicketHistory struct {
	ID        uint64        `gorm:"primaryKey;autoIncrement" json:"id"`
	TicketID  uint64        `gorm:"index;not null" json:"ticket_id"`
	OldStatus *TicketStatus `gorm:"type:varchar(50)" json:"old_status"`
	NewStatus TicketStatus  `gorm:"type:varchar(50);not null" json:"new_status"`
	ChangedBy *uint64       `gorm:"index" json:"changed_by"`
	Notes     *string       `gorm:"type:text" json:"notes"`
	CreatedAt time.Time     `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index;not null" json:"created_at"`

	// Relationships
	Ticket      *TroubleTicket `gorm:"foreignKey:TicketID" json:"-"`
	ChangedUser *User          `gorm:"foreignKey:ChangedBy" json:"changed_user"`
}

// TableName overrides the default table name for TicketHistory
func (TicketHistory) TableName() string {
	return "ticket_history"
}

// ActionTaken stores technical actions and evidence on a ticket.
type ActionTaken struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TicketID          uint64    `gorm:"index;not null" json:"ticket_id"`
	ActionDescription string    `gorm:"type:text;not null" json:"action_description"`
	SummaryProblem    string    `gorm:"type:text;not null" json:"summary_problem"`
	SummaryAction     string    `gorm:"type:text;not null" json:"summary_action"`
	Evidence          *string   `gorm:"type:text" json:"evidence"` // JSON array string for multiple files
	Notes             *string   `gorm:"type:text" json:"notes"`
	TakenBy           *uint64   `gorm:"index" json:"taken_by"`
	CreatedAt         time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index;not null" json:"created_at"`

	// Relationships
	Ticket    *TroubleTicket `gorm:"foreignKey:TicketID" json:"-"`
	TakenUser *User          `gorm:"foreignKey:TakenBy" json:"taken_user"`
}

// TableName overrides the default table name for ActionTaken
func (ActionTaken) TableName() string {
	return "action_taken"
}

func (t *TroubleTicket) CalculateDowntimeMinutes() int {
	if t.DowntimeStart == nil {
		return 0
	}

	var endTime time.Time
	if t.DowntimeEnd != nil {
		endTime = *t.DowntimeEnd
	} else {
		endTime = time.Now()
	}

	totalMinutes := int(endTime.Sub(*t.DowntimeStart).Minutes())

	currentPending := 0
	if t.PendingStart != nil && t.DowntimeEnd == nil {
		currentPending = int(time.Now().Sub(*t.PendingStart).Minutes())
	}

	downtime := totalMinutes - (t.TotalPendingMinutes + currentPending)
	if downtime < 0 {
		return 0
	}
	return downtime
}

func (t *TroubleTicket) UpdateDowntime() {
	minutes := t.CalculateDowntimeMinutes()
	t.TotalDowntimeMinutes = &minutes
}

