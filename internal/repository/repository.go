package repository

import (
	"fmt"
	"log"
	"strings"

	"serve-swimming-be/internal/domain"
	"serve-swimming-be/pkg/security"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AutoMigrate() error {
	// Drop old unique index on registration_code if present, and recreate as regular index to support multiple events per swimmer order
	_ = r.db.Exec("DROP INDEX IF EXISTS idx_registrations_registration_code").Error
	_ = r.db.Exec("CREATE INDEX IF NOT EXISTS idx_registrations_registration_code ON registrations(registration_code)").Error

	return r.db.AutoMigrate(
		&domain.Role{},
		&domain.User{},
		&domain.MasterMenu{},
		&domain.BannerSlide{},
		&domain.PageSection{},
		&domain.Tournament{},
		&domain.SwimmingEvent{},
		&domain.Participant{},
		&domain.Registration{},
		&domain.PoolConfig{},
		&domain.HeroConfig{},
		&domain.HeroStat{},
		&domain.AuditLog{},
		&domain.RaceResultLog{},
		&domain.SiteConfig{},
		&domain.TrainingProgram{},
		&domain.ProgramSectionConfig{},
		&domain.Coach{},
		&domain.CoachSectionConfig{},
		&domain.Facility{},
		&domain.FacilitySectionConfig{},
		&domain.Achievement{},
		&domain.AchievementSectionConfig{},
		&domain.Testimonial{},
		&domain.TestimonialSectionConfig{},
	)
}

func (r *Repository) SeedInitialData() error {
	if err := r.AutoMigrate(); err != nil {
		return err
	}

	// 0. Default Super Admin Role
	var superAdminRole domain.Role
	if err := r.db.Where("name = ?", "Super Admin").First(&superAdminRole).Error; err != nil {
		superAdminRole = domain.Role{
			Name:        "Super Admin",
			Description: "Role sistem dengan hak akses penuh ke seluruh modul & pengaturan",
			Permissions: `["*"]`,
			IsSystem:    true,
		}
		if err := r.db.Create(&superAdminRole).Error; err != nil {
			log.Printf("Error seeding Super Admin role: %v", err)
		}
	} else if !superAdminRole.IsSystem {
		r.db.Model(&superAdminRole).Update("is_system", true)
	}

	// 1. Default Admin User
	var userCount int64
	r.db.Model(&domain.User{}).Count(&userCount)
	if userCount == 0 {
		hashedPassword, _ := security.HashPassword("admin123")
		roleID := superAdminRole.ID
		admin := domain.User{
			Username:     "admin",
			Email:        "admin@akuatik-tangerang.id",
			PasswordHash: hashedPassword,
			RoleID:       &roleID,
			Role:         "Super Admin",
			Status:       "active",
		}
		if err := r.db.Create(&admin).Error; err != nil {
			log.Printf("Error seeding admin: %v", err)
		}
	} else {
		var adminUser domain.User
		if err := r.db.Where("username = ?", "admin").First(&adminUser).Error; err == nil {
			if adminUser.RoleID == nil || *adminUser.RoleID == 0 {
				r.db.Model(&adminUser).Updates(map[string]interface{}{
					"role_id": superAdminRole.ID,
					"role":    "Super Admin",
				})
			}
		}
	}

	// 2. Default Pool Config
	var poolCount int64
	r.db.Model(&domain.PoolConfig{}).Count(&poolCount)
	if poolCount == 0 {
		pool := domain.PoolConfig{
			CompetitionName: "TIME TRIAL 2025 AKUATIK INDONESIA KOTA TANGERANG",
			Location:        "Kolam Renang Gelanggang Kota Tangerang",
			MaxLanes:        3, // Match sample PDF format
		}
		r.db.Create(&pool)
	}

	// 2a. Default Tournaments
	var tourneyCount int64
	r.db.Model(&domain.Tournament{}).Count(&tourneyCount)
	if tourneyCount == 0 {
		tourneys := []domain.Tournament{
			{
				ID:                    1,
				Name:                  "KEJUARAAN TIME TRIAL AKUATIK INDONESIA 2026",
				Description:           "Kejuaraan time trial renang resmi berstandar PRSI / FINA untuk menjaring bibit atlet nasional.",
				Location:              "Kolam Renang Gelanggang Kota Tangerang",
				RegistrationStartDate: "2026-01-01",
				RegistrationEndDate:   "2026-12-31",
				EventStartDate:        "2026-10-01",
				EventEndDate:          "2026-10-03",
				IsActive:              true,
			},
			{
				ID:                    2,
				Name:                  "TURNAMEN RENANG HUT RI 17 2027",
				Description:           "Turnamen spesial memperingati Kemerdekaan RI ke-82 dengan total hadiah puluhan juta rupiah.",
				Location:              "Kolam Renang Senayan Jakarta",
				RegistrationStartDate: "2027-07-01",
				RegistrationEndDate:   "2027-08-10",
				EventStartDate:        "2027-08-17",
				EventEndDate:          "2027-08-18",
				IsActive:              true,
			},
			{
				ID:                    3,
				Name:                  "KEJUARAAN RENANG PELAJAR TERBUKA 2025",
				Description:           "Kejuaraan tingkat pelajar se-Jabodetabek (Pendaftaran Telah Ditutup).",
				Location:              "Kolam Renang Cikini Jakarta",
				RegistrationStartDate: "2025-01-01",
				RegistrationEndDate:   "2025-02-01",
				EventStartDate:        "2025-02-15",
				EventEndDate:          "2025-02-16",
				IsActive:              true,
			},
		}
		for _, t := range tourneys {
			r.db.Create(&t)
		}
	}

	// 2b. Default Hero Config (Kotak 1)
	var heroCfgCount int64
	r.db.Model(&domain.HeroConfig{}).Count(&heroCfgCount)
	if heroCfgCount == 0 {
		heroCfg := domain.HeroConfig{
			BadgeText:        "Official PRSI Certified Swim Academy & Event Partner",
			TitlePrefix:      "Akademi Renang Profesional &",
			TitleHighlight:   "Platform Kejuaraan Terintegrasi",
			Subtitle:         "Kurikulum renang berstandar internasional dari usia balita hingga atlet nasional, didukung sistem manajemen kompetisi renang digital modern.",
			Feature1:         "Pelatih Berlisensi Resmi PRSI / FINA",
			Feature2:         "Kolam Standar Olimpiade & Air Hangat",
			Feature3:         "Registrasi & Bagan Lomba Online (Bebas Login)",
			Feature4:         "Live Scoreboard & E-Sertifikat Instan",
			CTAPrimaryText:   "Daftar Kejuaraan",
			CTAPrimaryURL:    "#register",
			CTASecondaryText: "Lihat Bagan (Heat Sheet)",
			CTASecondaryURL:  "/buku-acara",
			CTATertiaryText:  "Login Portal Tim",
			CTATertiaryURL:   "http://localhost:3001",
			NoteText:         "Tamu & Penonton: Bebas melihat bagan lomba, jadwal, & pendaftaran langsung tanpa login.",
			TrustText1:       "Terdaftar & Diakui PRSI",
			TrustText2:       "35+ Klub Renang Bergabung",
		}
		r.db.Create(&heroCfg)
	}

	// 2c. Default Hero Stats (Kotak 3)
	var heroStatCount int64
	r.db.Model(&domain.HeroStat{}).Count(&heroStatCount)
	if heroStatCount == 0 {
		stats := []domain.HeroStat{
			{Value: "1,850+", Label: "Murid Aktif", SortOrder: 1},
			{Value: "45+", Label: "Pelatih Berlisensi", SortOrder: 2},
			{Value: "320+", Label: "Medali Kejuaraan", SortOrder: 3},
			{Value: "28+", Label: "Kompetisi Terselenggara", SortOrder: 4},
		}
		for _, s := range stats {
			r.db.Create(&s)
		}
	}

	// 3. Default Banners
	var bannerCount int64
	r.db.Model(&domain.BannerSlide{}).Count(&bannerCount)
	if bannerCount == 0 {
		banners := []domain.BannerSlide{
			{
				Title:            "Kejuaraan Renang Time Trial 2025",
				BadgeText:        "AKUATIK INDONESIA KOTA TANGERANG",
				Description:      "Ajang kompetisi resmi perenang bakat tingkat regional dan nasional. Daftarkan diri dan klub kamu sekarang!",
				ImageURL:         "https://images.unsplash.com/photo-1530549387789-4c1017266635?q=80&w=1600&auto=format&fit=crop",
				CTAPrimaryText:   "Daftar Lomba",
				CTAPrimaryURL:    "#register",
				CTASecondaryText: "Lihat Buku Acara",
				CTASecondaryURL:  "/buku-acara",
				SortOrder:        1,
				Status:           "published",
			},
			{
				Title:            "Fasilitas Kolam Standard Olahraga",
				BadgeText:        "50 METERS OLYMPIC POOL",
				Description:      "Sistem pencatatan waktu presisi tinggi dengan standar federasi Akuatik Indonesia.",
				ImageURL:         "https://images.unsplash.com/photo-1519315901367-f34ff9154487?q=80&w=1600&auto=format&fit=crop",
				CTAPrimaryText:   "Cek Starting List",
				CTAPrimaryURL:    "/starting-list",
				CTASecondaryText: "Lokasi Kolam",
				CTASecondaryURL:  "#venue",
				SortOrder:        2,
				Status:           "published",
			},
		}
		for _, b := range banners {
			r.db.Create(&b)
		}
	}

	// 4. Default Site Config
	var siteCfgCount int64
	r.db.Model(&domain.SiteConfig{}).Count(&siteCfgCount)
	if siteCfgCount == 0 {
		siteCfg := domain.SiteConfig{
			AppName:           "AKUATIK TANGERANG",
			AppTagline:        "TIME TRIAL CHAMPIONSHIP 2025",
			LogoURL:           "",
			WANumber:          "6281234567890",
			DefaultWATemplate: "Halo Admin Akuatik Tangerang, saya mau menanyakan informasi seputar kejuaraan dan program pelatihan renang.",
			FooterDescription: "Membangun Karakter, Mengasah Teknik, Mencetak Juara Renang Masa Depan",
			FooterText:        "© 2025 Akuatik Tangerang. All rights reserved. Platform Resmi Kejuaraan & Pembinaan Atlet Renang.",
			Email:             "info@akuatik-tangerang.id",
			Address:           "Aquatic Center Complex, Jl. Pintu Satu Senayan No. 8, Jakarta Pusat",
			YoutubeURL:        "https://youtube.com",
			InstagramURL:      "https://instagram.com",
			TiktokURL:         "https://tiktok.com",
			FacebookURL:        "https://facebook.com",
		}
		r.db.Create(&siteCfg)
	}

	// 5. Default Program Section Config
	var progSecCount int64
	r.db.Model(&domain.ProgramSectionConfig{}).Count(&progSecCount)
	if progSecCount == 0 {
		progSec := domain.ProgramSectionConfig{
			BadgeText:            "KURIKULUM BERJENJANG & TERSTRUKTUR",
			Title:                "Program Pelatihan Renang Unggulan",
			Subtitle:             "Dirancang secara ilmiah untuk membentuk fondasi renang yang kuat, aman, dan berorientasi prestasi untuk segala rentang usia.",
			AssessmentTitle:      "Bingung Memilih Kelas yang Tepat untuk Anak Anda?",
			AssessmentSubtitle:   "Ikuti sesi Free Water Assessment (Uji Kemampuan Air) selama 20 menit bersama Head Coach kami untuk menentukan level penempatan yang optimal.",
			AssessmentButtonText: "Jadwalkan Free Assessment ->",
			AssessmentWATemplate: "Halo Admin, saya ingin mendaftar sesi Free Water Assessment (Uji Kemampuan Air) 20 menit untuk anak saya.",
		}
		r.db.Create(&progSec)
	}

	// 6. Default Training Programs
	var progCount int64
	r.db.Model(&domain.TrainingProgram{}).Count(&progCount)
	if progCount == 0 {
		progs := []domain.TrainingProgram{
			{
				Category:     "Anak & Balita",
				Title:        "Baby & Toddler Aquatic",
				Subtitle:     "USIA DINI (6 BULAN - 3 TAHUN)",
				AgeBadge:     "6 Bln - 3 Thn",
				PopularBadge: "",
				Description:  "Program pengenalan air, refleks motorik, dan bonding orang tua bersama instruktur bersertifikasi infant swim.",
				Features:     "1 Coach : 2 Anak + Orang Tua\nKolam Air Hangat Suhu Terkontrol\nMainan Sensorik Air Edukatif\nSertifikat Milestone Kemampuan",
				Price:        "Rp 850.000 / bln",
				ImageURL:     "https://images.unsplash.com/photo-1530549387789-4c1017266635?w=800&auto=format&fit=crop&q=80",
				WATemplate:   "Halo Admin Akuatik Tangerang, saya ingin mendaftar program *Baby & Toddler Aquatic (6 Bln - 3 Thn)*. Mohon info jadwal kelas.",
				SortOrder:    1,
				IsActive:     true,
			},
			{
				Category:     "Anak & Balita",
				Title:        "Kids Learn to Swim (Reguler)",
				Subtitle:     "ANAK-ANAK (4 - 12 TAHUN)",
				AgeBadge:     "4 - 12 Thn",
				PopularBadge: "PALING POPULER",
				Description:  "Pembelajaran 4 gaya renang (Bebas, Dada, Punggung, Kupu) dengan penekanan pada keselamatan air (water safety) dan efisiensi tarikan.",
				Features:     "Maksimal 4 anak per kelas\nEvaluasi kenaikan level tiap 3 bulan\nPenguasaan 4 gaya kompetisi\nAkses Kolam Renang Standar",
				Price:        "Rp 750.000 / bln",
				ImageURL:     "https://images.unsplash.com/photo-1519315901367-f34ff9154487?w=800&auto=format&fit=crop&q=80",
				WATemplate:   "Halo Admin Akuatik Tangerang, saya ingin mendaftar program *Kids Learn to Swim (Reguler)*. Mohon informasi ketersediaan slot kelas.",
				SortOrder:    2,
				IsActive:     true,
			},
			{
				Category:     "Prestasi & Squad",
				Title:        "Prestasi & Squad Atlet (Club)",
				Subtitle:     "PEMBINAAN ATLET KOMPETISI",
				AgeBadge:     "KU 5 s/d KU 1",
				PopularBadge: "",
				Description:  "Latihan intensif pembinaan atlet kompetisi tingkat daerah, Kejurda, Kejurnas, dan Popda dengan analisis video stroke mekanik.",
				Features:     "Latihan 5-6x seminggu + Dryland\nAnalisa biomekanik berkala\nTarget ranking nasional & limit waktu\nPrioritas kejuaraan & training camp",
				Price:        "Rp 1.200.000 / bln",
				ImageURL:     "https://images.unsplash.com/photo-1560090995-01c3288b99d3?w=800&auto=format&fit=crop&q=80",
				WATemplate:   "Halo Admin Akuatik Tangerang, saya berminat mendaftar program *Prestasi & Squad Atlet (Club)*. Mohon info seleksi dan jadwal latihan.",
				SortOrder:    3,
				IsActive:     true,
			},
			{
				Category:     "Privat & Dewasa",
				Title:        "Private & Adult Master Swim",
				Subtitle:     "PRIVAT EKSKLUSIF / DEWASA",
				AgeBadge:     "Semua Usia",
				PopularBadge: "",
				Description:  "Bimbingan 1-on-1 fleksibel untuk perbaikan teknik, fitness, persiapan tes kedinasan (Akpol/Akmil), dan hydrotherapy kebugaran.",
				Features:     "Jadwal fleksibel (pagi/malam)\n1 Coach : 1 Murid khusus\nTarget disesuaikan kebutuhan murid\nVideo review setiap sesi",
				Price:        "Rp 1.500.000 / 4 sesi",
				ImageURL:     "https://images.unsplash.com/photo-1519315901367-f34ff9154487?w=800&auto=format&fit=crop&q=80",
				WATemplate:   "Halo Admin Akuatik Tangerang, saya ingin mendaftar kelas *Private & Adult Master Swim*. Mohon informasi jadwal pelatih privat.",
				SortOrder:    4,
				IsActive:     true,
			},
		}
		for _, p := range progs {
			r.db.Create(&p)
		}
	}

	// 7. Default Coach Section Config
	var coachSecCount int64
	r.db.Model(&domain.CoachSectionConfig{}).Count(&coachSecCount)
	if coachSecCount == 0 {
		coachSec := domain.CoachSectionConfig{
			BadgeText: "TIM KEPELATIHAN PROFESIONAL",
			Title:     "Didampingi Pelatih Bersertifikasi Nasional & FINA",
			Subtitle:  "Setiap pelatih di MASC Swim memiliki lisensi resmi, pengalaman kepelatihan bertahun-tahun, serta dedikasi tinggi dalam membimbing setiap perenang secara terukur dan aman.",
		}
		r.db.Create(&coachSec)
	}

	// 8. Default Coaches (Matching User Screenshot)
	var coachCount int64
	r.db.Model(&domain.Coach{}).Count(&coachCount)
	if coachCount == 0 {
		coaches := []domain.Coach{
			{
				Name:             "Coach Raditya Pratama, S.Or.",
				RoleTitle:        "Head Coach & Performance Director",
				LicenseBadge:     "FINA Level 3 & PRSI Level A",
				Experience:       "Pengalaman: 14+ Tahun Pengalaman (Mantan Atlet Pelatnas)",
				Description:      "Berpengalaman membina atlet Kejurnas dan PON dengan spesialisasi gaya dada dan gaya ganti.",
				VerificationText: "Verified Coach PB PRSI",
				PhotoURL:         "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=800&auto=format&fit=crop&q=80",
				SortOrder:        1,
				IsActive:         true,
			},
			{
				Name:             "Coach Nadia Kirana, M.Pd.",
				RoleTitle:        "Senior Infant & Kids Aquatic Specialist",
				LicenseBadge:     "Austswim Teacher of Swimming & Water Safety",
				Experience:       "Pengalaman: 9+ Tahun Pengalaman",
				Description:      "Spesialis metode pengajaran renang ramah anak, penanganan trauma air, dan perkembangan motorik.",
				VerificationText: "Verified Coach PB PRSI",
				PhotoURL:         "https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?w=800&auto=format&fit=crop&q=80",
				SortOrder:        2,
				IsActive:         true,
			},
			{
				Name:             "Coach Farhan Maulana",
				RoleTitle:        "Sprint & Relay Technical Coach",
				LicenseBadge:     "PRSI Level B & ASCA Level 2",
				Experience:       "Pengalaman: 8+ Tahun Pengalaman",
				Description:      "Fokus pada teknik eksplosif start, flip turn, dan biomekanik renang gaya bebas & kupu-kupu.",
				VerificationText: "Verified Coach PB PRSI",
				PhotoURL:         "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=800&auto=format&fit=crop&q=80",
				SortOrder:        3,
				IsActive:         true,
			},
			{
				Name:             "Coach Sarah Octavia, S.Ked.",
				RoleTitle:        "Aquatic Conditioning & Rehab Trainer",
				LicenseBadge:     "Certified Aqua Fitness & Hydrotherapy",
				Experience:       "Pengalaman: 6+ Tahun Pengalaman",
				Description:      "Membimbing program rehabilitasi cedera olahraga dan penguatan stamina perenang master.",
				VerificationText: "Verified Coach PB PRSI",
				PhotoURL:         "https://images.unsplash.com/photo-1580489944761-15a19d654956?w=800&auto=format&fit=crop&q=80",
				SortOrder:        4,
				IsActive:         true,
			},
		}
		for _, c := range coaches {
			r.db.Create(&c)
		}
	}

	// 9. Default Facility Section Config
	var facSecCount int64
	r.db.Model(&domain.FacilitySectionConfig{}).Count(&facSecCount)
	if facSecCount == 0 {
		facSec := domain.FacilitySectionConfig{
			BadgeText: "FASILITAS & STANDAR KOLAM",
			Title:     "Infrastruktur Kolam Renang Standar Internasional",
			Subtitle:  "Lingkungan latihan yang higienis, aman, dan dirancang khusus untuk kenyamanan murid dari usia balita hingga atlet profesional.",
		}
		r.db.Create(&facSec)
	}

	// 10. Default Facilities (Matching User Screenshot)
	var facCount int64
	r.db.Model(&domain.Facility{}).Count(&facCount)
	if facCount == 0 {
		facilities := []domain.Facility{
			{
				Title:       "Olympic Competition Pool (50m)",
				TagText:     "50m x 25m",
				Description: "Kolam standar FINA 50 meter dengan 8 lintasan, depth 2.0m, timing sensor pads ready, dan overflow gutters.",
				SpecsText:   "50m x 25m | 8 Lintasan | Kedalaman 2.0m",
				ImageURL:    "https://images.unsplash.com/photo-1576013551627-0cc20b96c2a7?w=800&auto=format&fit=crop&q=80",
				SortOrder:   1,
				IsActive:    true,
			},
			{
				Title:       "Semi-Indoor Heated Training Pool (25m)",
				TagText:     "25m x 12m",
				Description: "Kolam latihan semi-indoor dengan atap kanopi pelindung UV, sistem filter garam (saltwater) tanpa klorin menyengat.",
				SpecsText:   "25m x 12m | Suhu 29°C - 31°C | Ramah Kulit Sensitif",
				ImageURL:    "https://images.unsplash.com/photo-1575429198097-0414ec08e8cd?w=800&auto=format&fit=crop&q=80",
				SortOrder:   2,
				IsActive:    true,
			},
			{
				Title:       "Dryland Conditioning & Gym Center",
				TagText:     "Cardio & Strength Equipment",
				Description: "Area latihan darat khusus perenang yang dilengkapi pull benches, resistance bands, dan plyometric stations.",
				SpecsText:   "Cardio & Strength Equipment | Yoga Mats | Core Trainer",
				ImageURL:    "https://images.unsplash.com/photo-1534438327276-14e5300c3a48?w=800&auto=format&fit=crop&q=80",
				SortOrder:   3,
				IsActive:    true,
			},
		}
		for _, f := range facilities {
			r.db.Create(&f)
		}
	}

	// 11. Default Achievement Section Config & Items
	var achSecCount int64
	r.db.Model(&domain.AchievementSectionConfig{}).Count(&achSecCount)
	if achSecCount == 0 {
		achSec := domain.AchievementSectionConfig{
			BadgeText: "REKAM JEJAK PRESTASI",
			Title:     "Pencapaian Medali & Kejuaraan Resmi",
			Subtitle:  "Komitmen kami dalam pembinaan atlet terbukti dengan raihan medali di berbagai kejuaraan renang tingkat daerah maupun nasional.",
		}
		r.db.Create(&achSec)
	}

	var achCount int64
	r.db.Model(&domain.Achievement{}).Count(&achCount)
	if achCount == 0 {
		achievements := []domain.Achievement{
			{
				Title:      "Juara Umum 1 Kejurnas Renang Pelajar",
				Year:       "2026",
				MedalType:  "gold",
				EventName:  "Kejurnas Antar Perkumpulan Renang 2026",
				WinnerName: "Tim Prestasi MASC Swim Club",
				SortOrder:  1,
				IsActive:   true,
			},
			{
				Title:      "Emas 50m & 100m Gaya Bebas KU 3 Putra",
				Year:       "2025",
				MedalType:  "gold",
				EventName:  "Jakarta Open Swimming Championship",
				WinnerName: "Rayhan Al Fatih",
				SortOrder:  2,
				IsActive:   true,
			},
			{
				Title:      "Perak 4x50m Estafet Gaya Ganti Putri",
				Year:       "2025",
				MedalType:  "silver",
				EventName:  "Piala Gubernur Aquatic Cup",
				WinnerName: "Tim Estafet Putri MASC",
				SortOrder:  3,
				IsActive:   true,
			},
			{
				Title:      "Best Swimmer KU 4 Putri Nasional",
				Year:       "2025",
				MedalType:  "gold",
				EventName:  "Bandung Sprint Fest 2025",
				WinnerName: "Aisyah Putri Azzahra",
				SortOrder:  4,
				IsActive:   true,
			},
		}
		for _, a := range achievements {
			r.db.Create(&a)
		}
	}

	// 12. Default Testimonial Section Config & Items
	var testSecCount int64
	r.db.Model(&domain.TestimonialSectionConfig{}).Count(&testSecCount)
	if testSecCount == 0 {
		testSec := domain.TestimonialSectionConfig{
			BadgeText: "KEPUASAN ORANG TUA & OFFICIAL KLUB",
			Title:     "Apa Kata Mereka Tentang MASC Swim?",
			Subtitle:  "Pengalaman nyata orang tua murid dan pengurus klub yang merasakan manfaat kurikulum renang serta kemudahan sistem turnamen digital.",
		}
		r.db.Create(&testSec)
	}

	var testCount int64
	r.db.Model(&domain.Testimonial{}).Count(&testCount)
	if testCount == 0 {
		testimonials := []domain.Testimonial{
			{
				Name:      "Bapak Hendra Wijaya",
				RoleTitle: "Orang Tua Atlet (Rayhan, KU 3)",
				Content:   "Sistem pendaftaran kompetisi di MASC Swim sangat cepat dan transparan! Dari registrasi nomor lomba, pembayaran QRIS otomatis langsung lunas, hingga bagan dan live time di venue terintegrasi mulus.",
				Rating:    5,
				AvatarURL: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=400&auto=format&fit=crop&q=80",
				SortOrder: 1,
				IsActive:  true,
			},
			{
				Name:      "Ibu Ratna Dewi",
				RoleTitle: "Orang Tua Murid Kids Class",
				Content:   "Pelatihnya sangat telaten dan sabar menghadapi anak yang awalnya takut air. Dalam 2 bulan anak saya sudah percaya diri mengapung dan berenang gaya dada dengan benar.",
				Rating:    5,
				AvatarURL: "https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?w=400&auto=format&fit=crop&q=80",
				SortOrder: 2,
				IsActive:  true,
			},
			{
				Name:      "Coach Dedy Kurniawan",
				RoleTitle: "Manajer Klub Millennium Aquatic",
				Content:   "Fitur input waktu lomba lewat tablet dan modul bagan otomatisnya sangat memudahkan panitia di lapangan. Hasil lomba langsung tayang di layar venue dan sertifikat langsung bisa diunduh mandiri.",
				Rating:    5,
				AvatarURL: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=400&auto=format&fit=crop&q=80",
				SortOrder: 3,
				IsActive:  true,
			},
		}
		for _, t := range testimonials {
			r.db.Create(&t)
		}
	}

	// 4. Default Swimming Events (Matching PRSI/FINA National Standard & Screenshot)
	var eventCount int64
	r.db.Model(&domain.SwimmingEvent{}).Count(&eventCount)
	if eventCount == 0 {
		events := []domain.SwimmingEvent{
			// KU 2 Putra (Matching Screenshot 2)
			{EventCode: 103, EventName: "100m Gaya Kupu-kupu", Distance: "100 METER", Stroke: "BUTTERFLY", Gender: "PUTRA", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "08:30 WIB", IsActive: true},
			{EventCode: 105, EventName: "50m Gaya Punggung", Distance: "50 METER", Stroke: "BACKSTROKE", Gender: "PUTRA", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "09:00 WIB", IsActive: true},
			{EventCode: 109, EventName: "100m Gaya Bebas", Distance: "100 METER", Stroke: "FREESTYLE", Gender: "PUTRA", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "10:20 WIB", IsActive: true},
			{EventCode: 111, EventName: "50m Gaya Kupu-kupu", Distance: "50 METER", Stroke: "BUTTERFLY", Gender: "PUTRA", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "11:15 WIB", IsActive: true},
			{EventCode: 115, EventName: "100m Gaya Punggung", Distance: "100 METER", Stroke: "BACKSTROKE", Gender: "PUTRA", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "14:00 WIB", IsActive: true},
			{EventCode: 119, EventName: "50m Gaya Dada", Distance: "50 METER", Stroke: "BREASTSTROKE", Gender: "PUTRA", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "15:00 WIB", IsActive: true},
			{EventCode: 123, EventName: "100m Gaya Dada", Distance: "100 METER", Stroke: "BREASTSTROKE", Gender: "PUTRA", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "16:20 WIB", IsActive: true},
			{EventCode: 125, EventName: "50m Gaya Bebas", Distance: "50 METER", Stroke: "FREESTYLE", Gender: "PUTRA", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "17:00 WIB", IsActive: true},

			// KU 2 Putri
			{EventCode: 104, EventName: "100m Gaya Kupu-kupu", Distance: "100 METER", Stroke: "BUTTERFLY", Gender: "PUTRI", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "08:45 WIB", IsActive: true},
			{EventCode: 106, EventName: "50m Gaya Punggung", Distance: "50 METER", Stroke: "BACKSTROKE", Gender: "PUTRI", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "09:15 WIB", IsActive: true},
			{EventCode: 110, EventName: "100m Gaya Bebas", Distance: "100 METER", Stroke: "FREESTYLE", Gender: "PUTRI", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "10:35 WIB", IsActive: true},
			{EventCode: 112, EventName: "50m Gaya Kupu-kupu", Distance: "50 METER", Stroke: "BUTTERFLY", Gender: "PUTRI", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "11:30 WIB", IsActive: true},
			{EventCode: 116, EventName: "100m Gaya Punggung", Distance: "100 METER", Stroke: "BACKSTROKE", Gender: "PUTRI", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "14:15 WIB", IsActive: true},
			{EventCode: 120, EventName: "50m Gaya Dada", Distance: "50 METER", Stroke: "BREASTSTROKE", Gender: "PUTRI", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "15:15 WIB", IsActive: true},
			{EventCode: 124, EventName: "100m Gaya Dada", Distance: "100 METER", Stroke: "BREASTSTROKE", Gender: "PUTRI", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "16:35 WIB", IsActive: true},
			{EventCode: 126, EventName: "50m Gaya Bebas", Distance: "50 METER", Stroke: "FREESTYLE", Gender: "PUTRI", AgeGroup: "KU 2", Fee: 150000, ScheduleTime: "17:15 WIB", IsActive: true},

			// KU 3 Putra & Putri
			{EventCode: 201, EventName: "50m Gaya Bebas", Distance: "50 METER", Stroke: "FREESTYLE", Gender: "PUTRA", AgeGroup: "KU 3", Fee: 150000, ScheduleTime: "08:00 WIB", IsActive: true},
			{EventCode: 202, EventName: "50m Gaya Bebas", Distance: "50 METER", Stroke: "FREESTYLE", Gender: "PUTRI", AgeGroup: "KU 3", Fee: 150000, ScheduleTime: "08:15 WIB", IsActive: true},
			{EventCode: 203, EventName: "50m Gaya Dada", Distance: "50 METER", Stroke: "BREASTSTROKE", Gender: "PUTRA", AgeGroup: "KU 3", Fee: 150000, ScheduleTime: "09:30 WIB", IsActive: true},
			{EventCode: 204, EventName: "50m Gaya Dada", Distance: "50 METER", Stroke: "BREASTSTROKE", Gender: "PUTRI", AgeGroup: "KU 3", Fee: 150000, ScheduleTime: "09:45 WIB", IsActive: true},

			// KU 4 Putra & Putri
			{EventCode: 301, EventName: "50m Gaya Bebas", Distance: "50 METER", Stroke: "FREESTYLE", Gender: "PUTRA", AgeGroup: "KU 4", Fee: 150000, ScheduleTime: "08:00 WIB", IsActive: true},
			{EventCode: 302, EventName: "50m Gaya Bebas", Distance: "50 METER", Stroke: "FREESTYLE", Gender: "PUTRI", AgeGroup: "KU 4", Fee: 150000, ScheduleTime: "08:15 WIB", IsActive: true},

			// KU 1 & Senior
			{EventCode: 101, EventName: "200m Gaya Ganti Perorangan", Distance: "200 METER", Stroke: "INDIVIDUALMEDLEY", Gender: "PUTRA", AgeGroup: "KU 1", Fee: 150000, ScheduleTime: "07:30 WIB", IsActive: true},
			{EventCode: 102, EventName: "200m Gaya Ganti Perorangan", Distance: "200 METER", Stroke: "INDIVIDUALMEDLEY", Gender: "PUTRI", AgeGroup: "KU 1", Fee: 150000, ScheduleTime: "07:45 WIB", IsActive: true},
		}
		for _, e := range events {
			r.db.Create(&e)
		}
	}

	// 5. Seed Sample Swimmers & Starting List Registrations (Matching PDF OCR data)
	var participantCount int64
	r.db.Model(&domain.Participant{}).Count(&participantCount)
	if participantCount == 0 {
		sampleSwimmers := []struct {
			Name      string
			Gender    string
			Club      string
			PIC       string
			Contact   string
			EventCode int
			TimeSeed  string
		}{
			{"MUHAMMAD MIRZA", "PUTRA", "ENSC", "DHERWINA", "081546206665", 115, "00.01.13"},
			{"MUHAMMAD MIRZA", "PUTRA", "ENSC", "DHERWINA", "081546206665", 121, "00.02.37"},
			{"PUTRI MAHIRA FAJRIANI", "PUTRI", "ENSC", "DHERWINA", "081546206665", 122, "99.99.99"},
			{"PUTRI MAHIRA FAJRIANI", "PUTRI", "ENSC", "DHERWINA", "081546206665", 108, "99.99.99"},
			{"KHAFIDZH SAMSURI", "PUTRA", "FORTIUS AKUATIK", "EKO", "085921507876", 115, "01.27.84"},
			{"KHAFIDZH SAMSURI", "PUTRA", "FORTIUS AKUATIK", "EKO", "085921507876", 123, "01.33.83"},
			{"KHAFIDZH SAMSURI", "PUTRA", "FORTIUS AKUATIK", "EKO", "085921507876", 109, "01.10.64"},
			{"KHAFIDZH SAMSURI", "PUTRA", "FORTIUS AKUATIK", "EKO", "085921507876", 113, "99.99.99"},
			{"KHAFIDZH SAMSURI", "PUTRA", "FORTIUS AKUATIK", "EKO", "085921507876", 125, "00.32.41"},
			{"KHILMA NUR SYAMSIL AISY", "PUTRI", "FORTIUS AKUATIK", "EKO", "085921507876", 116, "01.31.90"},
			{"KHILMA NUR SYAMSIL AISY", "PUTRI", "FORTIUS AKUATIK", "EKO", "085921507876", 104, "99.99.99"},
			{"KHILMA NUR SYAMSIL AISY", "PUTRI", "FORTIUS AKUATIK", "EKO", "085921507876", 110, "01.13.87"},
			{"KHILMA NUR SYAMSIL AISY", "PUTRI", "FORTIUS AKUATIK", "EKO", "085921507876", 122, "99.99.99"},
			{"KHILMA NUR SYAMSIL AISY", "PUTRI", "FORTIUS AKUATIK", "EKO", "085921507876", 114, "02.39.58"},
			{"KHILMA NUR SYAMSIL AISY", "PUTRI", "FORTIUS AKUATIK", "EKO", "085921507876", 126, "00.34.69"},
			{"SITI VIRLY MAQFIRAH", "PUTRI", "LUMBA SWIMMING CLUB", "ZAINAL ARIFIN", "083808002880", 116, "01.22.59"},
			{"SITI VIRLY MAQFIRAH", "PUTRI", "LUMBA SWIMMING CLUB", "ZAINAL ARIFIN", "083808002880", 106, "00.35.05"},
			{"AISYAH JANNATY MAULANA", "PUTRI", "MASC KOTA TANGERANG", "BPK FAJAR YOGANTARA", "085954761666", 126, "00.29.87"},
			{"AISYAH JANNATY MAULANA", "PUTRI", "MASC KOTA TANGERANG", "BPK FAJAR YOGANTARA", "085954761666", 114, "02.31.94"},
			{"AISYAH JANNATY MAULANA", "PUTRI", "MASC KOTA TANGERANG", "BPK FAJAR YOGANTARA", "085954761666", 112, "00.30.42"},
			{"AISYAH JANNATY MAULANA", "PUTRI", "MASC KOTA TANGERANG", "BPK FAJAR YOGANTARA", "085954761666", 102, "02.42.00"},
			{"AFKAR NAUFAL FADHIL", "PUTRA", "TIRTA BENTENG SC", "DEDE PRIYANA", "081386142058", 125, "00.37.05"},
			{"AFKAR NAUFAL FADHIL", "PUTRA", "TIRTA BENTENG SC", "DEDE PRIYANA", "081386142058", 109, "01.27.23"},
			{"AFKAR NAUFAL FADHIL", "PUTRA", "TIRTA BENTENG SC", "DEDE PRIYANA", "081386142058", 113, "03.17.43"},
			{"AFKAR NAUFAL FADHIL", "PUTRA", "TIRTA BENTENG SC", "DEDE PRIYANA", "081386142058", 105, "00.44.72"},
			{"AFKAR NAUFAL FADHIL", "PUTRA", "TIRTA BENTENG SC", "DEDE PRIYANA", "081386142058", 115, "01.45.75"},
			{"FAHMI KUNCORO", "PUTRA", "WDS AQUATIC BANTEN", "PRATAMA SIAHAAN", "081312004484", 101, "99.99.99"},
			{"ANANG SUBARKAH", "PUTRA", "MASC KOTA TANGERANG", "BPK FAJAR YOGANTARA", "085954761666", 101, "02.31.37"},
			{"FATIMAH ANAS", "PUTRI", "MASC KOTA TANGERANG", "BPK FAJAR YOGANTARA", "085954761666", 102, "02.34.76"},
		}

		pMap := make(map[string]domain.Participant)

		for idx, sample := range sampleSwimmers {
			p, exists := pMap[sample.Name]
			if !exists {
				p = domain.Participant{
					Name:    sample.Name,
					Gender:  sample.Gender,
					Club:    sample.Club,
					PIC:     sample.PIC,
					Contact: sample.Contact,
					Email:   fmt.Sprintf("swimmer%d@akuatik.id", idx+1),
				}
				r.db.Create(&p)
				pMap[sample.Name] = p
			}

			var event domain.SwimmingEvent
			if err := r.db.Where("event_code = ?", sample.EventCode).First(&event).Error; err == nil {
				reg := domain.Registration{
					RegistrationCode: fmt.Sprintf("SWIM-2025-%04d", idx+1),
					ParticipantID:    p.ID,
					SwimmingEventID:  event.ID,
					TimeSeed:         sample.TimeSeed,
					PaymentStatus:    "verified",
				}
				r.db.Create(&reg)
			}
		}
	}

	// 5. Default Page Sections for Dynamic Landing Page Ordering
	var sectionCount int64
	r.db.Model(&domain.PageSection{}).Count(&sectionCount)
	if sectionCount == 0 {
		defaultSections := []domain.PageSection{
			{PageSlug: "homepage", SectionCode: "hero", BadgeText: "BERANDA & HERO", Title: "Banner Utama & Header", Description: "Slider gambar, judul hero, tombol pendaftaran, dan counter statistik", SortOrder: 1, IsPublished: true},
			{PageSlug: "homepage", SectionCode: "programs", BadgeText: "PROGRAM PELATIHAN", Title: "Program Pelatihan Renang Unggulan", Description: "Pilihan kurikulum renang dari balita hingga kelas prestasi", SortOrder: 2, IsPublished: true},
			{PageSlug: "homepage", SectionCode: "coaches", BadgeText: "TIM PELATIH", Title: "Tim Kepelatihan Profesional", Description: "Profil pelatih berlisensi FINA / Akuatik Indonesia", SortOrder: 3, IsPublished: true},
			{PageSlug: "homepage", SectionCode: "facilities", BadgeText: "FASILITAS KOLAM", Title: "Fasilitas Standar Internasional", Description: "Infrastruktur kolam lomba dan sarana latihan pendukung", SortOrder: 4, IsPublished: true},
			{PageSlug: "homepage", SectionCode: "achievements", BadgeText: "PRESTASI ATLET", Title: "Rekam Jejak Prestasi & Medali", Description: "Pencapaian medali kejuaraan resmi tingkat daerah dan nasional", SortOrder: 5, IsPublished: true},
			{PageSlug: "homepage", SectionCode: "testimonials", BadgeText: "TESTIMONI", Title: "Kepuasan Orang Tua & Klub", Description: "Ulasan dan testimoni dari wali murid dan perwakilan perkumpulan", SortOrder: 6, IsPublished: true},
			{PageSlug: "homepage", SectionCode: "events", BadgeText: "NOMOR LOMBA", Title: "Jadwal & Nomor Acara Kejuaraan", Description: "Tabel nomor acara turnamen renang yang sedang dibuka", SortOrder: 7, IsPublished: true},
			{PageSlug: "homepage", SectionCode: "status_checker", BadgeText: "CEK STATUS", Title: "Cek Status & Validasi Pendaftaran", Description: "Form pelacakan resi pendaftaran peserta secara publik", SortOrder: 8, IsPublished: true},
		}
		for _, s := range defaultSections {
			r.db.Create(&s)
		}
	}

	return nil
}

// Database Query Methods
func (r *Repository) FindUserByUsername(username string) (*domain.User, error) {
	var user domain.User
	err := r.db.Preload("RoleRel").Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *Repository) FindBanners() ([]domain.BannerSlide, error) {
	var banners []domain.BannerSlide
	err := r.db.Order("sort_order asc").Find(&banners).Error
	return banners, err
}

func (r *Repository) FindEvents() ([]domain.SwimmingEvent, error) {
	var events []domain.SwimmingEvent
	err := r.db.Preload("Tournament").Where("is_active = ?", true).Order("event_code asc").Find(&events).Error
	return events, err
}

func (r *Repository) FindEventByID(id uint) (*domain.SwimmingEvent, error) {
	var event domain.SwimmingEvent
	err := r.db.First(&event, id).Error
	return &event, err
}

func (r *Repository) CreateParticipant(p *domain.Participant) error {
	return r.db.Create(p).Error
}

func (r *Repository) CreateRegistration(reg *domain.Registration) error {
	return r.db.Create(reg).Error
}

func (r *Repository) FindRegistrations() ([]domain.Registration, error) {
	var regs []domain.Registration
	err := r.db.Preload("Participant").Preload("SwimmingEvent").Preload("SwimmingEvent.Tournament").Order("id desc").Find(&regs).Error
	return regs, err
}

func (r *Repository) FindRegistrationByCode(code string) ([]domain.Registration, error) {
	var regs []domain.Registration
	err := r.db.Preload("Participant").Preload("SwimmingEvent").Where("registration_code = ?", code).Find(&regs).Error
	return regs, err
}

func (r *Repository) UpdateRegistrationStatus(id uint, status string) error {
	return r.db.Model(&domain.Registration{}).Where("id = ?", id).Update("payment_status", status).Error
}

func (r *Repository) FindPoolConfig() (*domain.PoolConfig, error) {
	var cfg domain.PoolConfig
	err := r.db.First(&cfg).Error
	return &cfg, err
}

func (r *Repository) UpdatePoolConfigMaxLanes(maxLanes int) error {
	return r.db.Model(&domain.PoolConfig{}).Where("id > 0").Update("max_lanes", maxLanes).Error
}

func (r *Repository) UpdateRegistrationHeatLine(id uint, heat, line int) error {
	return r.db.Model(&domain.Registration{}).Where("id = ?", id).Updates(map[string]interface{}{
		"heat_number": heat,
		"line_number": line,
	}).Error
}

func (r *Repository) FindRegistrationByID(id uint) (*domain.Registration, error) {
	var reg domain.Registration
	err := r.db.Preload("Participant").Preload("SwimmingEvent").First(&reg, id).Error
	return &reg, err
}

func (r *Repository) FindOccupiedRegistration(eventID uint, heat, line int, excludeID uint) (*domain.Registration, error) {
	var reg domain.Registration
	err := r.db.Preload("Participant").Preload("SwimmingEvent").
		Where("swimming_event_id = ? AND heat_number = ? AND line_number = ? AND id != ?", eventID, heat, line, excludeID).
		First(&reg).Error
	return &reg, err
}

func (r *Repository) FindTournamentByID(id uint) (*domain.Tournament, error) {
	var t domain.Tournament
	err := r.db.Preload("Events").First(&t, id).Error
	return &t, err
}

func (r *Repository) SetTournamentBukuAcaraLock(tournamentID uint, isLocked bool) error {
	return r.db.Model(&domain.Tournament{}).Where("id = ?", tournamentID).Update("is_buku_acara_locked", isLocked).Error
}

func (r *Repository) SetTournamentBukuAcaraPublish(tournamentID uint, isPublished bool) error {
	return r.db.Model(&domain.Tournament{}).Where("id = ?", tournamentID).Update("is_buku_acara_published", isPublished).Error
}

func (r *Repository) ResetAllHeatLines() error {
	return r.db.Model(&domain.Registration{}).Where("id > 0").Updates(map[string]interface{}{
		"heat_number": 0,
		"line_number": 0,
	}).Error
}

func (r *Repository) ResetTournamentHeatLines(tournamentID uint) error {
	if tournamentID > 0 {
		return r.db.Exec(`
			UPDATE registrations 
			SET heat_number = 0, line_number = 0 
			WHERE swimming_event_id IN (SELECT id FROM swimming_events WHERE tournament_id = ?)
		`, tournamentID).Error
	}
	return r.ResetAllHeatLines()
}

func (r *Repository) FindVerifiedRegistrations() ([]domain.Registration, error) {
	var regs []domain.Registration
	err := r.db.Preload("Participant").Preload("SwimmingEvent").Preload("SwimmingEvent.Tournament").
		Where("LOWER(payment_status) = ?", "verified").
		Find(&regs).Error
	return regs, err
}

func (r *Repository) UpdateRaceResult(id uint, timeStr string, rank int) error {
	return r.db.Model(&domain.Registration{}).Where("id = ?", id).Updates(map[string]interface{}{
		"race_result_time": timeStr,
		"rank":             rank,
	}).Error
}

func (r *Repository) UpdateFinalRaceResult(id uint, timeStr string, rank int) error {
	return r.db.Model(&domain.Registration{}).Where("id = ?", id).Updates(map[string]interface{}{
		"final_result_time": timeStr,
		"final_rank":        rank,
	}).Error
}

func (r *Repository) UpdateRegistrationFinalHeatLine(id uint, heat, line int) error {
	return r.db.Model(&domain.Registration{}).Where("id = ?", id).Updates(map[string]interface{}{
		"final_heat_number": heat,
		"final_line_number": line,
		"is_finalist":       heat > 0 && line > 0,
	}).Error
}

func (r *Repository) ResetTournamentFinalHeatLines(tournamentID uint) error {
	if tournamentID > 0 {
		return r.db.Exec(`
			UPDATE registrations 
			SET final_heat_number = 0, final_line_number = 0, is_finalist = false, final_result_time = '', final_rank = 0
			WHERE swimming_event_id IN (SELECT id FROM swimming_events WHERE tournament_id = ?)
		`, tournamentID).Error
	}
	return r.db.Model(&domain.Registration{}).Where("id > 0").Updates(map[string]interface{}{
		"final_heat_number": 0,
		"final_line_number": 0,
		"is_finalist":       false,
		"final_result_time": "",
		"final_rank":        0,
	}).Error
}

func (r *Repository) FindOccupiedFinalRegistration(eventID uint, heat, line int, excludeID uint) (*domain.Registration, error) {
	var reg domain.Registration
	err := r.db.Preload("Participant").Preload("SwimmingEvent").
		Where("swimming_event_id = ? AND final_heat_number = ? AND final_line_number = ? AND id != ?", eventID, heat, line, excludeID).
		First(&reg).Error
	return &reg, err
}

func (r *Repository) SaveBanner(banner *domain.BannerSlide) error {
	if banner.ID > 0 {
		return r.db.Save(banner).Error
	}
	return r.db.Create(banner).Error
}

func (r *Repository) DeleteBanner(id uint) error {
	return r.db.Delete(&domain.BannerSlide{}, id).Error
}

func (r *Repository) FindHeroConfig() (*domain.HeroConfig, error) {
	var cfg domain.HeroConfig
	err := r.db.First(&cfg).Error
	return &cfg, err
}

func (r *Repository) SaveHeroConfig(cfg *domain.HeroConfig) error {
	if cfg.ID > 0 {
		return r.db.Save(cfg).Error
	}
	var existing domain.HeroConfig
	if err := r.db.First(&existing).Error; err == nil {
		cfg.ID = existing.ID
		return r.db.Save(cfg).Error
	}
	return r.db.Create(cfg).Error
}

func (r *Repository) FindHeroStats() ([]domain.HeroStat, error) {
	var stats []domain.HeroStat
	err := r.db.Order("sort_order asc").Find(&stats).Error
	return stats, err
}

func (r *Repository) SaveHeroStats(stats []domain.HeroStat) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&domain.HeroStat{}).Error; err != nil {
			return err
		}
		for idx, s := range stats {
			s.ID = 0
			s.SortOrder = idx + 1
			if err := tx.Create(&s).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) FindSiteConfig() (*domain.SiteConfig, error) {
	var cfg domain.SiteConfig
	err := r.db.First(&cfg).Error
	return &cfg, err
}

func (r *Repository) SaveSiteConfig(cfg *domain.SiteConfig) error {
	if cfg.ID > 0 {
		return r.db.Save(cfg).Error
	}
	var existing domain.SiteConfig
	if err := r.db.First(&existing).Error; err == nil {
		cfg.ID = existing.ID
		return r.db.Save(cfg).Error
	}
	return r.db.Create(cfg).Error
}

func (r *Repository) FindProgramSectionConfig() (*domain.ProgramSectionConfig, error) {
	var cfg domain.ProgramSectionConfig
	err := r.db.First(&cfg).Error
	return &cfg, err
}

func (r *Repository) SaveProgramSectionConfig(cfg *domain.ProgramSectionConfig) error {
	if cfg.ID > 0 {
		return r.db.Save(cfg).Error
	}
	var existing domain.ProgramSectionConfig
	if err := r.db.First(&existing).Error; err == nil {
		cfg.ID = existing.ID
		return r.db.Save(cfg).Error
	}
	return r.db.Create(cfg).Error
}

func (r *Repository) FindTrainingPrograms() ([]domain.TrainingProgram, error) {
	var progs []domain.TrainingProgram
	err := r.db.Order("sort_order asc, id asc").Find(&progs).Error
	return progs, err
}

func (r *Repository) SaveTrainingProgram(prog *domain.TrainingProgram) error {
	if prog.ID > 0 {
		return r.db.Save(prog).Error
	}
	return r.db.Create(prog).Error
}

func (r *Repository) DeleteTrainingProgram(id uint) error {
	return r.db.Delete(&domain.TrainingProgram{}, id).Error
}

func (r *Repository) FindCoachSectionConfig() (*domain.CoachSectionConfig, error) {
	var cfg domain.CoachSectionConfig
	err := r.db.First(&cfg).Error
	return &cfg, err
}

func (r *Repository) SaveCoachSectionConfig(cfg *domain.CoachSectionConfig) error {
	if cfg.ID > 0 {
		return r.db.Save(cfg).Error
	}
	var existing domain.CoachSectionConfig
	if err := r.db.First(&existing).Error; err == nil {
		cfg.ID = existing.ID
		return r.db.Save(cfg).Error
	}
	return r.db.Create(cfg).Error
}

func (r *Repository) FindCoaches() ([]domain.Coach, error) {
	var coaches []domain.Coach
	err := r.db.Order("sort_order asc, id asc").Find(&coaches).Error
	return coaches, err
}

func (r *Repository) SaveCoach(c *domain.Coach) error {
	if c.ID > 0 {
		return r.db.Save(c).Error
	}
	return r.db.Create(c).Error
}

func (r *Repository) DeleteCoach(id uint) error {
	return r.db.Delete(&domain.Coach{}, id).Error
}

func (r *Repository) FindFacilitySectionConfig() (*domain.FacilitySectionConfig, error) {
	var cfg domain.FacilitySectionConfig
	err := r.db.First(&cfg).Error
	return &cfg, err
}

func (r *Repository) SaveFacilitySectionConfig(cfg *domain.FacilitySectionConfig) error {
	if cfg.ID > 0 {
		return r.db.Save(cfg).Error
	}
	var existing domain.FacilitySectionConfig
	if err := r.db.First(&existing).Error; err == nil {
		cfg.ID = existing.ID
		return r.db.Save(cfg).Error
	}
	return r.db.Create(cfg).Error
}

func (r *Repository) FindFacilities() ([]domain.Facility, error) {
	var facilities []domain.Facility
	err := r.db.Order("sort_order asc, id asc").Find(&facilities).Error
	return facilities, err
}

func (r *Repository) SaveFacility(f *domain.Facility) error {
	if f.ID > 0 {
		return r.db.Save(f).Error
	}
	return r.db.Create(f).Error
}

func (r *Repository) DeleteFacility(id uint) error {
	return r.db.Delete(&domain.Facility{}, id).Error
}

// Achievements DB Methods
func (r *Repository) FindAchievementSectionConfig() (*domain.AchievementSectionConfig, error) {
	var cfg domain.AchievementSectionConfig
	err := r.db.First(&cfg).Error
	return &cfg, err
}

func (r *Repository) SaveAchievementSectionConfig(cfg *domain.AchievementSectionConfig) error {
	if cfg.ID > 0 {
		return r.db.Save(cfg).Error
	}
	var existing domain.AchievementSectionConfig
	if err := r.db.First(&existing).Error; err == nil {
		cfg.ID = existing.ID
		return r.db.Save(cfg).Error
	}
	return r.db.Create(cfg).Error
}

func (r *Repository) FindAchievements() ([]domain.Achievement, error) {
	var achievements []domain.Achievement
	err := r.db.Order("sort_order asc, id asc").Find(&achievements).Error
	return achievements, err
}

func (r *Repository) SaveAchievement(a *domain.Achievement) error {
	if a.ID > 0 {
		return r.db.Save(a).Error
	}
	return r.db.Create(a).Error
}

func (r *Repository) DeleteAchievement(id uint) error {
	return r.db.Delete(&domain.Achievement{}, id).Error
}

// Testimonials DB Methods
func (r *Repository) FindTestimonialSectionConfig() (*domain.TestimonialSectionConfig, error) {
	var cfg domain.TestimonialSectionConfig
	err := r.db.First(&cfg).Error
	return &cfg, err
}

func (r *Repository) SaveTestimonialSectionConfig(cfg *domain.TestimonialSectionConfig) error {
	if cfg.ID > 0 {
		return r.db.Save(cfg).Error
	}
	var existing domain.TestimonialSectionConfig
	if err := r.db.First(&existing).Error; err == nil {
		cfg.ID = existing.ID
		return r.db.Save(cfg).Error
	}
	return r.db.Create(cfg).Error
}

func (r *Repository) FindTestimonials() ([]domain.Testimonial, error) {
	var testimonials []domain.Testimonial
	err := r.db.Order("sort_order asc, id asc").Find(&testimonials).Error
	return testimonials, err
}

func (r *Repository) SaveTestimonial(t *domain.Testimonial) error {
	if t.ID > 0 {
		return r.db.Save(t).Error
	}
	return r.db.Create(t).Error
}

func (r *Repository) DeleteTestimonial(id uint) error {
	return r.db.Delete(&domain.Testimonial{}, id).Error
}

// Tournaments DB Methods
func (r *Repository) FindTournaments() ([]domain.Tournament, error) {
	var tourneys []domain.Tournament
	err := r.db.Preload("Events").Order("id asc").Find(&tourneys).Error
	return tourneys, err
}

func (r *Repository) SaveTournament(t *domain.Tournament) error {
	if t.ID > 0 {
		return r.db.Save(t).Error
	}
	return r.db.Create(t).Error
}

func (r *Repository) DeleteTournament(id uint) error {
	return r.db.Delete(&domain.Tournament{}, id).Error
}

// Events DB Methods

func (r *Repository) CreateEvent(evt *domain.SwimmingEvent) error {
	return r.db.Create(evt).Error
}

func (r *Repository) SaveEvent(evt *domain.SwimmingEvent) error {
	if evt.ID > 0 {
		return r.db.Save(evt).Error
	}
	return r.db.Create(evt).Error
}

func (r *Repository) DeleteEvent(id uint) error {
	return r.db.Delete(&domain.SwimmingEvent{}, id).Error
}

// PageSections DB Methods
func (r *Repository) FindPageSections(pageSlug string) ([]domain.PageSection, error) {
	var sections []domain.PageSection
	if pageSlug == "" {
		pageSlug = "homepage"
	}
	err := r.db.Where("page_slug = ?", pageSlug).Order("sort_order asc").Find(&sections).Error
	return sections, err
}

func (r *Repository) SavePageSection(sec *domain.PageSection) error {
	if sec.ID > 0 {
		return r.db.Save(sec).Error
	}
	return r.db.Create(sec).Error
}

func (r *Repository) BatchSavePageSections(sections []domain.PageSection) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, s := range sections {
			if s.ID > 0 {
				if err := tx.Model(&domain.PageSection{}).Where("id = ?", s.ID).Updates(map[string]interface{}{
					"sort_order":   s.SortOrder,
					"is_published": s.IsPublished,
					"title":        s.Title,
					"badge_text":   s.BadgeText,
					"description":  s.Description,
				}).Error; err != nil {
					return err
				}
			} else if s.SectionCode != "" {
				if err := tx.Model(&domain.PageSection{}).Where("section_code = ?", s.SectionCode).Updates(map[string]interface{}{
					"sort_order":   s.SortOrder,
					"is_published": s.IsPublished,
					"title":        s.Title,
					"badge_text":   s.BadgeText,
					"description":  s.Description,
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *Repository) ResetPageSections(pageSlug string) error {
	if pageSlug == "" {
		pageSlug = "homepage"
	}
	// Delete existing sections for this page slug
	if err := r.db.Where("page_slug = ?", pageSlug).Delete(&domain.PageSection{}).Error; err != nil {
		return err
	}
	// Re-insert default sections
	defaultSections := []domain.PageSection{
		{PageSlug: pageSlug, SectionCode: "hero", BadgeText: "BERANDA & HERO", Title: "Banner Utama & Header", Description: "Slider gambar, judul hero, tombol pendaftaran, dan counter statistik", SortOrder: 1, IsPublished: true},
		{PageSlug: pageSlug, SectionCode: "programs", BadgeText: "PROGRAM PELATIHAN", Title: "Program Pelatihan Renang Unggulan", Description: "Pilihan kurikulum renang dari balita hingga kelas prestasi", SortOrder: 2, IsPublished: true},
		{PageSlug: pageSlug, SectionCode: "coaches", BadgeText: "TIM PELATIH", Title: "Tim Kepelatihan Profesional", Description: "Profil pelatih berlisensi FINA / Akuatik Indonesia", SortOrder: 3, IsPublished: true},
		{PageSlug: pageSlug, SectionCode: "facilities", BadgeText: "FASILITAS KOLAM", Title: "Fasilitas Standar Internasional", Description: "Infrastruktur kolam lomba dan sarana latihan pendukung", SortOrder: 4, IsPublished: true},
		{PageSlug: pageSlug, SectionCode: "achievements", BadgeText: "PRESTASI ATLET", Title: "Rekam Jejak Prestasi & Medali", Description: "Pencapaian medali kejuaraan resmi tingkat daerah dan nasional", SortOrder: 5, IsPublished: true},
		{PageSlug: pageSlug, SectionCode: "testimonials", BadgeText: "TESTIMONI", Title: "Kepuasan Orang Tua & Klub", Description: "Ulasan dan testimoni dari wali murid dan perwakilan perkumpulan", SortOrder: 6, IsPublished: true},
		{PageSlug: pageSlug, SectionCode: "events", BadgeText: "NOMOR LOMBA", Title: "Jadwal & Nomor Acara Kejuaraan", Description: "Tabel nomor acara turnamen renang yang sedang dibuka", SortOrder: 7, IsPublished: true},
		{PageSlug: pageSlug, SectionCode: "status_checker", BadgeText: "CEK STATUS", Title: "Cek Status & Validasi Pendaftaran", Description: "Form pelacakan resi pendaftaran peserta secara publik", SortOrder: 8, IsPublished: true},
	}
	for _, s := range defaultSections {
		if err := r.db.Create(&s).Error; err != nil {
			return err
		}
	}
	return nil
}

// ----------------------------------------------------
// RACE RESULT AUDIT LOG METHODS
// ----------------------------------------------------

func (r *Repository) CreateRaceResultLog(log *domain.RaceResultLog) error {
	return r.db.Create(log).Error
}

func (r *Repository) FindRaceResultLogs(tournamentID uint, round string, action string, search string, limit, offset int) ([]domain.RaceResultLog, int64, error) {
	var logs []domain.RaceResultLog
	var total int64

	query := r.db.Model(&domain.RaceResultLog{})

	if tournamentID > 0 {
		query = query.Where("tournament_id = ?", tournamentID)
	}
	if round != "" && round != "ALL" {
		query = query.Where("LOWER(round) = ?", strings.ToLower(round))
	}
	if action != "" && action != "ALL" {
		query = query.Where("action = ?", action)
	}
	if strings.TrimSpace(search) != "" {
		like := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		query = query.Where("LOWER(swimmer_name) LIKE ? OR LOWER(club_name) LIKE ? OR LOWER(event_name) LIKE ? OR LOWER(operator_name) LIKE ? OR LOWER(notes) LIKE ?",
			like, like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *Repository) GetRaceResultLogStats(tournamentID uint) (map[string]interface{}, error) {
	var totalLogs, totalCreate, totalUpdate, totalDelete, totalSwap int64

	baseQuery := r.db.Model(&domain.RaceResultLog{})
	if tournamentID > 0 {
		baseQuery = baseQuery.Where("tournament_id = ?", tournamentID)
	}

	baseQuery.Count(&totalLogs)

	qCreate := r.db.Model(&domain.RaceResultLog{}).Where("action = ?", "CREATE")
	qUpdate := r.db.Model(&domain.RaceResultLog{}).Where("action = ?", "UPDATE")
	qDelete := r.db.Model(&domain.RaceResultLog{}).Where("action = ?", "DELETE")
	qSwap := r.db.Model(&domain.RaceResultLog{}).Where("action IN (?)", []string{"SWAP", "MOVE"})

	if tournamentID > 0 {
		qCreate = qCreate.Where("tournament_id = ?", tournamentID)
		qUpdate = qUpdate.Where("tournament_id = ?", tournamentID)
		qDelete = qDelete.Where("tournament_id = ?", tournamentID)
		qSwap = qSwap.Where("tournament_id = ?", tournamentID)
	}

	qCreate.Count(&totalCreate)
	qUpdate.Count(&totalUpdate)
	qDelete.Count(&totalDelete)
	qSwap.Count(&totalSwap)

	return map[string]interface{}{
		"total_logs":   totalLogs,
		"total_create": totalCreate,
		"total_update": totalUpdate,
		"total_delete": totalDelete,
		"total_swap":   totalSwap,
	}, nil
}

// --------------------------------------------------------------------------
// ROLE & USER MANAGEMENT REPOSITORY METHODS
// --------------------------------------------------------------------------

func (r *Repository) FindRoles() ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.Order("is_system desc, id asc").Find(&roles).Error
	return roles, err
}

func (r *Repository) FindRoleByID(id uint) (*domain.Role, error) {
	var role domain.Role
	err := r.db.First(&role, id).Error
	return &role, err
}

func (r *Repository) CreateRole(role *domain.Role) error {
	return r.db.Create(role).Error
}

func (r *Repository) UpdateRole(role *domain.Role) error {
	return r.db.Save(role).Error
}

func (r *Repository) DeleteRole(id uint) error {
	return r.db.Delete(&domain.Role{}, id).Error
}

func (r *Repository) CountUsersByRoleID(roleID uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

func (r *Repository) FindUsers() ([]domain.User, error) {
	var users []domain.User
	err := r.db.Preload("RoleRel").Order("id asc").Find(&users).Error
	return users, err
}

func (r *Repository) FindUserByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.Preload("RoleRel").First(&user, id).Error
	return &user, err
}

func (r *Repository) CreateUser(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) UpdateUser(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *Repository) DeleteUser(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}








