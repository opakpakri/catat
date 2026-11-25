package models

import "time"

type User struct {
	ID           int       `json:"id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	IsPINEnabled bool      `json:"is_pin_enabled"`
	Currency     string    `json:"currency"`
	CreatedAt    time.Time `json:"created_at"`
}

type Wallet struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Balance   float64   `json:"balance"`
	ColorHex  string    `json:"color_hex"`
	CreatedAt time.Time `json:"created_at"`
}

type Category struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Icon      string    `json:"icon"`
	ColorHex  string    `json:"color_hex"`
	CreatedAt time.Time `json:"created_at"`
}

type Transaction struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	WalletID         int       `json:"wallet_id"`
	WalletName       string    `json:"wallet_name,omitempty"`
	TargetWalletID   *int      `json:"target_wallet_id,omitempty"`
	TargetWalletName string    `json:"target_wallet_name,omitempty"`
	CategoryID       *int      `json:"category_id,omitempty"`
	CategoryName     string    `json:"category_name,omitempty"`
	CategoryIcon     string    `json:"category_icon,omitempty"`
	CategoryColor    string    `json:"category_color,omitempty"`
	Type             string    `json:"type"`
	Amount           float64   `json:"amount"`
	TransactionDate  time.Time `json:"transaction_date"`
	Note             string    `json:"note"`
	CreatedAt        time.Time `json:"created_at"`
}

type Budget struct {
	ID            int     `json:"id"`
	UserID        int     `json:"user_id"`
	CategoryID    int     `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	CategoryIcon  string  `json:"category_icon"`
	CategoryColor string  `json:"category_color"`
	AmountLimit   float64 `json:"amount_limit"`
	SpentAmount   float64 `json:"spent_amount"`
	PeriodMonth   int     `json:"period_month"`
	PeriodYear    int     `json:"period_year"`
}

type Debt struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	WalletID    *int      `json:"wallet_id,omitempty"`
	WalletName  string    `json:"wallet_name,omitempty"`
	PersonName  string    `json:"person_name"`
	Type        string    `json:"type"`
	TotalAmount float64   `json:"total_amount"`
	PaidAmount  float64   `json:"paid_amount"`
	DueDate     *string   `json:"due_date,omitempty"`
	Status      string    `json:"status"`
	Note        string    `json:"note"`
	CreatedAt   time.Time `json:"created_at"`
}

type CalendarEvent struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	DebtID      *int      `json:"debt_id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventDate   string    `json:"event_date"`
	EventTime   *string   `json:"event_time,omitempty"`
	IsCompleted bool      `json:"is_completed"`
	ColorHex    string    `json:"color_hex"`
	CreatedAt   time.Time `json:"created_at"`
}
