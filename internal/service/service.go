package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"serve-swimming-be/config"
	"serve-swimming-be/internal/domain"
	"serve-swimming-be/internal/dto"
	"serve-swimming-be/internal/repository"
	"serve-swimming-be/pkg/security"
)

type Service struct {
	repo *repository.Repository
	cfg  *config.Config
}

func NewService(repo *repository.Repository, cfg *config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

func (s *Service) AuthenticateAdmin(req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.repo.FindUserByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if !security.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid username or password")
	}

	token, err := security.GenerateJWT(user.ID, user.Username, user.Role, s.cfg.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token")
	}

	var permissions []string
	roleName := user.Role
	if user.RoleRel != nil {
		roleName = user.RoleRel.Name
		_ = json.Unmarshal([]byte(user.RoleRel.Permissions), &permissions)
	} else if user.Role == "Super Admin" || user.Role == "ADMIN" {
		permissions = []string{"*"}
	}

	return &dto.LoginResponse{
		Token: token,
		User: dto.UserSummary{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			Role:        roleName,
			RoleID:      user.RoleID,
			Permissions: permissions,
		},
	}, nil
}

func (s *Service) GetBanners() ([]domain.BannerSlide, error) {
	return s.repo.FindBanners()
}

func (s *Service) SaveBanner(banner *domain.BannerSlide) error {
	return s.repo.SaveBanner(banner)
}

func (s *Service) DeleteBanner(id uint) error {
	return s.repo.DeleteBanner(id)
}

func (s *Service) GetEvents() ([]domain.SwimmingEvent, error) {
	return s.repo.FindEvents()
}

func (s *Service) SaveEvent(evt *domain.SwimmingEvent) error {
	if evt.ID > 0 {
		return s.repo.SaveEvent(evt)
	}
	return s.repo.CreateEvent(evt)
}

func (s *Service) DeleteEvent(id uint) error {
	return s.repo.DeleteEvent(id)
}

func (s *Service) RegisterParticipant(req dto.RegisterParticipantRequest) (*dto.RegisterParticipantResponse, error) {
	// 1. Validasi Tanggal Lahir dan Penentuan Kelompok Umur (KU)
	derivedKU := "OPEN"
	if req.BirthDate != "" {
		birthTime, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			return nil, fmt.Errorf("format tanggal lahir tidak valid (gunakan format YYYY-MM-DD)")
		}
		if birthTime.After(time.Now()) {
			return nil, fmt.Errorf("tanggal lahir tidak boleh di masa depan")
		}

		// Perhitungan usia atlet berpatokan pada tahun kejuaraan (2026)
		champYear := 2026
		age := champYear - birthTime.Year()
		if age < 4 || age > 85 {
			return nil, fmt.Errorf("usia atlet (%d tahun) di luar batas kepesertaan yang diperbolehkan (minimal 4 tahun)", age)
		}

		// Kategori Kelompok Umur resmi Akuatik Indonesia (PB PRSI)
		if age <= 9 {
			derivedKU = "KU 5"
		} else if age <= 11 {
			derivedKU = "KU 4"
		} else if age <= 13 {
			derivedKU = "KU 3"
		} else if age <= 15 {
			derivedKU = "KU 2"
		} else if age <= 18 {
			derivedKU = "KU 1"
		} else {
			derivedKU = "Senior"
		}
	}

	if req.AgeGroup == "" {
		req.AgeGroup = derivedKU
	}

	// 2. Validasi Kesesuaian Nomor Lomba (Kelompok Umur & Jenis Kelamin)
	for _, selection := range req.EventSelections {
		event, err := s.repo.FindEventByID(selection.SwimmingEventID)
		if err != nil || event == nil {
			return nil, fmt.Errorf("nomor lomba dengan ID %d tidak ditemukan", selection.SwimmingEventID)
		}

		// Validasi Gender
		evGender := strings.ToUpper(strings.TrimSpace(event.Gender))
		partGender := strings.ToUpper(strings.TrimSpace(req.Gender))
		if evGender != "" && evGender != partGender {
			return nil, fmt.Errorf("nomor lomba '%s' (%s) tidak sesuai dengan jenis kelamin atlet (%s)", event.EventName, event.Gender, req.Gender)
		}

		// Validasi Kelompok Umur
		evKU := strings.ToUpper(strings.TrimSpace(event.AgeGroup))
		partKU := strings.ToUpper(strings.TrimSpace(req.AgeGroup))
		if evKU != "" && evKU != "OPEN" && evKU != "TERBUKA" && evKU != partKU {
			return nil, fmt.Errorf("nomor lomba '%s' (kategori %s) tidak sesuai dengan kelompok umur atlet (%s)", event.EventName, event.AgeGroup, req.AgeGroup)
		}
	}

	// Generate Unique Registration Code matching format REG-ASC-XXXXX
	bytes := make([]byte, 2)
	_, _ = rand.Read(bytes)
	numRand := 10000 + (int(bytes[0])<<8|int(bytes[1]))%90000
	regCode := fmt.Sprintf("REG-ASC-%d", numRand)

	participant := domain.Participant{
		Name:                req.Name,
		Gender:              req.Gender,
		Club:                req.Club,
		PIC:                 req.PIC,
		Contact:             req.Contact,
		Email:               req.Email,
		BirthDate:           req.BirthDate,
		AgeGroup:            req.AgeGroup,
		VerificationDocType: req.VerificationDocType,
		VerificationDocURL:  req.VerificationDocURL,
	}

	if err := s.repo.CreateParticipant(&participant); err != nil {
		return nil, fmt.Errorf("failed to save participant details: %v", err)
	}

	totalEvents := 0
	for _, selection := range req.EventSelections {
		timeSeed := selection.TimeSeed
		if timeSeed == "" || timeSeed == "NT" {
			timeSeed = "99:99.99"
		}

		reg := domain.Registration{
			RegistrationCode: regCode,
			ParticipantID:    participant.ID,
			SwimmingEventID:  selection.SwimmingEventID,
			TimeSeed:         timeSeed,
			PaymentStatus:    "pending",
			PaymentMethod:    req.PaymentMethod,
			SenderBankOwner:  req.SenderBankOwner,
			PaymentProofURL:  req.PaymentProofURL,
		}

		if err := s.repo.CreateRegistration(&reg); err != nil {
			log.Printf("[RegisterParticipant] Failed to insert registration event %d: %v", selection.SwimmingEventID, err)
		} else {
			totalEvents++
		}
	}

	return &dto.RegisterParticipantResponse{
		RegistrationCode: regCode,
		ParticipantName:  participant.Name,
		TotalEvents:      totalEvents,
		PaymentStatus:    "pending",
		Message:          "Pendaftaran berhasil dikirim! Silakan lakukan pembayaran dan periksa status secara berkala.",
	}, nil
}

func (s *Service) GetStartingList(tournamentID uint) ([]dto.StartingItemDTO, error) {
	regs, err := s.repo.FindRegistrations()
	if err != nil {
		return nil, err
	}

	var items []dto.StartingItemDTO
	counter := 1
	for _, r := range regs {
		if tournamentID > 0 && r.SwimmingEvent.TournamentID != tournamentID {
			continue
		}
		items = append(items, dto.StartingItemDTO{
			No:             counter,
			Nama:           r.Participant.Name,
			Gender:         r.Participant.Gender,
			TimeSeed:       r.TimeSeed,
			NomorLomba:     r.SwimmingEvent.EventName,
			Club:           r.Participant.Club,
			PIC:            r.Participant.PIC,
			Kontak:         r.Participant.Contact,
			Result:         r.RaceResultTime,
			Rank:           r.Rank,
			TournamentID:   r.SwimmingEvent.TournamentID,
			TournamentName: r.SwimmingEvent.Tournament.Name,
		})
		counter++
	}
	return items, nil
}

// getSpearheadLanes returns standard swimming spearhead lane assignment order
func getSpearheadLanes(maxLanes int) []int {
	switch maxLanes {
	case 3:
		return []int{2, 1, 3}
	case 4:
		return []int{2, 3, 1, 4}
	case 5:
		return []int{3, 2, 4, 1, 5}
	case 6:
		return []int{3, 4, 2, 5, 1, 6}
	case 8:
		return []int{4, 5, 3, 6, 2, 7, 1, 8}
	case 10:
		return []int{5, 6, 4, 7, 3, 8, 2, 9, 1, 10}
	default:
		lanes := make([]int, maxLanes)
		center := (maxLanes + 1) / 2
		lanes[0] = center
		idx := 1
		for d := 1; idx < maxLanes; d++ {
			if center-d >= 1 {
				lanes[idx] = center - d
				idx++
			}
			if idx < maxLanes && center+d <= maxLanes {
				lanes[idx] = center + d
				idx++
			}
		}
		return lanes
	}
}

// parseTimeSeedToSeconds converts seed time (e.g. 02.31.37 or 99.99.99) into seconds for precise sorting
func parseTimeSeedToSeconds(t string) float64 {
	t = strings.TrimSpace(t)
	if t == "" || t == "NT" || strings.HasPrefix(t, "99") {
		return 999999.0
	}
	t = strings.ReplaceAll(t, ":", ".")
	parts := strings.Split(t, ".")
	if len(parts) >= 3 {
		min, _ := strconv.ParseFloat(parts[0], 64)
		sec, _ := strconv.ParseFloat(parts[1], 64)
		ms, _ := strconv.ParseFloat(parts[2], 64)
		return min*60 + sec + ms/100
	} else if len(parts) == 2 {
		sec, _ := strconv.ParseFloat(parts[0], 64)
		ms, _ := strconv.ParseFloat(parts[1], 64)
		return sec + ms/100
	}
	val, err := strconv.ParseFloat(t, 64)
	if err == nil {
		return val
	}
	return 999999.0
}

// GetGroupLabel converts 1-based index (1, 2, 3...) to alphabet group (A, B, C... Z, AA...)
func GetGroupLabel(num int) string {
	if num <= 0 {
		return ""
	}
	result := ""
	for num > 0 {
		num--
		result = string(rune('A'+(num%26))) + result
		num /= 26
	}
	return result
}

// GetClusterLabel is alias of GetGroupLabel for backwards-compatibility
func GetClusterLabel(num int) string {
	return GetGroupLabel(num)
}

// GenerateBukuAcara: Heat & Line tournament assignment algorithm (ONLY verified participants)
func (s *Service) GenerateBukuAcara(maxLanes int, tournamentID uint, force bool) error {
	if maxLanes <= 0 {
		maxLanes = 3
	}

	// 0. If tournament is locked and not forced, reject
	if tournamentID > 0 {
		t, err := s.repo.FindTournamentByID(tournamentID)
		if err == nil && t != nil && t.IsBukuAcaraLocked && !force {
			return fmt.Errorf("Buku Acara untuk turnamen '%s' telah dipatenkan/dikunci. Silakan buka kunci terlebih dahulu untuk generate ulang", t.Name)
		}
	}

	_ = s.repo.UpdatePoolConfigMaxLanes(maxLanes)

	// 1. Reset heat & line numbers for this tournament
	_ = s.repo.ResetTournamentHeatLines(tournamentID)

	events, err := s.repo.FindEvents()
	if err != nil {
		return err
	}

	// 2. Fetch all registrations and filter only verified ones
	regs, err := s.repo.FindRegistrations()
	if err != nil {
		return err
	}

	eventGroup := make(map[uint][]domain.Registration)
	for _, r := range regs {
		if strings.ToLower(r.PaymentStatus) == "verified" {
			eventGroup[r.SwimmingEventID] = append(eventGroup[r.SwimmingEventID], r)
		}
	}

	spearhead := getSpearheadLanes(maxLanes)

	for _, event := range events {
		if tournamentID > 0 && event.TournamentID != tournamentID {
			continue
		}

		eventRegs := eventGroup[event.ID]
		if len(eventRegs) == 0 {
			continue
		}

		total := len(eventRegs)
		numHeats := (total + maxLanes - 1) / maxLanes
		if numHeats == 0 {
			numHeats = 1
		}

		categoryUpper := strings.ToUpper(strings.TrimSpace(event.HeatCategory))
		isGroup := categoryUpper == "GROUP" || categoryUpper == "CLUSTER"

		if isGroup {
			// ==============================================================
			// GROUP ABJAD ALGORITHM (A, B, C, D...):
			// Diurut dari seed time terlambat peserta sampai tercepat!
			// Group A = paling lambat (NT, 99.99.99, catatan waktu tertinggi)
			// Group B, C, D... = kecepatan meningkat bertahap hingga tercepat
			// ==============================================================

			// Sort by TimeSeed descending (slowest first: NT/99.99.99 and highest seconds first)
			sort.Slice(eventRegs, func(i, j int) bool {
				ti := parseTimeSeedToSeconds(eventRegs[i].TimeSeed)
				tj := parseTimeSeedToSeconds(eventRegs[j].TimeSeed)
				if ti == tj {
					return eventRegs[i].ID < eventRegs[j].ID
				}
				return ti > tj // ti > tj means higher seconds (slower) comes first
			})

			// Calculate group sizes (balance so no group has only 1 swimmer if total >= 2)
			groupSizes := make([]int, numHeats)
			base := total / numHeats
			rem := total % numHeats
			for i := 0; i < numHeats; i++ {
				groupSizes[i] = base
			}
			// Distribute remainder 1 by 1 starting from Group 1 (Group A)
			for i := 0; i < rem; i++ {
				groupSizes[i]++
			}

			offset := 0
			// Group 1 (Group A) gets the slowest swimmers, Group numHeats gets the fastest swimmers
			for c := 1; c <= numHeats; c++ {
				size := groupSizes[c-1]
				groupSwimmers := eventRegs[offset : offset+size]
				offset += size

				// Within this group, assign lanes according to spearhead sequence
				// (The faster swimmers in this group get center lanes)
				sort.Slice(groupSwimmers, func(i, j int) bool {
					ti := parseTimeSeedToSeconds(groupSwimmers[i].TimeSeed)
					tj := parseTimeSeedToSeconds(groupSwimmers[j].TimeSeed)
					if ti == tj {
						return groupSwimmers[i].ID < groupSwimmers[j].ID
					}
					return ti < tj // faster swimmers first inside this group
				})

				for sIdx, reg := range groupSwimmers {
					if sIdx < len(spearhead) {
						lineNum := spearhead[sIdx]
						_ = s.repo.UpdateRegistrationHeatLine(reg.ID, c, lineNum)
					}
				}
			}
		} else {
			// ==============================================================
			// EXISTING HEAT ALGORITHM (Seri 1, 2, 3...):
			// Diurut dari tercepat ke terlambat.
			// Seri terakhir (final heat) = perenang tercepat.
			// ==============================================================

			// Sort by TimeSeed ascending (faster times first: Rank 1 at index 0, 99.99.99 / NT last)
			sort.Slice(eventRegs, func(i, j int) bool {
				ti := parseTimeSeedToSeconds(eventRegs[i].TimeSeed)
				tj := parseTimeSeedToSeconds(eventRegs[j].TimeSeed)
				if ti == tj {
					return eventRegs[i].ID < eventRegs[j].ID
				}
				return ti < tj
			})

			// Calculate heat sizes so NO HEAT has only 1 swimmer (minimum 2 swimmers per heat if total >= 2)
			heatSizes := make([]int, numHeats)
			base := total / numHeats
			rem := total % numHeats
			for i := 0; i < numHeats; i++ {
				heatSizes[i] = base
			}
			// Distribute remainder 1 by 1 to the fastest heats (from heat numHeats downwards)
			for i := 0; i < rem; i++ {
				heatSizes[numHeats-1-i]++
			}

			// Map swimmers to heats:
			// Heat numHeats (final heat) gets the fastest swimmers (eventRegs[0 : heatSizes[numHeats-1]])
			// Earlier heats get subsequent swimmers down to Heat 1 (slowest swimmers)
			offset := 0
			for h := numHeats; h >= 1; h-- {
				size := heatSizes[h-1]
				heatSwimmers := eventRegs[offset : offset+size]
				offset += size

				// In this heat, assign lines according to spearhead sequence [2, 1, 3...]
				for sIdx, reg := range heatSwimmers {
					if sIdx < len(spearhead) {
						lineNum := spearhead[sIdx]
						_ = s.repo.UpdateRegistrationHeatLine(reg.ID, h, lineNum)
					}
				}
			}
		}
	}

	return nil
}

func (s *Service) GetBukuAcara(tournamentID uint, round string) ([]dto.BukuAcaraEventGroupDTO, error) {
	poolCfg, _ := s.repo.FindPoolConfig()
	maxLanes := 3
	if poolCfg != nil && poolCfg.MaxLanes > 0 {
		maxLanes = poolCfg.MaxLanes
	}

	events, err := s.repo.FindEvents()
	if err != nil {
		return nil, err
	}

	regs, err := s.repo.FindRegistrations()
	if err != nil {
		return nil, err
	}

	isFinal := strings.ToLower(strings.TrimSpace(round)) == "final"

	regMap := make(map[uint][]domain.Registration)
	hasFinalistsMap := make(map[uint]bool)

	for _, r := range regs {
		if strings.ToLower(r.PaymentStatus) == "verified" {
			if r.IsFinalist && r.FinalHeatNumber > 0 && r.FinalLineNumber > 0 {
				hasFinalistsMap[r.SwimmingEventID] = true
			}

			if isFinal {
				// Babak Final: hanya peserta lolos final dengan heat dan line final
				if r.IsFinalist && r.FinalHeatNumber > 0 && r.FinalLineNumber > 0 {
					regMap[r.SwimmingEventID] = append(regMap[r.SwimmingEventID], r)
				}
			} else {
				// Putaran Awal (Penyisihan): peserta terverifikasi dengan heat dan line
				if r.HeatNumber > 0 && r.LineNumber > 0 {
					regMap[r.SwimmingEventID] = append(regMap[r.SwimmingEventID], r)
				}
			}
		}
	}

	var result []dto.BukuAcaraEventGroupDTO
	for _, ev := range events {
		if tournamentID > 0 && ev.TournamentID != tournamentID {
			continue
		}

		eRegs := regMap[ev.ID]
		if len(eRegs) == 0 {
			continue
		}

		categoryUpper := strings.ToUpper(strings.TrimSpace(ev.HeatCategory))
		isGroup := categoryUpper == "GROUP" || categoryUpper == "CLUSTER"
		heatCategoryStr := "HEAT"
		if isGroup {
			heatCategoryStr = "GROUP"
		}

		maxHeat := 1
		slotMap := make(map[string]domain.Registration)
		for _, r := range eRegs {
			heatNum := r.HeatNumber
			lineNum := r.LineNumber
			if isFinal {
				heatNum = r.FinalHeatNumber
				lineNum = r.FinalLineNumber
			}

			if heatNum > maxHeat {
				maxHeat = heatNum
			}
			key := fmt.Sprintf("%d-%d", heatNum, lineNum)
			slotMap[key] = r
		}

		var heatItems []dto.BukuAcaraHeatItemDTO
		for h := 1; h <= maxHeat; h++ {
			heatLabel := fmt.Sprintf("%d", h)
			if isGroup {
				heatLabel = GetGroupLabel(h)
			}

			for l := 1; l <= maxLanes; l++ {
				key := fmt.Sprintf("%d-%d", h, l)
				if r, exists := slotMap[key]; exists {
					res := r.RaceResultTime
					rnk := r.Rank
					timeSeed := r.TimeSeed

					if isFinal {
						timeSeed = r.RaceResultTime // Seed time di babak final adalah hasil putaran awal!
						res = r.FinalResultTime
						rnk = r.FinalRank
					}

					if res == "" && rnk > 0 {
						res = fmt.Sprintf("Rank %d", rnk)
					}
					if res == "" {
						res = "-"
					}

					heatItems = append(heatItems, dto.BukuAcaraHeatItemDTO{
						RegistrationID:    r.ID,
						Heat:              h,
						HeatLabel:         heatLabel,
						HeatCategory:      heatCategoryStr,
						Line:              l,
						Nama:              r.Participant.Name,
						Gender:            r.Participant.Gender,
						Club:              r.Participant.Club,
						TimeSeed:          timeSeed,
						Result:            res,
						Rank:              rnk,
						IsEmpty:           false,
						IsFinalist:        r.IsFinalist,
						PreliminaryResult: r.RaceResultTime,
						PreliminaryRank:   r.Rank,
					})
				} else {
					heatItems = append(heatItems, dto.BukuAcaraHeatItemDTO{
						Heat:         h,
						HeatLabel:    heatLabel,
						HeatCategory: heatCategoryStr,
						Line:         l,
						Nama:         "(KOSONG)",
						Gender:       "-",
						Club:         "-",
						TimeSeed:     "-",
						Result:       "-",
						IsEmpty:      true,
					})
				}
			}
		}

		roundStr := "preliminary"
		if isFinal {
			roundStr = "final"
		}

		result = append(result, dto.BukuAcaraEventGroupDTO{
			EventID:      ev.ID,
			TournamentID: ev.TournamentID,
			EventCode:    ev.EventCode,
			EventName:    ev.EventName,
			Distance:     ev.Distance,
			Stroke:       ev.Stroke,
			Gender:       ev.Gender,
			AgeGroup:     ev.AgeGroup,
			HeatCategory: heatCategoryStr,
			MaxLanes:     maxLanes,
			Round:        roundStr,
			HasFinalists: hasFinalistsMap[ev.ID],
			Heats:        heatItems,
		})
	}

	return result, nil
}

func (s *Service) GetRegistrations() ([]domain.Registration, error) {
	return s.repo.FindRegistrations()
}

func (s *Service) GetRegistrationByCode(code string) ([]domain.Registration, error) {
	return s.repo.FindRegistrationByCode(code)
}

func (s *Service) VerifyPayment(id uint, status string) error {
	if strings.ToLower(status) != "verified" {
		_ = s.repo.UpdateRegistrationHeatLine(id, 0, 0)
	}
	return s.repo.UpdateRegistrationStatus(id, status)
}

// RecordRaceResult records result with lock enforcement
func (s *Service) RecordRaceResult(id uint, timeStr string, rank int, round string, operatorName, ipAddress string) error {
	if operatorName == "" {
		operatorName = "Admin Panitia"
	}

	reg, err := s.repo.FindRegistrationByID(id)
	if err != nil {
		return fmt.Errorf("Data perenang tidak ditemukan: %w", err)
	}

	// 1. Enforce lock check on tournament: Buku Acara must be locked before recording results!
	if reg.SwimmingEvent.TournamentID > 0 {
		tourney, err := s.repo.FindTournamentByID(reg.SwimmingEvent.TournamentID)
		if err == nil && tourney != nil && !tourney.IsBukuAcaraLocked {
			return fmt.Errorf("Buku Acara untuk turnamen '%s' belum dipatenkan/dikunci. Kunci Buku Acara terlebih dahulu sebelum mencatat hasil lomba", tourney.Name)
		}
	}

	isFinal := strings.ToLower(strings.TrimSpace(round)) == "final"

	oldTime := reg.RaceResultTime
	oldRank := reg.Rank
	heatNum := reg.HeatNumber
	lineNum := reg.LineNumber
	roundLabel := "Putaran Awal"

	if isFinal {
		oldTime = reg.FinalResultTime
		oldRank = reg.FinalRank
		heatNum = reg.FinalHeatNumber
		lineNum = reg.FinalLineNumber
		roundLabel = "Babak Final"
	}

	// Determine Action: CREATE, UPDATE, DELETE
	action := "UPDATE"
	actionDesc := "Memperbarui"
	if strings.TrimSpace(oldTime) == "" && strings.TrimSpace(timeStr) != "" {
		action = "CREATE"
		actionDesc = "Mencatat baru"
	} else if strings.TrimSpace(oldTime) != "" && strings.TrimSpace(timeStr) == "" {
		action = "DELETE"
		actionDesc = "Menghapus"
	}

	// Execute update
	if isFinal {
		if err := s.repo.UpdateFinalRaceResult(id, timeStr, rank); err != nil {
			return err
		}
	} else {
		if err := s.repo.UpdateRaceResult(id, timeStr, rank); err != nil {
			return err
		}
	}

	// Record audit log
	oldValText := fmt.Sprintf("Waktu: %s, Rank: %d", oldTime, oldRank)
	if oldTime == "" {
		oldValText = "(Kosong / Belum Ada)"
	}
	newValText := fmt.Sprintf("Waktu: %s, Rank: %d", timeStr, rank)
	if timeStr == "" {
		newValText = "(Dihapus)"
	}

	notes := fmt.Sprintf("%s hasil lomba untuk perenang %s pada %s - %s (Seri %d, Line %d). Sebelumnya: %s -> Sekarang: %s",
		actionDesc, reg.Participant.Name, reg.SwimmingEvent.EventName, roundLabel, heatNum, lineNum, oldValText, newValText)

	_ = s.repo.CreateRaceResultLog(&domain.RaceResultLog{
		TournamentID:   reg.SwimmingEvent.TournamentID,
		RegistrationID: reg.ID,
		SwimmerName:    reg.Participant.Name,
		ClubName:       reg.Participant.Club,
		EventCode:      reg.SwimmingEvent.EventCode,
		EventName:      reg.SwimmingEvent.EventName,
		Round:          roundLabel,
		HeatNumber:     heatNum,
		LineNumber:     lineNum,
		Action:         action,
		OldValue:       oldValText,
		NewValue:       newValText,
		OperatorName:   operatorName,
		Notes:          notes,
		IPAddress:      ipAddress,
		CreatedAt:      time.Now(),
	})

	return nil
}

func (s *Service) SetTournamentBukuAcaraLock(tournamentID uint, isLocked bool, operatorName, ipAddress string) error {
	if operatorName == "" {
		operatorName = "Admin Panitia"
	}

	tourney, err := s.repo.FindTournamentByID(tournamentID)
	if err != nil {
		return err
	}

	if err := s.repo.SetTournamentBukuAcaraLock(tournamentID, isLocked); err != nil {
		return err
	}

	action := "LOCK"
	actionDesc := "Mengunci / mematenkan Buku Acara turnamen"
	if !isLocked {
		action = "UNLOCK"
		actionDesc = "Membuka kunci Buku Acara turnamen"
	}

	_ = s.repo.CreateRaceResultLog(&domain.RaceResultLog{
		TournamentID: tournamentID,
		Action:       action,
		OldValue:     fmt.Sprintf("Terkunci: %v", tourney.IsBukuAcaraLocked),
		NewValue:     fmt.Sprintf("Terkunci: %v", isLocked),
		OperatorName: operatorName,
		Notes:        fmt.Sprintf("%s: %s", actionDesc, tourney.Name),
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
	})

	return nil
}

func (s *Service) SetTournamentBukuAcaraPublish(tournamentID uint, isPublished bool, operatorName, ipAddress string) error {
	if operatorName == "" {
		operatorName = "Admin Panitia"
	}

	tourney, err := s.repo.FindTournamentByID(tournamentID)
	if err != nil {
		return err
	}

	if err := s.repo.SetTournamentBukuAcaraPublish(tournamentID, isPublished); err != nil {
		return err
	}

	action := "PUBLISH"
	actionDesc := "Mempublikasikan Buku Acara ke halaman publik"
	if !isPublished {
		action = "UNPUBLISH"
		actionDesc = "Menarik publikasi Buku Acara dari halaman publik (Draft)"
	}

	_ = s.repo.CreateRaceResultLog(&domain.RaceResultLog{
		TournamentID: tournamentID,
		Action:       action,
		OldValue:     fmt.Sprintf("Terpublikasi: %v", tourney.IsBukuAcaraPublished),
		NewValue:     fmt.Sprintf("Terpublikasi: %v", isPublished),
		OperatorName: operatorName,
		Notes:        fmt.Sprintf("%s: %s", actionDesc, tourney.Name),
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
	})

	return nil
}

func (s *Service) SwapRegistrationHeatLine(id uint, targetHeat, targetLine int, swapIfOccupied bool, round string, operatorName, ipAddress string) (*domain.Registration, *domain.Registration, error) {
	if operatorName == "" {
		operatorName = "Admin Panitia"
	}

	if targetHeat <= 0 || targetLine <= 0 {
		return nil, nil, fmt.Errorf("Nomor seri (Heat) dan lintasan (Line) harus lebih besar dari 0")
	}

	regA, err := s.repo.FindRegistrationByID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("Data perenang tidak ditemukan: %w", err)
	}

	isFinal := strings.ToLower(strings.TrimSpace(round)) == "final"
	roundLabel := "Putaran Awal"
	if isFinal {
		roundLabel = "Babak Final"
	}

	oldHeat := regA.HeatNumber
	oldLine := regA.LineNumber
	if isFinal {
		oldHeat = regA.FinalHeatNumber
		oldLine = regA.FinalLineNumber
	}

	// Same position, no change
	if oldHeat == targetHeat && oldLine == targetLine {
		return regA, nil, nil
	}

	// Check if target is occupied by another swimmer in the same swimming event
	var regB *domain.Registration
	var occErr error
	if isFinal {
		regB, occErr = s.repo.FindOccupiedFinalRegistration(regA.SwimmingEventID, targetHeat, targetLine, regA.ID)
	} else {
		regB, occErr = s.repo.FindOccupiedRegistration(regA.SwimmingEventID, targetHeat, targetLine, regA.ID)
	}
	occupied := (occErr == nil && regB != nil && regB.ID > 0)

	if occupied {
		if !swapIfOccupied {
			return nil, nil, fmt.Errorf("Lintasan %d pada Seri %d sudah ditempati oleh %s", targetLine, targetHeat, regB.Participant.Name)
		}
		// Swap regB into regA's old position
		if isFinal {
			if err := s.repo.UpdateRegistrationFinalHeatLine(regB.ID, oldHeat, oldLine); err != nil {
				return nil, nil, fmt.Errorf("Gagal memindahkan perenang di posisi tujuan: %w", err)
			}
			regB.FinalHeatNumber = oldHeat
			regB.FinalLineNumber = oldLine
		} else {
			if err := s.repo.UpdateRegistrationHeatLine(regB.ID, oldHeat, oldLine); err != nil {
				return nil, nil, fmt.Errorf("Gagal memindahkan perenang lain: %w", err)
			}
			regB.HeatNumber = oldHeat
			regB.LineNumber = oldLine
		}
	}

	// Move regA to target position
	if isFinal {
		if err := s.repo.UpdateRegistrationFinalHeatLine(regA.ID, targetHeat, targetLine); err != nil {
			return nil, nil, fmt.Errorf("Gagal memindahkan perenang ke posisi baru: %w", err)
		}
		regA.FinalHeatNumber = targetHeat
		regA.FinalLineNumber = targetLine
	} else {
		if err := s.repo.UpdateRegistrationHeatLine(regA.ID, targetHeat, targetLine); err != nil {
			return nil, nil, fmt.Errorf("Gagal memindahkan perenang ke posisi baru: %w", err)
		}
		regA.HeatNumber = targetHeat
		regA.LineNumber = targetLine
	}

	var swappedB *domain.Registration = nil
	if occupied {
		swappedB = regB

		// Audit log for SWAP
		_ = s.repo.CreateRaceResultLog(&domain.RaceResultLog{
			TournamentID:   regA.SwimmingEvent.TournamentID,
			RegistrationID: regA.ID,
			SwimmerName:    regA.Participant.Name,
			ClubName:       regA.Participant.Club,
			EventCode:      regA.SwimmingEvent.EventCode,
			EventName:      regA.SwimmingEvent.EventName,
			Round:          roundLabel,
			HeatNumber:     targetHeat,
			LineNumber:     targetLine,
			Action:         "SWAP",
			OldValue:       fmt.Sprintf("Seri %d, Line %d", oldHeat, oldLine),
			NewValue:       fmt.Sprintf("Seri %d, Line %d", targetHeat, targetLine),
			OperatorName:   operatorName,
			Notes:          fmt.Sprintf("Tukar posisi lintasan antara %s (ke Seri %d Line %d) dan %s (ke Seri %d Line %d) pada %s (%s)", regA.Participant.Name, targetHeat, targetLine, regB.Participant.Name, oldHeat, oldLine, regA.SwimmingEvent.EventName, roundLabel),
			IPAddress:      ipAddress,
			CreatedAt:      time.Now(),
		})
	} else {
		// Audit log for MOVE to empty line
		_ = s.repo.CreateRaceResultLog(&domain.RaceResultLog{
			TournamentID:   regA.SwimmingEvent.TournamentID,
			RegistrationID: regA.ID,
			SwimmerName:    regA.Participant.Name,
			ClubName:       regA.Participant.Club,
			EventCode:      regA.SwimmingEvent.EventCode,
			EventName:      regA.SwimmingEvent.EventName,
			Round:          roundLabel,
			HeatNumber:     targetHeat,
			LineNumber:     targetLine,
			Action:         "MOVE",
			OldValue:       fmt.Sprintf("Seri %d, Line %d", oldHeat, oldLine),
			NewValue:       fmt.Sprintf("Seri %d, Line %d", targetHeat, targetLine),
			OperatorName:   operatorName,
			Notes:          fmt.Sprintf("Pindah posisi lintasan %s dari Seri %d Line %d ke Seri %d Line %d (Lintasan Kosong) pada %s (%s)", regA.Participant.Name, oldHeat, oldLine, targetHeat, targetLine, regA.SwimmingEvent.EventName, roundLabel),
			IPAddress:      ipAddress,
			CreatedAt:      time.Now(),
		})
	}

	return regA, swappedB, nil
}

// GenerateFinalRound calculates finalists from preliminary heat winners + next fastest times
func (s *Service) GenerateFinalRound(tournamentID uint, maxLanes int, qualifyMode string, operatorName, ipAddress string) error {
	if maxLanes <= 0 {
		poolCfg, _ := s.repo.FindPoolConfig()
		if poolCfg != nil && poolCfg.MaxLanes > 0 {
			maxLanes = poolCfg.MaxLanes
		} else {
			maxLanes = 3
		}
	}

	// Reset existing final round assignments for this tournament
	_ = s.repo.ResetTournamentFinalHeatLines(tournamentID)

	events, err := s.repo.FindEvents()
	if err != nil {
		return err
	}

	regs, err := s.repo.FindRegistrations()
	if err != nil {
		return err
	}

	eventMap := make(map[uint][]domain.Registration)
	for _, r := range regs {
		if strings.ToLower(r.PaymentStatus) == "verified" && r.HeatNumber > 0 {
			eventMap[r.SwimmingEventID] = append(eventMap[r.SwimmingEventID], r)
		}
	}

	spearhead := getSpearheadLanes(maxLanes)

	for _, ev := range events {
		if tournamentID > 0 && ev.TournamentID != tournamentID {
			continue
		}

		eRegs := eventMap[ev.ID]
		if len(eRegs) == 0 {
			continue
		}

		// Collect registrations that have race_result_time recorded
		var withResults []domain.Registration
		for _, r := range eRegs {
			res := strings.TrimSpace(r.RaceResultTime)
			if res != "" && res != "-" {
				withResults = append(withResults, r)
			}
		}

		if len(withResults) == 0 {
			continue
		}

		// Qualification: Juara di setiap heat (Rank 1 / best time) + perenang tercepat berikutnya
		heatGroup := make(map[int][]domain.Registration)
		for _, r := range withResults {
			heatGroup[r.HeatNumber] = append(heatGroup[r.HeatNumber], r)
		}

		for h := range heatGroup {
			sort.Slice(heatGroup[h], func(i, j int) bool {
				ti := parseTimeSeedToSeconds(heatGroup[h][i].RaceResultTime)
				tj := parseTimeSeedToSeconds(heatGroup[h][j].RaceResultTime)
				return ti < tj
			})
		}

		var finalists []domain.Registration
		usedIDs := make(map[uint]bool)

		// 1. Juara di setiap heat (Pemenang Heat 1, Heat 2, dst.)
		var heatNums []int
		for h := range heatGroup {
			heatNums = append(heatNums, h)
		}
		sort.Ints(heatNums)

		for _, h := range heatNums {
			if len(heatGroup[h]) > 0 {
				winner := heatGroup[h][0]
				finalists = append(finalists, winner)
				usedIDs[winner.ID] = true
				if len(finalists) >= maxLanes {
					break
				}
			}
		}

		// 2. Perenang tercepat berikutnya untuk mengisi sisa lintasan final
		if len(finalists) < maxLanes {
			var remaining []domain.Registration
			for _, r := range withResults {
				if !usedIDs[r.ID] {
					remaining = append(remaining, r)
				}
			}
			sort.Slice(remaining, func(i, j int) bool {
				ti := parseTimeSeedToSeconds(remaining[i].RaceResultTime)
				tj := parseTimeSeedToSeconds(remaining[j].RaceResultTime)
				return ti < tj
			})

			for _, r := range remaining {
				finalists = append(finalists, r)
				usedIDs[r.ID] = true
				if len(finalists) >= maxLanes {
					break
				}
			}
		}

		// 3. Urutkan semua finalis berdasarkan waktu putaran awal mereka
		sort.Slice(finalists, func(i, j int) bool {
			ti := parseTimeSeedToSeconds(finalists[i].RaceResultTime)
			tj := parseTimeSeedToSeconds(finalists[j].RaceResultTime)
			return ti < tj
		})

		// 4. Tempatkan ke Final Heat 1 menggunakan rumus spearhead (lintasan tengah untuk ranking 1)
		for rankIdx, finalist := range finalists {
			if rankIdx < len(spearhead) {
				lane := spearhead[rankIdx]
				_ = s.repo.UpdateRegistrationFinalHeatLine(finalist.ID, 1, lane)
			}
		}
	}

	if operatorName == "" {
		operatorName = "Admin Panitia"
	}
	_ = s.repo.CreateRaceResultLog(&domain.RaceResultLog{
		TournamentID: tournamentID,
		Action:       "GENERATE_FINAL",
		OldValue:     "-",
		NewValue:     fmt.Sprintf("Maksimal Lintasan: %d, Mode: %s", maxLanes, qualifyMode),
		OperatorName: operatorName,
		Notes:        fmt.Sprintf("Sistem mengkalkulasi juara heat dan waktu terbaik, serta menyusun bagan babak final kejuaraan turnamen ID %d", tournamentID),
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
	})

	return nil
}

func (s *Service) GetPoolConfig() (*domain.PoolConfig, error) {
	return s.repo.FindPoolConfig()
}

func (s *Service) GetHeroConfig() (*domain.HeroConfig, error) {
	return s.repo.FindHeroConfig()
}

func (s *Service) SaveHeroConfig(cfg *domain.HeroConfig) error {
	return s.repo.SaveHeroConfig(cfg)
}

func (s *Service) GetHeroStats() ([]domain.HeroStat, error) {
	return s.repo.FindHeroStats()
}

func (s *Service) SaveHeroStats(stats []domain.HeroStat) error {
	return s.repo.SaveHeroStats(stats)
}

func (s *Service) GetSiteConfig() (*domain.SiteConfig, error) {
	return s.repo.FindSiteConfig()
}

func (s *Service) SaveSiteConfig(cfg *domain.SiteConfig) error {
	return s.repo.SaveSiteConfig(cfg)
}

func (s *Service) GetProgramSectionConfig() (*domain.ProgramSectionConfig, error) {
	return s.repo.FindProgramSectionConfig()
}

func (s *Service) SaveProgramSectionConfig(cfg *domain.ProgramSectionConfig) error {
	return s.repo.SaveProgramSectionConfig(cfg)
}

func (s *Service) GetTrainingPrograms() ([]domain.TrainingProgram, error) {
	return s.repo.FindTrainingPrograms()
}

func (s *Service) SaveTrainingProgram(prog *domain.TrainingProgram) error {
	return s.repo.SaveTrainingProgram(prog)
}

func (s *Service) DeleteTrainingProgram(id uint) error {
	return s.repo.DeleteTrainingProgram(id)
}

func (s *Service) GetCoachSectionConfig() (*domain.CoachSectionConfig, error) {
	return s.repo.FindCoachSectionConfig()
}

func (s *Service) SaveCoachSectionConfig(cfg *domain.CoachSectionConfig) error {
	return s.repo.SaveCoachSectionConfig(cfg)
}

func (s *Service) GetCoaches() ([]domain.Coach, error) {
	return s.repo.FindCoaches()
}

func (s *Service) SaveCoach(c *domain.Coach) error {
	return s.repo.SaveCoach(c)
}

func (s *Service) DeleteCoach(id uint) error {
	return s.repo.DeleteCoach(id)
}

func (s *Service) GetFacilitySectionConfig() (*domain.FacilitySectionConfig, error) {
	return s.repo.FindFacilitySectionConfig()
}

func (s *Service) SaveFacilitySectionConfig(cfg *domain.FacilitySectionConfig) error {
	return s.repo.SaveFacilitySectionConfig(cfg)
}

func (s *Service) GetFacilities() ([]domain.Facility, error) {
	return s.repo.FindFacilities()
}

func (s *Service) SaveFacility(f *domain.Facility) error {
	return s.repo.SaveFacility(f)
}

func (s *Service) DeleteFacility(id uint) error {
	return s.repo.DeleteFacility(id)
}

// Achievements
func (s *Service) GetAchievementSectionConfig() (*domain.AchievementSectionConfig, error) {
	return s.repo.FindAchievementSectionConfig()
}

func (s *Service) SaveAchievementSectionConfig(cfg *domain.AchievementSectionConfig) error {
	return s.repo.SaveAchievementSectionConfig(cfg)
}

func (s *Service) GetAchievements() ([]domain.Achievement, error) {
	return s.repo.FindAchievements()
}

func (s *Service) SaveAchievement(a *domain.Achievement) error {
	return s.repo.SaveAchievement(a)
}

func (s *Service) DeleteAchievement(id uint) error {
	return s.repo.DeleteAchievement(id)
}

// Testimonials
func (s *Service) GetTestimonialSectionConfig() (*domain.TestimonialSectionConfig, error) {
	return s.repo.FindTestimonialSectionConfig()
}

func (s *Service) SaveTestimonialSectionConfig(cfg *domain.TestimonialSectionConfig) error {
	return s.repo.SaveTestimonialSectionConfig(cfg)
}

func (s *Service) GetTestimonials() ([]domain.Testimonial, error) {
	return s.repo.FindTestimonials()
}

func (s *Service) SaveTestimonial(t *domain.Testimonial) error {
	return s.repo.SaveTestimonial(t)
}

func (s *Service) DeleteTestimonial(id uint) error {
	return s.repo.DeleteTestimonial(id)
}

// Tournaments
func (s *Service) GetTournaments() ([]domain.Tournament, error) {
	return s.repo.FindTournaments()
}

func (s *Service) SaveTournament(t *domain.Tournament) error {
	return s.repo.SaveTournament(t)
}

func (s *Service) DeleteTournament(id uint) error {
	return s.repo.DeleteTournament(id)
}

// Page Sections (Landing Page Dynamic Ordering)
func (s *Service) GetPageSections(pageSlug ...string) ([]domain.PageSection, error) {
	slug := "homepage"
	if len(pageSlug) > 0 && pageSlug[0] != "" {
		slug = pageSlug[0]
	}
	return s.repo.FindPageSections(slug)
}

func (s *Service) SavePageSection(sec *domain.PageSection) error {
	return s.repo.SavePageSection(sec)
}

func (s *Service) BatchSavePageSections(sections []domain.PageSection) error {
	return s.repo.BatchSavePageSections(sections)
}

func (s *Service) ResetPageSections(pageSlug ...string) error {
	slug := "homepage"
	if len(pageSlug) > 0 && pageSlug[0] != "" {
		slug = pageSlug[0]
	}
	return s.repo.ResetPageSections(slug)
}

// ----------------------------------------------------
// RACE RESULT AUDIT LOG METHODS
// ----------------------------------------------------

func (s *Service) GetRaceResultLogs(tournamentID uint, round string, action string, search string, limit, offset int) ([]domain.RaceResultLog, int64, error) {
	return s.repo.FindRaceResultLogs(tournamentID, round, action, search, limit, offset)
}

func (s *Service) GetRaceResultLogStats(tournamentID uint) (map[string]interface{}, error) {
	return s.repo.GetRaceResultLogStats(tournamentID)
}

// ----------------------------------------------------
// ROLE MANAGEMENT SERVICE METHODS
// ----------------------------------------------------

func (s *Service) GetRoles() ([]dto.RoleResponse, error) {
	roles, err := s.repo.FindRoles()
	if err != nil {
		return nil, err
	}

	var results []dto.RoleResponse
	for _, r := range roles {
		var perms []string
		_ = json.Unmarshal([]byte(r.Permissions), &perms)

		userCount, _ := s.repo.CountUsersByRoleID(r.ID)

		results = append(results, dto.RoleResponse{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Permissions: perms,
			IsSystem:    r.IsSystem,
			UsersCount:  int(userCount),
			CreatedAt:   r.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   r.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return results, nil
}

func (s *Service) CreateRole(req dto.RoleRequest) (*dto.RoleResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("nama role tidak boleh kosong")
	}

	permBytes, err := json.Marshal(req.Permissions)
	if err != nil {
		permBytes = []byte("[]")
	}

	role := domain.Role{
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Permissions: string(permBytes),
		IsSystem:    false,
	}

	if err := s.repo.CreateRole(&role); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, fmt.Errorf("nama role sudah terdaftar")
		}
		return nil, err
	}

	return &dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: req.Permissions,
		IsSystem:    role.IsSystem,
		UsersCount:  0,
		CreatedAt:   role.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *Service) UpdateRole(id uint, req dto.RoleRequest) (*dto.RoleResponse, error) {
	role, err := s.repo.FindRoleByID(id)
	if err != nil {
		return nil, fmt.Errorf("role tidak ditemukan")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("nama role tidak boleh kosong")
	}

	// Protect Super Admin system role
	if role.IsSystem {
		role.Description = strings.TrimSpace(req.Description)
		// Super Admin always retains all permissions
		role.Permissions = `["*"]`
	} else {
		role.Name = name
		role.Description = strings.TrimSpace(req.Description)
		permBytes, _ := json.Marshal(req.Permissions)
		role.Permissions = string(permBytes)
	}

	if err := s.repo.UpdateRole(role); err != nil {
		return nil, err
	}

	var perms []string
	_ = json.Unmarshal([]byte(role.Permissions), &perms)
	userCount, _ := s.repo.CountUsersByRoleID(role.ID)

	return &dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: perms,
		IsSystem:    role.IsSystem,
		UsersCount:  int(userCount),
		CreatedAt:   role.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *Service) DeleteRole(id uint) error {
	role, err := s.repo.FindRoleByID(id)
	if err != nil {
		return fmt.Errorf("role tidak ditemukan")
	}

	if role.IsSystem {
		return fmt.Errorf("role sistem bawaan (%s) tidak dapat dihapus", role.Name)
	}

	userCount, _ := s.repo.CountUsersByRoleID(id)
	if userCount > 0 {
		return fmt.Errorf("role ini tidak dapat dihapus karena masih digunakan oleh %d user admin", userCount)
	}

	return s.repo.DeleteRole(id)
}

// ----------------------------------------------------
// USER MANAGEMENT SERVICE METHODS
// ----------------------------------------------------

func (s *Service) GetUsers() ([]dto.UserDetailResponse, error) {
	users, err := s.repo.FindUsers()
	if err != nil {
		return nil, err
	}

	var results []dto.UserDetailResponse
	for _, u := range users {
		roleName := u.Role
		var roleResp *dto.RoleResponse

		if u.RoleRel != nil {
			roleName = u.RoleRel.Name
			var perms []string
			_ = json.Unmarshal([]byte(u.RoleRel.Permissions), &perms)
			roleResp = &dto.RoleResponse{
				ID:          u.RoleRel.ID,
				Name:        u.RoleRel.Name,
				Description: u.RoleRel.Description,
				Permissions: perms,
				IsSystem:    u.RoleRel.IsSystem,
			}
		}

		results = append(results, dto.UserDetailResponse{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			RoleID:    u.RoleID,
			RoleName:  roleName,
			Role:      roleResp,
			Status:    u.Status,
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: u.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return results, nil
}

func (s *Service) CreateUser(req dto.UserCreateRequest) (*dto.UserDetailResponse, error) {
	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)

	if username == "" || email == "" || password == "" {
		return nil, fmt.Errorf("username, email, dan password wajib diisi")
	}

	if req.RoleID == 0 {
		return nil, fmt.Errorf("role wajib dipilih")
	}

	role, err := s.repo.FindRoleByID(req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("role yang dipilih tidak valid")
	}

	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("gagal mengenkripsi password")
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}

	roleID := role.ID
	user := domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		RoleID:       &roleID,
		Role:         role.Name,
		Status:       status,
	}

	if err := s.repo.CreateUser(&user); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, fmt.Errorf("username atau email sudah terdaftar")
		}
		return nil, err
	}

	var perms []string
	_ = json.Unmarshal([]byte(role.Permissions), &perms)

	return &dto.UserDetailResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RoleID:   user.RoleID,
		RoleName: role.Name,
		Role: &dto.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Permissions: perms,
			IsSystem:    role.IsSystem,
		},
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *Service) UpdateUser(id uint, req dto.UserUpdateRequest) (*dto.UserDetailResponse, error) {
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("user tidak ditemukan")
	}

	email := strings.TrimSpace(req.Email)
	if email != "" {
		user.Email = email
	}

	if req.RoleID > 0 {
		role, err := s.repo.FindRoleByID(req.RoleID)
		if err != nil {
			return nil, fmt.Errorf("role tidak valid")
		}
		user.RoleID = &role.ID
		user.Role = role.Name
		user.RoleRel = role
	}

	if req.Status != "" {
		user.Status = req.Status
	}

	if strings.TrimSpace(req.Password) != "" {
		hashed, err := security.HashPassword(strings.TrimSpace(req.Password))
		if err != nil {
			return nil, fmt.Errorf("gagal mengenkripsi password baru")
		}
		user.PasswordHash = hashed
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}

	var roleResp *dto.RoleResponse
	if user.RoleRel != nil {
		var perms []string
		_ = json.Unmarshal([]byte(user.RoleRel.Permissions), &perms)
		roleResp = &dto.RoleResponse{
			ID:          user.RoleRel.ID,
			Name:        user.RoleRel.Name,
			Description: user.RoleRel.Description,
			Permissions: perms,
			IsSystem:    user.RoleRel.IsSystem,
		}
	}

	return &dto.UserDetailResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		RoleID:    user.RoleID,
		RoleName:  user.Role,
		Role:      roleResp,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *Service) DeleteUser(id uint, currentUserID uint) error {
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		return fmt.Errorf("user tidak ditemukan")
	}

	if user.Username == "admin" {
		return fmt.Errorf("user admin utama tidak dapat dihapus")
	}

	if user.ID == currentUserID {
		return fmt.Errorf("anda tidak dapat menghapus akun Anda sendiri")
	}

	return s.repo.DeleteUser(id)
}







