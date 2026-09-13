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
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
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
	MaxLanes int `json:"max_lanes"` // 3, 6, 8, 10 lines
}

type RecordResultRequest struct {
	RaceResultTime string `json:"race_result_time"` // mm.ss.ms
	Rank           int    `json:"rank"`
}

type StartingItemDTO struct {
	No        int    `json:"no"`
	Nama      string `json:"nama"`
	Gender    string `json:"jenis_kelamin"`
	TimeSeed  string `json:"time_seed"`
	NomorLomba string `json:"nomor_lomba"`
	Club      string `json:"club"`
	PIC       string `json:"pic"`
	Kontak    string `json:"kontak"`
}

type BukuAcaraHeatItemDTO struct {
	Heat     int    `json:"heat"`
	Line     int    `json:"line"`
	Nama     string `json:"nama"`
	Gender   string `json:"jenis_kelamin"`
	Club     string `json:"club"`
	TimeSeed string `json:"time_seed"`
	Result   string `json:"result"`
}

type BukuAcaraEventGroupDTO struct {
	EventCode int                    `json:"event_code"`
	EventName string                 `json:"event_name"`
	Gender    string                 `json:"gender"`
	Heats     []BukuAcaraHeatItemDTO `json:"heats"`
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
}
