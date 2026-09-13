package domain

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:100;uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"size:150;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Role         string    `gorm:"size:50;default:'ADMIN'" json:"role"`
	Status       string    `gorm:"size:20;default:'active'" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type MasterMenu struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	ParentID  *uint        `gorm:"index" json:"parent_id"`
	Children  []MasterMenu `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Name      string       `gorm:"size:100;not null" json:"name"`
	Slug      string       `gorm:"size:150;uniqueIndex;not null" json:"slug"`
	URL       string       `gorm:"size:255;not null" json:"url"`
	SortOrder int          `gorm:"default:0" json:"sort_order"`
	Icon      string       `gorm:"size:50" json:"icon"`
	IsVisible bool         `gorm:"default:true" json:"is_visible"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type BannerSlide struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Title            string    `gorm:"size:255;not null" json:"title"`
	BadgeText        string    `gorm:"size:150" json:"badge_text"`
	Description      string    `gorm:"type:text" json:"description"`
	ImageURL         string    `gorm:"type:text;not null" json:"image_url"`
	CTAPrimaryText   string    `gorm:"size:100" json:"cta_primary_text"`
	CTAPrimaryURL    string    `gorm:"size:255" json:"cta_primary_url"`
	CTASecondaryText string    `gorm:"size:100" json:"cta_secondary_text"`
	CTASecondaryURL  string    `gorm:"size:255" json:"cta_secondary_url"`
	SortOrder        int       `gorm:"default:0" json:"sort_order"`
	Status           string    `gorm:"size:20;default:'published'" json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type PageSection struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PageSlug       string    `gorm:"size:100;index;default:'homepage'" json:"page_slug"`
	SectionCode    string    `gorm:"size:100;uniqueIndex;not null" json:"section_code"`
	BadgeText      string    `gorm:"size:150" json:"badge_text"`
	Title          string    `gorm:"size:255;not null" json:"title"`
	Description    string    `gorm:"type:text" json:"description"`
	ContentPayload string    `gorm:"type:text" json:"content_payload"`
	SortOrder      int       `gorm:"default:0" json:"sort_order"`
	IsPublished    bool      `gorm:"default:true" json:"is_published"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Tournament struct {
	ID                    uint            `gorm:"primaryKey" json:"id"`
	Name                  string          `gorm:"size:255;not null" json:"name"`
	Description           string          `gorm:"type:text" json:"description"`
	Location              string          `gorm:"size:255" json:"location"`
	RegistrationStartDate string          `gorm:"size:50" json:"registration_start_date"` // YYYY-MM-DD
	RegistrationEndDate   string          `gorm:"size:50" json:"registration_end_date"`   // YYYY-MM-DD
	EventStartDate        string          `gorm:"size:50" json:"event_start_date"`        // YYYY-MM-DD
	EventEndDate          string          `gorm:"size:50" json:"event_end_date"`          // YYYY-MM-DD
	IsActive              bool            `gorm:"default:true" json:"is_active"`
	Events                []SwimmingEvent `gorm:"foreignKey:TournamentID" json:"events,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type SwimmingEvent struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	TournamentID uint        `gorm:"index;default:1" json:"tournament_id"`
	Tournament   *Tournament `gorm:"foreignKey:TournamentID" json:"tournament,omitempty"`
	EventCode    int         `gorm:"index;not null" json:"event_code"` // e.g. 101, 102, 103...
	EventName    string      `gorm:"size:255;not null" json:"event_name"`   // e.g. 100m Gaya Kupu-kupu
	Distance     string      `gorm:"size:50;not null" json:"distance"`     // e.g. 50 METER, 100 METER, 200 METER
	Stroke       string      `gorm:"size:50;not null" json:"stroke"`       // FREESTYLE, BREASTSTROKE, BACKSTROKE, BUTTERFLY, INDIVIDUALMEDLEY
	Gender       string      `gorm:"size:20;not null" json:"gender"`       // PUTRA, PUTRI
	AgeGroup     string      `gorm:"size:50;default:'OPEN'" json:"age_group"` // KU 4, KU 3, KU 2, KU 1, Senior, OPEN
	Fee          float64     `gorm:"default:150000" json:"fee"`
	ScheduleTime string      `gorm:"size:100;default:'08:00 WIB'" json:"schedule_time"`
	IsActive     bool        `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type Participant struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	Name                string    `gorm:"size:255;not null" json:"name"`
	Gender              string    `gorm:"size:20;not null" json:"gender"` // PUTRA / PUTRI
	Club                string    `gorm:"size:255;not null" json:"club"`  // e.g. MASC Swim Club / SD Al-Azhar
	PIC                 string    `gorm:"size:150" json:"pic"`
	Contact             string    `gorm:"size:50" json:"contact"`
	Email               string    `gorm:"size:150" json:"email"`
	BirthDate           string    `gorm:"size:50" json:"birth_date"`
	AgeGroup            string    `gorm:"size:50" json:"age_group"`
	VerificationDocType string    `gorm:"size:100;default:'Akte Kelahiran'" json:"verification_doc_type"`
	VerificationDocURL  string    `gorm:"type:text" json:"verification_doc_url"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Registration struct {
	ID               uint          `gorm:"primaryKey" json:"id"`
	RegistrationCode string        `gorm:"size:50;uniqueIndex;not null" json:"registration_code"` // e.g. REG-ASC-35552
	ParticipantID    uint          `gorm:"index;not null" json:"participant_id"`
	Participant      Participant   `gorm:"foreignKey:ParticipantID" json:"participant,omitempty"`
	SwimmingEventID  uint          `gorm:"index;not null" json:"swimming_event_id"`
	SwimmingEvent    SwimmingEvent `gorm:"foreignKey:SwimmingEventID" json:"swimming_event,omitempty"`
	TimeSeed         string        `gorm:"size:50;not null;default:'99.99.99'" json:"time_seed"` // mm.ss.ms format, e.g. 00:30.00 or NT
	HeatNumber       int           `gorm:"default:0" json:"heat_number"`                         // Assigned in Buku Acara
	LineNumber       int           `gorm:"default:0" json:"line_number"`                         // Assigned in Buku Acara
	PaymentStatus    string        `gorm:"size:30;default:'pending'" json:"payment_status"`       // pending, verified, rejected
	PaymentMethod    string        `gorm:"size:100;default:'Transfer Bank BCA'" json:"payment_method"`
	SenderBankOwner  string        `gorm:"size:255" json:"sender_bank_owner"`
	PaymentProofURL  string        `gorm:"type:text" json:"payment_proof_url"`
	RaceResultTime   string        `gorm:"size:50" json:"race_result_time"` // Catatan waktu hasil lomba
	Rank             int           `gorm:"default:0" json:"rank"`            // Juara 1, 2, 3...
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

type PoolConfig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	CompetitionName string    `gorm:"size:255;default:'TIME TRIAL 2025 AKUATIK INDONESIA KOTA TANGERANG'" json:"competition_name"`
	Location        string    `gorm:"size:255;default:'Kolam Renang Gelanggang Kota Tangerang'" json:"location"`
	MaxLanes        int       `gorm:"default:3" json:"max_lanes"` // Number of pool lines (default 3 lines as in sample PDF)
	UpdatedAt       time.Time `json:"updated_at"`
}

type HeroConfig struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	BadgeText        string    `gorm:"size:255" json:"badge_text"`
	TitlePrefix      string    `gorm:"size:255" json:"title_prefix"`
	TitleHighlight   string    `gorm:"size:255" json:"title_highlight"`
	Subtitle         string    `gorm:"type:text" json:"subtitle"`
	Feature1         string    `gorm:"size:255" json:"feature_1"`
	Feature2         string    `gorm:"size:255" json:"feature_2"`
	Feature3         string    `gorm:"size:255" json:"feature_3"`
	Feature4         string    `gorm:"size:255" json:"feature_4"`
	CTAPrimaryText   string    `gorm:"size:100" json:"cta_primary_text"`
	CTAPrimaryURL    string    `gorm:"size:255" json:"cta_primary_url"`
	CTASecondaryText string    `gorm:"size:100" json:"cta_secondary_text"`
	CTASecondaryURL  string    `gorm:"size:255" json:"cta_secondary_url"`
	CTATertiaryText  string    `gorm:"size:100" json:"cta_tertiary_text"`
	CTATertiaryURL   string    `gorm:"size:255" json:"cta_tertiary_url"`
	NoteText         string    `gorm:"type:text" json:"note_text"`
	TrustText1       string    `gorm:"size:255" json:"trust_text_1"`
	TrustText2       string    `gorm:"size:255" json:"trust_text_2"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type HeroStat struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Value     string    `gorm:"size:100;not null" json:"value"`
	Label     string    `gorm:"size:150;not null" json:"label"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Action    string    `gorm:"size:100;not null" json:"action"`
	Entity    string    `gorm:"size:100;not null" json:"entity"`
	EntityID  string    `gorm:"size:100" json:"entity_id"`
	OldValues string    `gorm:"type:text" json:"old_values"`
	NewValues string    `gorm:"type:text" json:"new_values"`
	IPAddress string    `gorm:"size:45" json:"ip_address"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type SiteConfig struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	AppName           string    `gorm:"size:255;default:'AKUATIK TANGERANG'" json:"app_name"`
	AppTagline        string    `gorm:"size:255;default:'TIME TRIAL CHAMPIONSHIP 2025'" json:"app_tagline"`
	LogoURL           string    `gorm:"type:text" json:"logo_url"`
	WANumber          string    `gorm:"size:50;default:'6281234567890'" json:"wa_number"`
	DefaultWATemplate string    `gorm:"type:text" json:"default_wa_template"`
	FooterDescription string    `gorm:"type:text;default:'Membangun Karakter, Mengasah Teknik, Mencetak Juara Renang Masa Depan'" json:"footer_description"`
	FooterText        string    `gorm:"type:text" json:"footer_text"`
	Email             string    `gorm:"size:150;default:'info@akuatik-tangerang.id'" json:"email"`
	Address           string    `gorm:"type:text;default:'Aquatic Center Complex, Jl. Pintu Satu Senayan No. 8, Jakarta Pusat'" json:"address"`
	YoutubeURL        string    `gorm:"size:255" json:"youtube_url"`
	InstagramURL      string    `gorm:"size:255" json:"instagram_url"`
	TiktokURL         string    `gorm:"size:255" json:"tiktok_url"`
	FacebookURL       string    `gorm:"size:255" json:"facebook_url"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type TrainingProgram struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Category     string    `gorm:"size:100;not null" json:"category"`
	Title        string    `gorm:"size:255;not null" json:"title"`
	Subtitle     string    `gorm:"size:255" json:"subtitle"`
	AgeBadge     string    `gorm:"size:100" json:"age_badge"`
	PopularBadge string    `gorm:"size:100" json:"popular_badge"`
	Description  string    `gorm:"type:text" json:"description"`
	Features     string    `gorm:"type:text" json:"features"`
	Price        string    `gorm:"size:100" json:"price"`
	ImageURL     string    `gorm:"type:text" json:"image_url"`
	WATemplate   string    `gorm:"type:text" json:"wa_template"`
	SortOrder    int       `gorm:"default:0" json:"sort_order"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProgramSectionConfig struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	BadgeText            string    `gorm:"size:255;default:'KURIKULUM BERJENJANG & TERSTRUKTUR'" json:"badge_text"`
	Title                string    `gorm:"size:255;default:'Program Pelatihan Renang Unggulan'" json:"title"`
	Subtitle             string    `gorm:"type:text" json:"subtitle"`
	AssessmentTitle      string    `gorm:"size:255;default:'Bingung Memilih Kelas yang Tepat untuk Anak Anda?'" json:"assessment_title"`
	AssessmentSubtitle   string    `gorm:"type:text" json:"assessment_subtitle"`
	AssessmentButtonText string    `gorm:"size:150;default:'Jadwalkan Free Assessment ->'" json:"assessment_button_text"`
	AssessmentWATemplate string    `gorm:"type:text" json:"assessment_wa_template"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type Coach struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Name             string    `gorm:"size:255;not null" json:"name"`
	RoleTitle        string    `gorm:"size:255;not null" json:"role_title"`
	LicenseBadge     string    `gorm:"size:255" json:"license_badge"`
	Experience       string    `gorm:"size:255" json:"experience"`
	Description      string    `gorm:"type:text" json:"description"`
	VerificationText string    `gorm:"size:255;default:'Verified Coach PB PRSI'" json:"verification_text"`
	PhotoURL         string    `gorm:"type:text" json:"photo_url"`
	SortOrder        int       `gorm:"default:0" json:"sort_order"`
	IsActive         bool      `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CoachSectionConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BadgeText string    `gorm:"size:255;default:'TIM KEPELATIHAN PROFESIONAL'" json:"badge_text"`
	Title     string    `gorm:"size:255;default:'Didampingi Pelatih Bersertifikasi Nasional & FINA'" json:"title"`
	Subtitle  string    `gorm:"type:text" json:"subtitle"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Facility struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	TagText     string    `gorm:"size:100" json:"tag_text"` // Floating badge on top right of image, e.g., '50m x 25m'
	Description string    `gorm:"type:text" json:"description"`
	SpecsText   string    `gorm:"size:255" json:"specs_text"` // Bottom specs text with check mark
	ImageURL    string    `gorm:"type:text" json:"image_url"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FacilitySectionConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BadgeText string    `gorm:"size:255;default:'FASILITAS & STANDAR KOLAM'" json:"badge_text"`
	Title     string    `gorm:"size:255;default:'Infrastruktur Kolam Renang Standar Internasional'" json:"title"`
	Subtitle  string    `gorm:"type:text" json:"subtitle"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Achievement struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Title      string    `gorm:"size:255;not null" json:"title"` // e.g., 'Juara Umum 1 Kejurnas Renang Pelajar'
	Year       string    `gorm:"size:50;not null" json:"year"`   // e.g., '2026'
	MedalType  string    `gorm:"size:50;default:'gold'" json:"medal_type"` // 'gold', 'silver', 'bronze'
	EventName  string    `gorm:"size:255" json:"event_name"`     // e.g., 'Kejurnas Antar Perkumpulan Renang 2026'
	WinnerName string    `gorm:"size:255" json:"winner_name"`    // e.g., 'Tim Prestasi MASC Swim Club' or 'Rayhan Al Fatih'
	SortOrder  int       `gorm:"default:0" json:"sort_order"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AchievementSectionConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BadgeText string    `gorm:"size:255;default:'REKAM JEJAK PRESTASI'" json:"badge_text"`
	Title     string    `gorm:"size:255;default:'Pencapaian Medali & Kejuaraan Resmi'" json:"title"`
	Subtitle  string    `gorm:"type:text" json:"subtitle"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Testimonial struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`       // e.g., 'Bapak Hendra Wijaya'
	RoleTitle string    `gorm:"size:255;not null" json:"role_title"` // e.g., 'Orang Tua Atlet (Rayhan, KU 3)'
	Content   string    `gorm:"type:text;not null" json:"content"`   // Quote content
	Rating    int       `gorm:"default:5" json:"rating"`             // 1 to 5 stars
	AvatarURL string    `gorm:"type:text" json:"avatar_url"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TestimonialSectionConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BadgeText string    `gorm:"size:255;default:'KEPUASAN ORANG TUA & OFFICIAL KLUB'" json:"badge_text"`
	Title     string    `gorm:"size:255;default:'Apa Kata Mereka Tentang MASC Swim?'" json:"title"`
	Subtitle  string    `gorm:"type:text" json:"subtitle"`
	UpdatedAt time.Time `json:"updated_at"`
}





