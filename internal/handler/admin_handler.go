package handler

import (
	"strconv"

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

	if req.PaymentStatus != "verified" && req.PaymentStatus != "rejected" {
		return response.Error(c, fiber.StatusBadRequest, "Status must be 'verified' or 'rejected'", nil)
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

	if err := h.svc.GenerateBukuAcara(req.MaxLanes); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to generate buku acara", err.Error())
	}

	bukuAcara, _ := h.svc.GetBukuAcara()
	return response.Success(c, fiber.StatusOK, "Buku acara generated successfully", bukuAcara)
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

	if err := h.svc.RecordRaceResult(uint(id), req.RaceResultTime, req.Rank); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to record race result", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "Race result recorded successfully", nil)
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





