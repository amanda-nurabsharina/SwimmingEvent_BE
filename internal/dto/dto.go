package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  UserSummary `json:"user"`
}

type UserSummary struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	RoleID      *uint    `json:"role_id,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type RoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type RoleResponse struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	IsSystem    bool     `json:"is_system"`
	UsersCount  int      `json:"users_count"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type UserCreateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleID   uint   `json:"role_id"`
	Status   string `json:"status"`
}

type UserUpdateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	RoleID   uint   `json:"role_id"`
	Status   string `json:"status"`
}

type UserDetailResponse struct {
	ID        uint          `json:"id"`
	Username  string        `json:"username"`
	Email     string        `json:"email"`
	RoleID    *uint         `json:"role_id"`
	RoleName  string        `json:"role_name"`
	Role      *RoleResponse `json:"role,omitempty"`
	Status    string        `json:"status"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

type RegisterParticipantRequest struct {
	Name                string                 `json:"name" binding:"required"`
	Gender              string                 `json:"gender" binding:"required"`
	Club                string                 `json:"club" binding:"required"`
	PIC                 string                 `json:"pic"`
	Contact             string                 `json:"contact"`
	Email               string                 `json:"email"`
	BirthDate           string                 `json:"birth_date"`
	AgeGroup            string                 `json:"age_group"`
	VerificationDocType string                 `json:"verification_doc_type"`
	VerificationDocURL  string                 `json:"verification_doc_url"`
	PaymentMethod       string                 `json:"payment_method"`
	SenderBankOwner     string                 `json:"sender_bank_owner"`
	PaymentProofURL     string                 `json:"payment_proof_url"`
	EventSelections     []EventSelectionDetail `json:"event_selections" binding:"required"`
}

type EventSelectionDetail struct {
	SwimmingEventID uint   `json:"swimming_event_id" binding:"required"`
	TimeSeed        string `json:"time_seed"` // e.g. 00:30.00 or NT
}

type RegisterParticipantResponse struct {
	RegistrationCode string   `json:"registration_code"`
	ParticipantName  string   `json:"participant_name"`
	TotalEvents      int      `json:"total_events"`
	PaymentStatus    string   `json:"payment_status"`
	Message          string   `json:"message"`
}

type VerificationRequest struct {
	PaymentStatus string `json:"payment_status"` // verified / rejected
}

type GenerateBukuAcaraRequest struct {
	MaxLanes     int  `json:"max_lanes"` // 3, 6, 8, 10 lines
	TournamentID uint `json:"tournament_id"`
	Force        bool `json:"force"` // Force regenerate even if locked
}

type RecordResultRequest struct {
	RaceResultTime string `json:"race_result_time"` // mm.ss.ms
	FinalTime      string `json:"final_time"`       // alias for frontend flexibility
	Status         string `json:"status"`           // OK, DQ, DNF, DNS
	Rank           int    `json:"rank"`
	Round          string `json:"round"`            // "preliminary" or "final"
}

type SwapHeatLineRequest struct {
	TargetHeat     int    `json:"target_heat"`
	TargetLine     int    `json:"target_line"`
	SwapIfOccupied bool   `json:"swap_if_occupied"`
	Round          string `json:"round"` // "preliminary" or "final"
}

type LockBukuAcaraRequest struct {
	IsLocked bool `json:"is_locked"`
}

type PublishBukuAcaraRequest struct {
	IsPublished bool `json:"is_published"`
}

type GenerateFinalRoundRequest struct {
	TournamentID uint   `json:"tournament_id"`
	MaxLanes     int    `json:"max_lanes"`
	QualifyMode  string `json:"qualify_mode"` // "heat_winners_and_fastest" or "top_fastest"
}

type StartingItemDTO struct {
	No             int    `json:"no"`
	Nama           string `json:"nama"`
	Gender         string `json:"jenis_kelamin"`
	TimeSeed       string `json:"time_seed"`
	NomorLomba     string `json:"nomor_lomba"`
	Club           string `json:"club"`
	PIC            string `json:"pic"`
	Kontak         string `json:"kontak"`
	Result         string `json:"result,omitempty"`
	Rank           int    `json:"rank,omitempty"`
	TournamentID   uint   `json:"tournament_id,omitempty"`
	TournamentName string `json:"tournament_name,omitempty"`
}

type BukuAcaraHeatItemDTO struct {
	RegistrationID    uint   `json:"registration_id,omitempty"`
	Heat              int    `json:"heat"`
	HeatLabel         string `json:"heat_label,omitempty"`
	HeatCategory      string `json:"heat_category,omitempty"`
	Line              int    `json:"line"`
	Nama              string `json:"nama"`
	Gender            string `json:"jenis_kelamin"`
	Club              string `json:"club"`
	TimeSeed          string `json:"time_seed"`
	Result            string `json:"result"`
	Rank              int    `json:"rank,omitempty"`
	IsEmpty           bool   `json:"is_empty"`
	IsFinalist        bool   `json:"is_finalist"`
	PreliminaryResult string `json:"preliminary_result,omitempty"`
	PreliminaryRank   int    `json:"preliminary_rank,omitempty"`
}

type BukuAcaraEventGroupDTO struct {
	EventID      uint                   `json:"event_id"`
	TournamentID uint                   `json:"tournament_id"`
	EventCode    int                    `json:"event_code"`
	EventName    string                 `json:"event_name"`
	Distance     string                 `json:"distance"`
	Stroke       string                 `json:"stroke"`
	Gender       string                 `json:"gender"`
	AgeGroup     string                 `json:"age_group"`
	HeatCategory string                 `json:"heat_category"`
	MaxLanes     int                    `json:"max_lanes"`
	Round        string                 `json:"round"`
	HasFinalists bool                   `json:"has_finalists"`
	Heats        []BukuAcaraHeatItemDTO `json:"heats"`
}

type SaveTournamentRequest struct {
	ID                    uint   `json:"id"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	Location              string `json:"location"`
	RegistrationStartDate string `json:"registration_start_date"`
	RegistrationEndDate   string `json:"registration_end_date"`
	EventStartDate        string `json:"event_start_date"`
	EventEndDate          string `json:"event_end_date"`
	IsActive              bool   `json:"is_active"`
	IsBukuAcaraLocked     bool   `json:"is_buku_acara_locked"`
	IsBukuAcaraPublished  bool   `json:"is_buku_acara_published"`
}
