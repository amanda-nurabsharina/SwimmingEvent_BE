package main

import (
	"fmt"
	"log"

	"serve-swimming-be/config"
	"serve-swimming-be/internal/handler"
	"serve-swimming-be/internal/middleware"
	"serve-swimming-be/internal/repository"
	"serve-swimming-be/internal/service"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	var db *gorm.DB
	var err error

	if cfg.DBDriver == "postgres" {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else {
		db, err = gorm.Open(sqlite.Open("swimming_event.db"), &gorm.Config{})
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Swimming Competition Database connected successfully")

	repo := repository.NewRepository(db)
	if err := repo.SeedInitialData(); err != nil {
		log.Fatalf("Failed to seed initial database data: %v", err)
	}

	svc := service.NewService(repo, cfg)
	publicHandler := handler.NewPublicHandler(svc)
	adminHandler := handler.NewAdminHandler(svc, cfg)
	uploadHandler := handler.NewUploadHandler()

	app := fiber.New(fiber.Config{
		AppName:       "Swimming Competition Event API",
		CaseSensitive: true,
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.SetupCORS(cfg.AllowedOrigins))
	app.Use(middleware.SetupRateLimiter(cfg.RateLimitMax))

	app.Static("/uploads", "./uploads")

	api := app.Group("/api/v1")

	// 1. PUBLIC ENDPOINTS
	pub := api.Group("/public", middleware.RequireAPIKey(cfg.APISecretKey))
	pub.Get("/homepage", publicHandler.GetHomepageData)
	pub.Get("/page-sections", publicHandler.GetPageSections)
	pub.Get("/tournaments", publicHandler.GetTournaments)
	pub.Get("/events", publicHandler.GetEvents)
	pub.Post("/register", publicHandler.Register)
	pub.Get("/starting-list", publicHandler.GetStartingList)
	pub.Get("/buku-acara", publicHandler.GetBukuAcara)
	pub.Get("/registration-status/:code", publicHandler.CheckRegistrationCode)

	// Public upload endpoint for payment proof
	pub.Post("/upload-proof", uploadHandler.UploadImage)

	// 2. AUTH ENDPOINTS
	auth := api.Group("/auth")
	auth.Post("/login", adminHandler.Login)

	// 3. ADMIN CMS ENDPOINTS (JWT Protected)
	admin := api.Group("/admin", middleware.RequireJWT(cfg.JWTSecret))
	admin.Post("/upload", uploadHandler.UploadImage)
	admin.Get("/registrations", adminHandler.GetRegistrations)
	admin.Put("/registrations/:id/verify", adminHandler.VerifyPayment)
	admin.Post("/buku-acara/generate", adminHandler.GenerateBukuAcara)
	admin.Put("/registrations/:id/result", adminHandler.RecordRaceResult)

	// Tournament Master CMS
	admin.Get("/tournaments", adminHandler.GetTournaments)
	admin.Post("/tournaments", adminHandler.SaveTournament)
	admin.Delete("/tournaments/:id", adminHandler.DeleteTournament)

	// Banner & Hero CMS
	admin.Get("/banners", adminHandler.GetBanners)
	admin.Post("/banners", adminHandler.SaveBanner)
	admin.Delete("/banners/:id", adminHandler.DeleteBanner)
	admin.Post("/banners/batch", adminHandler.BatchSaveBanners)
	admin.Get("/hero/config", adminHandler.GetHeroConfig)
	admin.Post("/hero/config", adminHandler.SaveHeroConfig)
	admin.Get("/hero/stats", adminHandler.GetHeroStats)
	admin.Post("/hero/stats", adminHandler.SaveHeroStats)

	// Site Config & Programs CMS
	admin.Get("/site-config", adminHandler.GetSiteConfig)
	admin.Post("/site-config", adminHandler.SaveSiteConfig)
	admin.Get("/programs/section-config", adminHandler.GetProgramSectionConfig)
	admin.Post("/programs/section-config", adminHandler.SaveProgramSectionConfig)
	admin.Get("/programs", adminHandler.GetTrainingPrograms)
	admin.Post("/programs", adminHandler.SaveTrainingProgram)
	admin.Delete("/programs/:id", adminHandler.DeleteTrainingProgram)

	// Coaches CMS
	admin.Get("/coaches/section-config", adminHandler.GetCoachSectionConfig)
	admin.Post("/coaches/section-config", adminHandler.SaveCoachSectionConfig)
	admin.Get("/coaches", adminHandler.GetCoaches)
	admin.Post("/coaches", adminHandler.SaveCoach)
	admin.Delete("/coaches/:id", adminHandler.DeleteCoach)

	// Facilities CMS
	admin.Get("/facilities/section-config", adminHandler.GetFacilitySectionConfig)
	admin.Post("/facilities/section-config", adminHandler.SaveFacilitySectionConfig)
	admin.Get("/facilities", adminHandler.GetFacilities)
	admin.Post("/facilities", adminHandler.SaveFacility)
	admin.Delete("/facilities/:id", adminHandler.DeleteFacility)

	// Achievements CMS
	admin.Get("/achievements/section-config", adminHandler.GetAchievementSectionConfig)
	admin.Post("/achievements/section-config", adminHandler.SaveAchievementSectionConfig)
	admin.Get("/achievements", adminHandler.GetAchievements)
	admin.Post("/achievements", adminHandler.SaveAchievement)
	admin.Delete("/achievements/:id", adminHandler.DeleteAchievement)

	// Testimonials CMS
	admin.Get("/testimonials/section-config", adminHandler.GetTestimonialSectionConfig)
	admin.Post("/testimonials/section-config", adminHandler.SaveTestimonialSectionConfig)
	admin.Get("/testimonials", adminHandler.GetTestimonials)
	admin.Post("/testimonials", adminHandler.SaveTestimonial)
	admin.Delete("/testimonials/:id", adminHandler.DeleteTestimonial)

	// Event Master CMS
	admin.Get("/events", adminHandler.GetEvents)
	admin.Post("/events", adminHandler.SaveEvent)
	admin.Delete("/events/:id", adminHandler.DeleteEvent)

	// Page Sections Sort & Layout CMS
	admin.Get("/page-sections", adminHandler.GetPageSections)
	admin.Post("/page-sections", adminHandler.SavePageSection)
	admin.Post("/page-sections/batch", adminHandler.BatchSavePageSections)
	admin.Post("/page-sections/reset", adminHandler.ResetPageSections)

	log.Printf("Swimming Event API starting on port :%s in %s mode...", cfg.Port, cfg.AppEnv)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
