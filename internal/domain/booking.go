package domain

import (
	"time"
)

type Booking struct {
	Role           int32           `json:"role_id"`
	ClientID       int64           `json:"client_id"`
	BookingType    int32           `json:"booking_type"`
	StartTime      time.Time       `json:"start_time"`
	EndTime        time.Time       `json:"end_time"`
	TrainerID      *int64          `json:"trainer_id,omitempty"`
	LessonID       *int64          `json:"lesson_id,omitempty"`
	EquipmentItems []EquipmentItem `json:"equipment_items"`
	Status         int32           `json:"status"`
	Amount         int64           `json:"amount"`
	Currency       int64           `json:"currency"`
	CreatedAt      time.Time       `json:"created_at"`
}

type EquipmentItem struct {
	EquipmentID int64 `json:"equipment_id"`
	UnitID      int32 `json:"unit_id"`
}
