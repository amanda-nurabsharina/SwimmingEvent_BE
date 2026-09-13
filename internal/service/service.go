package service

import (
	"crypto/rand"
	"fmt"
	"log"
	"sort"

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

	return &dto.LoginResponse{
		Token: token,
		User: dto.UserSummary{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
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
	// Generate Unique Registration Code matching Screenshot format REG-ASC-XXXXX
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
		if timeSeed == "" {
			timeSeed = "NT"
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

func (s *Service) GetStartingList() ([]dto.StartingItemDTO, error) {
	regs, err := s.repo.FindRegistrations()
	if err != nil {
		return nil, err
	}

	var items []dto.StartingItemDTO
	for idx, r := range regs {
		items = append(items, dto.StartingItemDTO{
			No:         idx + 1,
			Nama:       r.Participant.Name,
			Gender:     r.Participant.Gender,
			TimeSeed:   r.TimeSeed,
			NomorLomba: r.SwimmingEvent.EventName,
			Club:       r.Participant.Club,
			PIC:        r.Participant.PIC,
			Kontak:     r.Participant.Contact,
		})
	}
	return items, nil
}

// GenerateBukuAcara: Heat & Line tournament assignment algorithm
func (s *Service) GenerateBukuAcara(maxLanes int) error {
	if maxLanes <= 0 {
		maxLanes = 3
	}

	_ = s.repo.UpdatePoolConfigMaxLanes(maxLanes)
	events, err := s.repo.FindEvents()
	if err != nil {
		return err
	}

	regs, err := s.repo.FindRegistrations()
	if err != nil {
		return err
	}

	// Group registrations by SwimmingEventID
	eventGroup := make(map[uint][]domain.Registration)
	for _, r := range regs {
		eventGroup[r.SwimmingEventID] = append(eventGroup[r.SwimmingEventID], r)
	}

	for _, event := range events {
		eventRegs := eventGroup[event.ID]
		if len(eventRegs) == 0 {
			continue
		}

		// Sort by TimeSeed ascending (faster times first; 99.99.99 last)
		sort.Slice(eventRegs, func(i, j int) bool {
			return eventRegs[i].TimeSeed < eventRegs[j].TimeSeed
		})

		// Assign Heat and Line numbers
		for idx, reg := range eventRegs {
			heat := (idx / maxLanes) + 1
			line := (idx % maxLanes) + 1
			_ = s.repo.UpdateRegistrationHeatLine(reg.ID, heat, line)
		}
	}

	return nil
}

func (s *Service) GetBukuAcara() ([]dto.BukuAcaraEventGroupDTO, error) {
	events, err := s.repo.FindEvents()
	if err != nil {
		return nil, err
	}

	regs, err := s.repo.FindRegistrations()
	if err != nil {
		return nil, err
	}

	regMap := make(map[uint][]domain.Registration)
	for _, r := range regs {
		regMap[r.SwimmingEventID] = append(regMap[r.SwimmingEventID], r)
	}

	var result []dto.BukuAcaraEventGroupDTO
	for _, ev := range events {
		eRegs := regMap[ev.ID]
		if len(eRegs) == 0 {
			continue
		}

		// Sort by Heat number then Line number
		sort.Slice(eRegs, func(i, j int) bool {
			if eRegs[i].HeatNumber == eRegs[j].HeatNumber {
				return eRegs[i].LineNumber < eRegs[j].LineNumber
			}
			return eRegs[i].HeatNumber < eRegs[j].HeatNumber
		})

		var heatItems []dto.BukuAcaraHeatItemDTO
		for _, r := range eRegs {
			res := r.RaceResultTime
			if res == "" && r.Rank > 0 {
				res = fmt.Sprintf("Rank %d", r.Rank)
			}
			heatItems = append(heatItems, dto.BukuAcaraHeatItemDTO{
				Heat:     r.HeatNumber,
				Line:     r.LineNumber,
				Nama:     r.Participant.Name,
				Gender:   r.Participant.Gender,
				Club:     r.Participant.Club,
				TimeSeed: r.TimeSeed,
				Result:   res,
			})
		}

		result = append(result, dto.BukuAcaraEventGroupDTO{
			EventCode: ev.EventCode,
			EventName: ev.EventName,
			Gender:    ev.Gender,
			Heats:     heatItems,
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
	return s.repo.UpdateRegistrationStatus(id, status)
}

func (s *Service) RecordRaceResult(id uint, timeStr string, rank int) error {
	return s.repo.UpdateRaceResult(id, timeStr, rank)
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





