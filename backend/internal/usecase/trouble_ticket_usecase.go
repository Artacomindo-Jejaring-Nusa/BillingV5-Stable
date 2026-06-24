package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"billing-backend/config"
	"billing-backend/internal/domain"
)

type troubleTicketUsecase struct {
	repo         domain.TroubleTicketRepository
	systemRepo   domain.SystemRepository
	userRepo     domain.UserRepository
	notifUsecase domain.NotificationUsecase
	cfg          *config.Config
}

func NewTroubleTicketUsecase(r domain.TroubleTicketRepository, sr domain.SystemRepository, ur domain.UserRepository, nu domain.NotificationUsecase, cfg *config.Config) domain.TroubleTicketUsecase {
	return &troubleTicketUsecase{
		repo:         r,
		systemRepo:   sr,
		userRepo:     ur,
		notifUsecase: nu,
		cfg:          cfg,
	}
}

func (u *troubleTicketUsecase) logActivity(ctx context.Context, userID uint64, action string, details string) {
	log := &domain.ActivityLog{
		UserID:    userID,
		Action:    action,
		Details:   &details,
		Timestamp: time.Now(),
	}
	_ = u.systemRepo.CreateActivityLog(ctx, log)
}

func (u *troubleTicketUsecase) generateTicketNumber(ctx context.Context) (string, error) {
	last, err := u.repo.GetLastTicket(ctx)
	if err != nil {
		return "", err
	}
	prefix := "TFTTH"
	nextNum := 1
	if last != nil && last.TicketNumber != "" {
		parts := strings.Split(last.TicketNumber, "-")
		if len(parts) == 2 {
			var lastNum int
			_, err := fmt.Sscanf(parts[1], "%d", &lastNum)
			if err == nil {
				nextNum = lastNum + 1
			}
		}
	}
	return fmt.Sprintf("%s-%04d", prefix, nextNum), nil
}

func (u *troubleTicketUsecase) FetchAll(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]domain.TroubleTicket, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 15
	}
	offset := (page - 1) * pageSize
	return u.repo.GetAll(ctx, pageSize, offset, filters)
}

func (u *troubleTicketUsecase) GetByID(ctx context.Context, id uint64) (*domain.TroubleTicket, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *troubleTicketUsecase) Create(ctx context.Context, ticket *domain.TroubleTicket, userID uint64) error {
	ticketNo, err := u.generateTicketNumber(ctx)
	if err != nil {
		return err
	}
	ticket.TicketNumber = ticketNo
	ticket.Status = domain.TicketStatusOpen
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()
	
	now := time.Now()
	ticket.DowntimeStart = &now

	if err := u.repo.Create(ctx, ticket); err != nil {
		return err
	}

	// Create History
	history := &domain.TicketHistory{
		TicketID:  ticket.ID,
		OldStatus: nil,
		NewStatus: domain.TicketStatusOpen,
		ChangedBy: &userID,
		Notes:     &[]string{"Ticket created"}[0],
		CreatedAt: time.Now(),
	}
	_ = u.repo.CreateHistory(ctx, history)

	// Add action taken if initial evidence is provided
	if ticket.Evidence != nil && *ticket.Evidence != "" {
		action := &domain.ActionTaken{
			TicketID:          ticket.ID,
			ActionDescription: "Ticket created with evidence",
			SummaryProblem:    "Initial evidence provided during ticket creation",
			SummaryAction:     "Evidence uploaded with initial ticket",
			Evidence:          ticket.Evidence,
			TakenBy:           &userID,
			CreatedAt:         time.Now(),
		}
		_ = u.repo.CreateActionTaken(ctx, action)
	}

	// Log Activity
	u.logActivity(ctx, userID, "Create Trouble Ticket", fmt.Sprintf("Created ticket number %s for Pelanggan ID %d", ticket.TicketNumber, ticket.PelangganID))

	// Send WA to technician in background
	if ticket.AssignedTo != nil {
		u.sendTechnicianWA(ctx, *ticket.AssignedTo, ticket)
	}

	return nil
}

func (u *troubleTicketUsecase) sendTechnicianWA(ctx context.Context, technicianID uint64, ticket *domain.TroubleTicket) {
	// Send WA notification via outbox in background
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Fetch the full ticket details with preloads to ensure Pelanggan and DataTeknis are populated
		fullTicket, err := u.repo.GetByID(bgCtx, ticket.ID)
		if err != nil {
			log.Printf("WA Warning: Failed to fetch full ticket %d: %v\n", ticket.ID, err)
			// fallback to the passed ticket pointer
			fullTicket = ticket
		}

		// 1. Prioritas: cek phone_no dari User
		phone := ""
		user, err := u.userRepo.GetByID(bgCtx, technicianID)
		if err == nil && user != nil && user.PhoneNo != nil && *user.PhoneNo != "" {
			phone = *user.PhoneNo
		}

		// 2. Fallback: cek system_settings (backward-compatible)
		if phone == "" {
			key := fmt.Sprintf("TECHNICIAN_PHONE_ID_%d", technicianID)
			setting, err := u.systemRepo.GetSettingByKey(bgCtx, key)
			if err == nil && setting != nil && setting.SettingValue != nil && *setting.SettingValue != "" {
				phone = *setting.SettingValue
			}
		}

		if phone == "" {
			log.Printf("WA Warning: Phone not found for technician %d\n", technicianID)
			return
		}

		// Construct detailed message
		var customerName = "-"
		var customerPhone = "-"
		var customerAddress = "-"
		if fullTicket.Pelanggan != nil {
			customerName = fullTicket.Pelanggan.Nama
			customerPhone = fullTicket.Pelanggan.NoTelp

			// Build complete address: Alamat, Blok, Unit, AlamatCustom
			var addrParts []string
			if fullTicket.Pelanggan.Alamat != "" {
				addrParts = append(addrParts, fullTicket.Pelanggan.Alamat)
			}
			if fullTicket.Pelanggan.Blok != "" {
				addrParts = append(addrParts, "Blok " + fullTicket.Pelanggan.Blok)
			}
			if fullTicket.Pelanggan.Unit != "" {
				addrParts = append(addrParts, "Unit " + fullTicket.Pelanggan.Unit)
			}
			if fullTicket.Pelanggan.AlamatCustom != nil && *fullTicket.Pelanggan.AlamatCustom != "" {
				addrParts = append(addrParts, *fullTicket.Pelanggan.AlamatCustom)
			}

			if len(addrParts) > 0 {
				customerAddress = strings.Join(addrParts, ", ")
			}
		}

		var techDetails = ""
		if fullTicket.DataTeknis != nil {
			dt := fullTicket.DataTeknis
			pppoeUsername := "-"
			if dt.IDPelanggan != "" {
				pppoeUsername = dt.IDPelanggan
			}
			pppoePassword := "-"
			if dt.PasswordPppoe != "" {
				pppoePassword = dt.PasswordPppoe
			}
			ipAddress := "-"
			if dt.IPPelanggan != nil && *dt.IPPelanggan != "" {
				ipAddress = *dt.IPPelanggan
			}
			sn := "-"
			if dt.Sn != nil && *dt.Sn != "" {
				sn = *dt.Sn
			}
			oltName := "-"
			if dt.Olt != nil && *dt.Olt != "" {
				oltName = *dt.Olt
			} else if dt.OltCustom != nil && *dt.OltCustom != "" {
				oltName = *dt.OltCustom
			}
			pon := "-"
			if dt.Pon != nil {
				pon = fmt.Sprintf("%d", *dt.Pon)
			}
			odpCode := "-"
			if dt.Odp != nil && dt.Odp.KodeOdp != "" {
				odpCode = dt.Odp.KodeOdp
			} else if dt.OdpID != nil {
				odpCode = fmt.Sprintf("ID: %d", *dt.OdpID)
			}
			portOdp := "-"
			if dt.PortOdp != nil {
				portOdp = fmt.Sprintf("%d", *dt.PortOdp)
			}
			onuPower := "-"
			if dt.OnuPower != nil {
				onuPower = fmt.Sprintf("%d dBm", *dt.OnuPower)
			}

			techDetails = fmt.Sprintf(
				"*Data Teknis:*\n"+
					"- PPPoE Username: `%s`\n"+
					"- PPPoE Password: `%s`\n"+
					"- IP Pelanggan: `%s`\n"+
					"- Serial Number ONU: `%s`\n"+
					"- OLT: `%s` (PON Port: `%s`)\n"+
					"- ODP: `%s` (Port ODP: `%s`)\n"+
					"- ONU Power: `%s`\n\n",
				pppoeUsername, pppoePassword, ipAddress, sn, oltName, pon, odpCode, portOdp, onuPower,
			)
		}

		waMsg := fmt.Sprintf(
			"🔔 *TIKET GANGGUAN BARU*\n\n"+
				"No Tiket: *%s*\n"+
				"Judul: %s\n"+
				"Pelanggan ID: %d\n"+
				"Nama Pelanggan: *%s*\n"+
				"No. Telp Pelanggan: %s\n"+
				"Alamat Pelanggan: %s\n\n"+
				"%s"+
				"*Deskripsi:*\n%s\n\n"+
				"Segera cek dashboard untuk detailnya. Semangat!",
			fullTicket.TicketNumber, fullTicket.Title, fullTicket.PelangganID, customerName,
			customerPhone, customerAddress, techDetails, fullTicket.Description,
		)

		// Kirim via outbox pattern (write-first, send-optimistic, retry on failure)
		refType := "trouble_ticket"
		err = u.notifUsecase.SendWhatsApp(bgCtx, phone, waMsg, refType, ticket.ID, nil)
		if err != nil {
			log.Printf("Failed to enqueue WA for technician %d: %v\n", technicianID, err)
		} else {
			log.Printf("WA enqueued for technician %d (phone: %s)\n", technicianID, phone)
		}

		// Cek jika alamat mengandung Parama atau Waringin untuk notifikasi Leader (Yohan)
		addrLower := strings.ToLower(customerAddress)
		if strings.Contains(addrLower, "parama") || strings.Contains(addrLower, "waringin") {
			yohanUsers, _, err := u.userRepo.GetAll(bgCtx, 5, 0, "Yohan", nil)
			if err == nil && len(yohanUsers) > 0 {
				var yohanUser *domain.User
				for _, usr := range yohanUsers {
					if strings.Contains(strings.ToLower(usr.Name), "yohan") && usr.PhoneNo != nil && *usr.PhoneNo != "" {
						yohanUser = &usr
						break
					}
				}
				if yohanUser != nil {
					yohanPhone := *yohanUser.PhoneNo
					yohanMsg := "🔔 *[NOTIFIKASI LEADER]*\n\n" + waMsg
					err = u.notifUsecase.SendWhatsApp(bgCtx, yohanPhone, yohanMsg, refType, ticket.ID, nil)
					if err != nil {
						log.Printf("Failed to enqueue WA for leader Yohan: %v\n", err)
					} else {
						log.Printf("WA enqueued for leader Yohan (phone: %s)\n", yohanPhone)
					}
				} else {
					log.Printf("WA Warning: User Yohan found but has no phone number\n")
				}
			} else {
				log.Printf("WA Warning: User Yohan not found: %v\n", err)
			}
		}
	}()
}

func (u *troubleTicketUsecase) Update(ctx context.Context, id uint64, ticket *domain.TroubleTicket, userID uint64) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	existing.Title = ticket.Title
	existing.Description = ticket.Description
	existing.Category = ticket.Category
	existing.Priority = ticket.Priority
	existing.Evidence = ticket.Evidence
	existing.UpdatedAt = time.Now()

	if err := u.repo.Update(ctx, existing); err != nil {
		return err
	}

	u.logActivity(ctx, userID, "Update Trouble Ticket", fmt.Sprintf("Updated ticket ID: %d, Number: %s", existing.ID, existing.TicketNumber))
	return nil
}

func (u *troubleTicketUsecase) UpdateStatus(ctx context.Context, id uint64, status domain.TicketStatus, notes string, actionDesc string, problemSummary string, actionSummary string, evidence string, userID uint64) error {
	ticket, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	oldStatus := ticket.Status
	if oldStatus == status {
		return nil
	}

	ticket.Status = status
	ticket.UpdatedAt = time.Now()

	now := time.Now()
	// Auto-set resolved_at
	if status == domain.TicketStatusResolved || status == domain.TicketStatusClosed {
		ticket.ResolvedAt = &now
		if ticket.DowntimeStart != nil && ticket.DowntimeEnd == nil {
			ticket.DowntimeEnd = &now
		}
		if ticket.PendingStart != nil {
			pendingDuration := int(now.Sub(*ticket.PendingStart).Minutes())
			ticket.TotalPendingMinutes += pendingDuration
			ticket.PendingStart = nil
		}
		ticket.UpdateDowntime()
	} else if status == domain.TicketStatusPendingCustomer || status == domain.TicketStatusPendingVendor {
		if ticket.PendingStart == nil {
			ticket.PendingStart = &now
		}
		ticket.UpdateDowntime()
	} else if (oldStatus == domain.TicketStatusPendingCustomer || oldStatus == domain.TicketStatusPendingVendor) && (status == domain.TicketStatusOpen || status == domain.TicketStatusInProgress) {
		if ticket.PendingStart != nil {
			pendingDuration := int(now.Sub(*ticket.PendingStart).Minutes())
			ticket.TotalPendingMinutes += pendingDuration
			ticket.PendingStart = nil
		}
		ticket.UpdateDowntime()
	} else if (status == domain.TicketStatusOpen || status == domain.TicketStatusInProgress) && !(oldStatus == domain.TicketStatusOpen || oldStatus == domain.TicketStatusInProgress) {
		if ticket.DowntimeStart == nil || (ticket.DowntimeEnd != nil && ticket.DowntimeStart.Before(*ticket.DowntimeEnd)) {
			ticket.DowntimeStart = &now
			ticket.DowntimeEnd = nil
			ticket.PendingStart = nil
			ticket.TotalPendingMinutes = 0
			ticket.UpdateDowntime()
		}
	}

	if err := u.repo.Update(ctx, ticket); err != nil {
		return err
	}

	// Create History
	history := &domain.TicketHistory{
		TicketID:  ticket.ID,
		OldStatus: &oldStatus,
		NewStatus: status,
		ChangedBy: &userID,
		Notes:     &notes,
		CreatedAt: time.Now(),
	}
	_ = u.repo.CreateHistory(ctx, history)

	// Create Action Taken if provided
	if actionDesc != "" || problemSummary != "" || actionSummary != "" {
		var evPtr *string
		if evidence != "" {
			evPtr = &evidence
		}
		action := &domain.ActionTaken{
			TicketID:          ticket.ID,
			ActionDescription: actionDesc,
			SummaryProblem:    problemSummary,
			SummaryAction:     actionSummary,
			Evidence:          evPtr,
			Notes:             &notes,
			TakenBy:           &userID,
			CreatedAt:         time.Now(),
		}
		_ = u.repo.CreateActionTaken(ctx, action)
	}

	u.logActivity(ctx, userID, "Update Ticket Status", fmt.Sprintf("Updated status of ticket ID: %d to %s", ticket.ID, string(status)))
	return nil
}

func (u *troubleTicketUsecase) UpdateDowntime(ctx context.Context, id uint64, update map[string]interface{}, userID uint64) error {
	ticket, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if val, ok := update["downtime_start"]; ok {
		if t, ok := val.(time.Time); ok {
			ticket.DowntimeStart = &t
		}
	}
	if val, ok := update["downtime_end"]; ok {
		if t, ok := val.(time.Time); ok {
			ticket.DowntimeEnd = &t
		}
	}

	ticket.UpdateDowntime()
	ticket.UpdatedAt = time.Now()

	if err := u.repo.Update(ctx, ticket); err != nil {
		return err
	}

	u.logActivity(ctx, userID, "Update Ticket Downtime", fmt.Sprintf("Updated downtime of ticket ID: %d", ticket.ID))
	return nil
}

func (u *troubleTicketUsecase) AssignTicket(ctx context.Context, id uint64, assignedTo uint64, notes string, userID uint64) error {
	ticket, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	ticket.AssignedTo = &assignedTo
	ticket.UpdatedAt = time.Now()

	if ticket.Status == domain.TicketStatusOpen {
		oldStatus := ticket.Status
		ticket.Status = domain.TicketStatusInProgress
		
		historyNotes := fmt.Sprintf("Auto-assigned to technician. %s", notes)
		history := &domain.TicketHistory{
			TicketID:  ticket.ID,
			OldStatus: &oldStatus,
			NewStatus: domain.TicketStatusInProgress,
			ChangedBy: &userID,
			Notes:     &historyNotes,
			CreatedAt: time.Now(),
		}
		_ = u.repo.CreateHistory(ctx, history)
	}

	if err := u.repo.Update(ctx, ticket); err != nil {
		return err
	}

	u.logActivity(ctx, userID, "Assign Trouble Ticket", fmt.Sprintf("Assigned ticket ID: %d to technician ID: %d", ticket.ID, assignedTo))

	// Send WA to technician in background
	u.sendTechnicianWA(ctx, assignedTo, ticket)

	return nil
}

func (u *troubleTicketUsecase) AddAction(ctx context.Context, id uint64, action *domain.ActionTaken, userID uint64) error {
	action.TicketID = id
	action.TakenBy = &userID
	action.CreatedAt = time.Now()

	if err := u.repo.CreateActionTaken(ctx, action); err != nil {
		return err
	}

	// Optionally update ticket evidence
	if action.Evidence != nil && *action.Evidence != "" {
		ticket, err := u.repo.GetByID(ctx, id)
		if err == nil {
			ticket.Evidence = action.Evidence
			_ = u.repo.Update(ctx, ticket)
		}
	}

	u.logActivity(ctx, userID, "Add Ticket Action", fmt.Sprintf("Added action taken for ticket ID: %d", id))
	return nil
}

func (u *troubleTicketUsecase) GetHistory(ctx context.Context, ticketID uint64) ([]domain.TicketHistory, error) {
	return u.repo.GetHistory(ctx, ticketID)
}

func (u *troubleTicketUsecase) GetActions(ctx context.Context, ticketID uint64) ([]domain.ActionTaken, error) {
	return u.repo.GetActionsTaken(ctx, ticketID)
}

func (u *troubleTicketUsecase) Delete(ctx context.Context, id uint64, userID uint64) error {
	ticket, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if ticket.Status != domain.TicketStatusResolved && ticket.Status != domain.TicketStatusClosed && ticket.Status != domain.TicketStatusCancelled {
		return errors.New("only resolved, closed, or cancelled tickets can be deleted")
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	u.logActivity(ctx, userID, "Delete Trouble Ticket", fmt.Sprintf("Deleted ticket ID: %d, Number: %s", ticket.ID, ticket.TicketNumber))
	return nil
}

func (u *troubleTicketUsecase) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	return u.repo.GetStatistics(ctx)
}

func (u *troubleTicketUsecase) GetMonthlyTrends(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return u.repo.GetMonthlyTrends(ctx, months)
}

func (u *troubleTicketUsecase) GetCategoryPerformance(ctx context.Context, dateFrom, dateTo *time.Time) ([]map[string]interface{}, error) {
	return u.repo.GetCategoryPerformance(ctx, dateFrom, dateTo)
}

func (u *troubleTicketUsecase) GetUserPerformance(ctx context.Context, dateFrom, dateTo *time.Time) ([]map[string]interface{}, error) {
	return u.repo.GetUserPerformance(ctx, dateFrom, dateTo)
}

func (u *troubleTicketUsecase) GetDowntimeAnalysis(ctx context.Context, dateFrom, dateTo *time.Time) (map[string]interface{}, error) {
	return u.repo.GetDowntimeAnalysis(ctx, dateFrom, dateTo)
}
