package handler

import (
	"serve-swimming-be/internal/dto"
	"serve-swimming-be/internal/service"
	"serve-swimming-be/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type PublicHandler struct {
	svc *service.Service
}

func NewPublicHandler(svc *service.Service) *PublicHandler {
	return &PublicHandler{svc: svc}
}

func (h *PublicHandler) GetHomepageData(c *fiber.Ctx) error {
	banners, _ := h.svc.GetBanners()
	tournaments, _ := h.svc.GetTournaments()
	events, _ := h.svc.GetEvents()
	poolCfg, _ := h.svc.GetPoolConfig()
	heroCfg, _ := h.svc.GetHeroConfig()
	heroStats, _ := h.svc.GetHeroStats()
	siteCfg, _ := h.svc.GetSiteConfig()
	programs, _ := h.svc.GetTrainingPrograms()
	progSecCfg, _ := h.svc.GetProgramSectionConfig()
	coaches, _ := h.svc.GetCoaches()
	coachSecCfg, _ := h.svc.GetCoachSectionConfig()
	facilities, _ := h.svc.GetFacilities()
	facSecCfg, _ := h.svc.GetFacilitySectionConfig()
	achievements, _ := h.svc.GetAchievements()
	achSecCfg, _ := h.svc.GetAchievementSectionConfig()
	testimonials, _ := h.svc.GetTestimonials()
	testSecCfg, _ := h.svc.GetTestimonialSectionConfig()
	pageSections, _ := h.svc.GetPageSections("homepage")

	return response.Success(c, fiber.StatusOK, "Homepage data fetched successfully", fiber.Map{
		"banners":                    banners,
		"tournaments":                tournaments,
		"events":                     events,
		"pool_config":                poolCfg,
		"hero_config":                heroCfg,
		"hero_stats":                 heroStats,
		"site_config":                siteCfg,
		"training_programs":          programs,
		"program_section_config":     progSecCfg,
		"coaches":                    coaches,
		"coach_section_config":       coachSecCfg,
		"facilities":                  facilities,
		"facility_section_config":     facSecCfg,
		"achievements":                achievements,
		"achievement_section_config": achSecCfg,
		"testimonials":               testimonials,
		"testimonial_section_config": testSecCfg,
		"page_sections":              pageSections,
	})
}

func (h *PublicHandler) GetTournaments(c *fiber.Ctx) error {
	tournaments, err := h.svc.GetTournaments()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch tournaments", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Tournaments fetched successfully", tournaments)
}

func (h *PublicHandler) GetEvents(c *fiber.Ctx) error {
	events, err := h.svc.GetEvents()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch events", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Events fetched successfully", events)
}

func (h *PublicHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterParticipantRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if req.Name == "" || len(req.EventSelections) == 0 {
		return response.Error(c, fiber.StatusBadRequest, "Name and event selections are required", nil)
	}

	res, err := h.svc.RegisterParticipant(req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Registration failed", err.Error())
	}

	return response.Success(c, fiber.StatusCreated, "Registration created successfully", res)
}

func (h *PublicHandler) GetStartingList(c *fiber.Ctx) error {
	list, err := h.svc.GetStartingList()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch starting list", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Starting list fetched successfully", list)
}

func (h *PublicHandler) GetBukuAcara(c *fiber.Ctx) error {
	bukuAcara, err := h.svc.GetBukuAcara()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch buku acara", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Buku acara fetched successfully", bukuAcara)
}

func (h *PublicHandler) CheckRegistrationCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return response.Error(c, fiber.StatusBadRequest, "Registration code required", nil)
	}

	regs, err := h.svc.GetRegistrationByCode(code)
	if err != nil || len(regs) == 0 {
		return response.Error(c, fiber.StatusNotFound, "Registration code not found", nil)
	}

	return response.Success(c, fiber.StatusOK, "Registration details fetched successfully", regs)
}

func (h *PublicHandler) GetPageSections(c *fiber.Ctx) error {
	sections, err := h.svc.GetPageSections("homepage")
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch page sections", err.Error())
	}
	return response.Success(c, fiber.StatusOK, "Page sections fetched successfully", sections)
}
