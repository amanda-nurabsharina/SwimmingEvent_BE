package handler

import (
	"fmt"
	"strconv"
	"strings"

	"serve-swimming-be/config"
	"serve-swimming-be/internal/domain"
	"serve-swimming-be/internal/dto"
	"serve-swimming-be/internal/service"
	"serve-swimming-be/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	svc *service.Service
	cfg *config.Config
}

func NewAdminHandler(svc *service.Service, cfg *config.Config) *AdminHandler {
	return &AdminHandler{svc: svc, cfg: cfg}
}

func (h *AdminHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid login credentials format", err.Error())
	}

	res, err := h.svc.AuthenticateAdmin(req)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Login successful", res)
}

func (h *AdminHandler) GetRegistrations(c *fiber.Ctx) error {
	regs, err := h.svc.GetRegistrations()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch registrations", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Registrations fetched successfully", regs)
}

func (h *AdminHandler) VerifyPayment(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid registration ID", nil)
	}

	var req dto.VerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid verification request payload", err.Error())
	}

	if req.PaymentStatus != "verified" && req.PaymentStatus != "rejected" && req.PaymentStatus != "pending" {
		return response.Error(c, fiber.StatusBadRequest, "Status must be 'verified', 'pending', or 'rejected'", nil)
	}

	if err := h.svc.VerifyPayment(uint(id), req.PaymentStatus); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update payment status", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Payment status updated successfully", nil)
}

func (h *AdminHandler) GenerateBukuAcara(c *fiber.Ctx) error {
	var req dto.GenerateBukuAcaraRequest
	_ = c.BodyParser(&req)
	if req.MaxLanes <= 0 {
		req.MaxLanes = 3
	}

	if err := h.svc.GenerateBukuAcara(req.MaxLanes, req.TournamentID, req.Force); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	bukuAcara, _ := h.svc.GetBukuAcara(req.TournamentID, "preliminary")
	return response.Success(c, fiber.StatusOK, "Buku acara generated successfully", bukuAcara)
}

func (h *AdminHandler) GetBukuAcara(c *fiber.Ctx) error {
	tourneyIDStr := c.Query("tournament_id")
	tourneyID, _ := strconv.ParseUint(tourneyIDStr, 10, 64)
	round := c.Query("round")
	bukuAcara, err := h.svc.GetBukuAcara(uint(tourneyID), round)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch buku acara", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Buku acara fetched successfully", bukuAcara)
}

func (h *AdminHandler) GenerateFinalRound(c *fiber.Ctx) error {
	var req dto.GenerateFinalRoundRequest
	_ = c.BodyParser(&req)

	operator, _ := c.Locals("username").(string)
	if operator == "" {
		operator = "Admin Panitia"
	}
	ip := c.IP()

	if err := h.svc.GenerateFinalRound(req.TournamentID, req.MaxLanes, req.QualifyMode, operator, ip); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	bukuAcara, _ := h.svc.GetBukuAcara(req.TournamentID, "final")
	return response.Success(c, fiber.StatusOK, "Bagan babak final berhasil di-generate berdasarkan juara heat", bukuAcara)
}

func (h *AdminHandler) RecordRaceResult(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid registration ID", nil)
	}

	var req dto.RecordResultRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	timeResult := req.RaceResultTime
	if timeResult == "" {
		timeResult = req.FinalTime
	}
	if req.Status != "" && req.Status != "OK" {
		if timeResult == "" {
			timeResult = req.Status
		} else {
			timeResult = fmt.Sprintf("%s (%s)", timeResult, req.Status)
		}
	}

	operator, _ := c.Locals("username").(string)
	if operator == "" {
		operator = "Admin Panitia"
	}
	ip := c.IP()

	if err := h.svc.RecordRaceResult(uint(id), timeResult, req.Rank, req.Round, operator, ip); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Race result recorded successfully", nil)
}

func (h *AdminHandler) LockTournamentBukuAcara(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid tournament ID", nil)
	}

	var req dto.LockBukuAcaraRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	operator, _ := c.Locals("username").(string)
	if operator == "" {
		operator = "Admin Panitia"
	}
	ip := c.IP()

	if err := h.svc.SetTournamentBukuAcaraLock(uint(id), req.IsLocked, operator, ip); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update lock status", err.Error())
	}

	msg := "Buku acara berhasil dipatenkan/dikunci"
	if !req.IsLocked {
		msg = "Kunci buku acara berhasil dibuka"
	}
	return response.Success(c, fiber.StatusOK, msg, fiber.Map{
		"tournament_id": id,
		"is_locked":     req.IsLocked,
	})
}

func (h *AdminHandler) PublishTournamentBukuAcara(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid tournament ID", nil)
	}

	var req dto.PublishBukuAcaraRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	operator, _ := c.Locals("username").(string)
	if operator == "" {
		operator = "Admin Panitia"
	}
	ip := c.IP()

	if err := h.svc.SetTournamentBukuAcaraPublish(uint(id), req.IsPublished, operator, ip); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update publication status", err.Error())
	}

	msg := "Buku acara berhasil dipublikasikan ke halaman publik"
	if !req.IsPublished {
		msg = "Publikasi buku acara berhasil ditarik dari halaman publik"
	}
	return response.Success(c, fiber.StatusOK, msg, fiber.Map{
		"tournament_id": id,
		"is_published":  req.IsPublished,
	})
}

func (h *AdminHandler) SwapRegistrationHeatLine(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid registration ID", nil)
	}

	var req dto.SwapHeatLineRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	operator, _ := c.Locals("username").(string)
	if operator == "" {
		operator = "Admin Panitia"
	}
	ip := c.IP()

	regA, regB, err := h.svc.SwapRegistrationHeatLine(uint(id), req.TargetHeat, req.TargetLine, req.SwapIfOccupied, req.Round, operator, ip)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	isFinal := strings.ToLower(strings.TrimSpace(req.Round)) == "final"
	hA, lA := regA.HeatNumber, regA.LineNumber
	if isFinal {
		hA, lA = regA.FinalHeatNumber, regA.FinalLineNumber
	}

	msg := fmt.Sprintf("Perenang %s berhasil dipindahkan ke Heat %d Lintasan %d", regA.Participant.Name, hA, lA)
	if regB != nil {
		hB, lB := regB.HeatNumber, regB.LineNumber
		if isFinal {
			hB, lB = regB.FinalHeatNumber, regB.FinalLineNumber
		}
		msg = fmt.Sprintf("Posisi berhasil ditukar: %s di Heat %d Line %d, dan %s di Heat %d Line %d",
			regA.Participant.Name, hA, lA,
			regB.Participant.Name, hB, lB)
	}

	return response.Success(c, fiber.StatusOK, msg, fiber.Map{
		"swimmer":        regA,
		"swappedSwimmer": regB,
	})
}

func (h *AdminHandler) GetRaceResultLogs(c *fiber.Ctx) error {
	tourneyIDStr := c.Query("tournament_id")
	tourneyID, _ := strconv.ParseUint(tourneyIDStr, 10, 64)

	round := c.Query("round")
	action := c.Query("action")
	search := c.Query("search")

	limitStr := c.Query("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	offsetStr := c.Query("offset", "0")
	offset, _ := strconv.Atoi(offsetStr)

	logs, total, err := h.svc.GetRaceResultLogs(uint(tourneyID), round, action, search, limit, offset)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Gagal mengambil log catatan hasil lomba", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Log catatan hasil lomba berhasil diambil", fiber.Map{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *AdminHandler) GetRaceResultLogStats(c *fiber.Ctx) error {
	tourneyIDStr := c.Query("tournament_id")
	tourneyID, _ := strconv.ParseUint(tourneyIDStr, 10, 64)

	stats, err := h.svc.GetRaceResultLogStats(uint(tourneyID))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Gagal mengambil statistik log", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Statistik log berhasil diambil", stats)
}

func (h *AdminHandler) GetBanners(c *fiber.Ctx) error {
	banners, err := h.svc.GetBanners()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch banners", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Banners fetched successfully", banners)
}

func (h *AdminHandler) SaveBanner(c *fiber.Ctx) error {
	var banner domain.BannerSlide
	if err := c.BodyParser(&banner); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid banner payload", err.Error())
	}

	if err := h.svc.SaveBanner(&banner); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save banner", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Banner saved successfully", banner)
}

func (h *AdminHandler) DeleteBanner(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid banner ID", nil)
	}

	if err := h.svc.DeleteBanner(uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete banner", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Banner deleted successfully", nil)
}

func (h *AdminHandler) BatchSaveBanners(c *fiber.Ctx) error {
	var banners []domain.BannerSlide
	if err := c.BodyParser(&banners); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	for _, b := range banners {
		_ = h.svc.SaveBanner(&b)
	}

	allBanners, _ := h.svc.GetBanners()
	return response.Success(c, fiber.StatusOK, "Batch banners saved successfully", allBanners)
}

func (h *AdminHandler) GetHeroConfig(c *fiber.Ctx) error {
	cfg, err := h.svc.GetHeroConfig()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch hero config", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Hero config fetched successfully", cfg)
}

func (h *AdminHandler) SaveHeroConfig(c *fiber.Ctx) error {
	var cfg domain.HeroConfig
	if err := c.BodyParser(&cfg); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid hero config payload", err.Error())
	}

	if err := h.svc.SaveHeroConfig(&cfg); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save hero config", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Hero config saved successfully", cfg)
}

func (h *AdminHandler) GetHeroStats(c *fiber.Ctx) error {
	stats, err := h.svc.GetHeroStats()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch hero stats", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Hero stats fetched successfully", stats)
}

func (h *AdminHandler) SaveHeroStats(c *fiber.Ctx) error {
	var stats []domain.HeroStat
	if err := c.BodyParser(&stats); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid hero stats payload", err.Error())
	}

	if err := h.svc.SaveHeroStats(stats); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save hero stats", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Hero stats saved successfully", stats)
}

func (h *AdminHandler) GetSiteConfig(c *fiber.Ctx) error {
	cfg, err := h.svc.GetSiteConfig()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch site config", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Site config fetched successfully", cfg)
}

func (h *AdminHandler) SaveSiteConfig(c *fiber.Ctx) error {
	var cfg domain.SiteConfig
	if err := c.BodyParser(&cfg); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveSiteConfig(&cfg); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save site config", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Site config saved successfully", cfg)
}

func (h *AdminHandler) GetProgramSectionConfig(c *fiber.Ctx) error {
	cfg, err := h.svc.GetProgramSectionConfig()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch program section config", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Program section config fetched successfully", cfg)
}

func (h *AdminHandler) SaveProgramSectionConfig(c *fiber.Ctx) error {
	var cfg domain.ProgramSectionConfig
	if err := c.BodyParser(&cfg); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveProgramSectionConfig(&cfg); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save program section config", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Program section config saved successfully", cfg)
}

func (h *AdminHandler) GetTrainingPrograms(c *fiber.Ctx) error {
	progs, err := h.svc.GetTrainingPrograms()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch training programs", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Training programs fetched successfully", progs)
}

func (h *AdminHandler) SaveTrainingProgram(c *fiber.Ctx) error {
	var prog domain.TrainingProgram
	if err := c.BodyParser(&prog); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveTrainingProgram(&prog); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save training program", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Training program saved successfully", prog)
}

func (h *AdminHandler) DeleteTrainingProgram(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid program ID", nil)
	}

	if err := h.svc.DeleteTrainingProgram(uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete training program", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Training program deleted successfully", nil)
}

func (h *AdminHandler) GetCoachSectionConfig(c *fiber.Ctx) error {
	cfg, err := h.svc.GetCoachSectionConfig()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch coach section config", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Coach section config fetched successfully", cfg)
}

func (h *AdminHandler) SaveCoachSectionConfig(c *fiber.Ctx) error {
	var cfg domain.CoachSectionConfig
	if err := c.BodyParser(&cfg); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveCoachSectionConfig(&cfg); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save coach section config", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Coach section config saved successfully", cfg)
}

func (h *AdminHandler) GetCoaches(c *fiber.Ctx) error {
	coaches, err := h.svc.GetCoaches()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch coaches", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Coaches fetched successfully", coaches)
}

func (h *AdminHandler) SaveCoach(c *fiber.Ctx) error {
	var coach domain.Coach
	if err := c.BodyParser(&coach); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveCoach(&coach); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save coach", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Coach saved successfully", coach)
}

func (h *AdminHandler) DeleteCoach(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid coach ID", nil)
	}

	if err := h.svc.DeleteCoach(uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete coach", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Coach deleted successfully", nil)
}

func (h *AdminHandler) GetFacilitySectionConfig(c *fiber.Ctx) error {
	cfg, err := h.svc.GetFacilitySectionConfig()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch facility section config", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Facility section config fetched successfully", cfg)
}

func (h *AdminHandler) SaveFacilitySectionConfig(c *fiber.Ctx) error {
	var cfg domain.FacilitySectionConfig
	if err := c.BodyParser(&cfg); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveFacilitySectionConfig(&cfg); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save facility section config", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Facility section config saved successfully", cfg)
}

func (h *AdminHandler) GetFacilities(c *fiber.Ctx) error {
	facilities, err := h.svc.GetFacilities()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch facilities", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Facilities fetched successfully", facilities)
}

func (h *AdminHandler) SaveFacility(c *fiber.Ctx) error {
	var fac domain.Facility
	if err := c.BodyParser(&fac); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveFacility(&fac); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save facility", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Facility saved successfully", fac)
}

func (h *AdminHandler) DeleteFacility(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid facility ID", nil)
	}

	if err := h.svc.DeleteFacility(uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete facility", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Facility deleted successfully", nil)
}

// Achievements Handlers
func (h *AdminHandler) GetAchievementSectionConfig(c *fiber.Ctx) error {
	cfg, err := h.svc.GetAchievementSectionConfig()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch achievement section config", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Achievement section config fetched successfully", cfg)
}

func (h *AdminHandler) SaveAchievementSectionConfig(c *fiber.Ctx) error {
	var cfg domain.AchievementSectionConfig
	if err := c.BodyParser(&cfg); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveAchievementSectionConfig(&cfg); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save achievement section config", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Achievement section config saved successfully", cfg)
}

func (h *AdminHandler) GetAchievements(c *fiber.Ctx) error {
	achievements, err := h.svc.GetAchievements()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch achievements", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Achievements fetched successfully", achievements)
}

func (h *AdminHandler) SaveAchievement(c *fiber.Ctx) error {
	var ach domain.Achievement
	if err := c.BodyParser(&ach); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveAchievement(&ach); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save achievement", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Achievement saved successfully", ach)
}

func (h *AdminHandler) DeleteAchievement(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid achievement ID", nil)
	}

	if err := h.svc.DeleteAchievement(uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete achievement", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Achievement deleted successfully", nil)
}

// Testimonials Handlers
func (h *AdminHandler) GetTestimonialSectionConfig(c *fiber.Ctx) error {
	cfg, err := h.svc.GetTestimonialSectionConfig()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch testimonial section config", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Testimonial section config fetched successfully", cfg)
}

func (h *AdminHandler) SaveTestimonialSectionConfig(c *fiber.Ctx) error {
	var cfg domain.TestimonialSectionConfig
	if err := c.BodyParser(&cfg); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveTestimonialSectionConfig(&cfg); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save testimonial section config", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Testimonial section config saved successfully", cfg)
}

func (h *AdminHandler) GetTestimonials(c *fiber.Ctx) error {
	testimonials, err := h.svc.GetTestimonials()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch testimonials", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Testimonials fetched successfully", testimonials)
}

func (h *AdminHandler) SaveTestimonial(c *fiber.Ctx) error {
	var test domain.Testimonial
	if err := c.BodyParser(&test); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveTestimonial(&test); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save testimonial", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Testimonial saved successfully", test)
}

func (h *AdminHandler) DeleteTestimonial(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid testimonial ID", nil)
	}

	if err := h.svc.DeleteTestimonial(uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete testimonial", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Testimonial deleted successfully", nil)
}

// Event Master Handlers
func (h *AdminHandler) GetEvents(c *fiber.Ctx) error {
	events, err := h.svc.GetEvents()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch events", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Events fetched successfully", events)
}

func (h *AdminHandler) SaveEvent(c *fiber.Ctx) error {
	var evt domain.SwimmingEvent
	if err := c.BodyParser(&evt); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveEvent(&evt); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save event", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Event saved successfully", evt)
}

func (h *AdminHandler) DeleteEvent(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid event ID", nil)
	}

	if err := h.svc.DeleteEvent(uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete event", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Event deleted successfully", nil)
}

// Tournament Master Handlers
func (h *AdminHandler) GetTournaments(c *fiber.Ctx) error {
	tournaments, err := h.svc.GetTournaments()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch tournaments", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Tournaments fetched successfully", tournaments)
}

func (h *AdminHandler) SaveTournament(c *fiber.Ctx) error {
	var t domain.Tournament
	if err := c.BodyParser(&t); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SaveTournament(&t); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save tournament", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Tournament saved successfully", t)
}

func (h *AdminHandler) DeleteTournament(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid tournament ID", nil)
	}

	if err := h.svc.DeleteTournament(uint(id)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete tournament", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Tournament deleted successfully", nil)
}

// Page Sections CMS Handlers
func (h *AdminHandler) GetPageSections(c *fiber.Ctx) error {
	slug := c.Query("page_slug", "homepage")
	sections, err := h.svc.GetPageSections(slug)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch page sections", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Page sections fetched successfully", sections)
}

func (h *AdminHandler) SavePageSection(c *fiber.Ctx) error {
	var sec domain.PageSection
	if err := c.BodyParser(&sec); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payload", err.Error())
	}

	if err := h.svc.SavePageSection(&sec); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save page section", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Page section saved successfully", sec)
}

func (h *AdminHandler) BatchSavePageSections(c *fiber.Ctx) error {
	var req struct {
		Sections []domain.PageSection `json:"sections"`
	}

	// Support either array directly or { sections: [...] }
	if err := c.BodyParser(&req); err != nil || len(req.Sections) == 0 {
		var directArr []domain.PageSection
		if err2 := c.BodyParser(&directArr); err2 == nil && len(directArr) > 0 {
			req.Sections = directArr
		}
	}

	if len(req.Sections) == 0 {
		return response.Error(c, fiber.StatusBadRequest, "No sections provided", nil)
	}

	if err := h.svc.BatchSavePageSections(req.Sections); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to batch save page sections", err.Error())
	}

	allSections, _ := h.svc.GetPageSections("homepage")
	return response.Success(c, fiber.StatusOK, "Batch page sections saved successfully", allSections)
}

func (h *AdminHandler) ResetPageSections(c *fiber.Ctx) error {
	slug := c.Query("page_slug", "homepage")
	if err := h.svc.ResetPageSections(slug); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to reset page sections", err.Error())
	}
	allSections, _ := h.svc.GetPageSections(slug)
	return response.Success(c, fiber.StatusOK, "Page sections reset to default successfully", allSections)
}

// ----------------------------------------------------
// ROLE & PERMISSION HANDLERS
// ----------------------------------------------------

func (h *AdminHandler) GetRoles(c *fiber.Ctx) error {
	roles, err := h.svc.GetRoles()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch roles", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Roles fetched successfully", roles)
}

func (h *AdminHandler) CreateRole(c *fiber.Ctx) error {
	var req dto.RoleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid role payload", err.Error())
	}

	role, err := h.svc.CreateRole(req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "Role created successfully", role)
}

func (h *AdminHandler) UpdateRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid role ID", nil)
	}

	var req dto.RoleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid role payload", err.Error())
	}

	role, err := h.svc.UpdateRole(uint(id), req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Role updated successfully", role)
}

func (h *AdminHandler) DeleteRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid role ID", nil)
	}

	if err := h.svc.DeleteRole(uint(id)); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Role deleted successfully", nil)
}

// ----------------------------------------------------
// USER MANAGEMENT HANDLERS
// ----------------------------------------------------

func (h *AdminHandler) GetUsers(c *fiber.Ctx) error {
	users, err := h.svc.GetUsers()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch users", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Users fetched successfully", users)
}

func (h *AdminHandler) CreateUser(c *fiber.Ctx) error {
	var req dto.UserCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user payload", err.Error())
	}

	user, err := h.svc.CreateUser(req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "User created successfully", user)
}

func (h *AdminHandler) UpdateUser(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", nil)
	}

	var req dto.UserUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user payload", err.Error())
	}

	user, err := h.svc.UpdateUser(uint(id), req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "User updated successfully", user)
}

func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", nil)
	}

	currentUserID, _ := c.Locals("user_id").(uint)

	if err := h.svc.DeleteUser(uint(id), currentUserID); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "User deleted successfully", nil)
}





