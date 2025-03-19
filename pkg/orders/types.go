package orders

import (
	"time"
)

type MessageAndSuccess struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type APIError struct {
	Error string `json:"error"`
}

type Account struct {
	ID      int    `json:"ID"`
	Email   string `json:"Email"`
	UserKey string `json:"UserKCID"`

	FirstName *string `json:"FirstName"`
	LastName  *string `json:"LastName"`
	Phone     *string `json:"Phone"`
	Street    *string `json:"Street"`
	City      *string `json:"City"`
	State     *string `json:"State"`
	Postcode  *string `json:"Postcode"`
	Country   *string `json:"Country"`
}

type Order struct {
	ID           int       `json:"ID"`
	AccountID    int       `json:"AccountID"`
	Status       string    `json:"Status"`
	ProductType  string    `json:"ProductType"`
	PaymentDate  time.Time `json:"PaymentDate"`
	Type         string    `json:"type"`
	Quantity     int       `json:"Quantity"`
	StartingDate time.Time `json:"StartingDate"`
	Flag         string    `json:"Flag"`
}

type Payment struct {
	ID            int       `json:"ID"`
	OrderID       int       `json:"OrderID"`
	Amount        float64   `json:"Amount"`
	Currency      string    `json:"Currency"`
	CCNumber      string    `json:"CCNumber"`
	PaymentStatus string    `json:"PaymentStatus"`
	PaymentType   string    `json:"PaymentType"`
	CreatedAt     time.Time `json:"created_at"`
	Status        string    `json:"Status"`
}

type Special struct {
	Id          int       `json:"id" gorm:"primary_key"`
	KeycloakId  string    `json:"keycloak_id"`
	Email       string    `json:"email"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Category    string    `json:"category"`
	SubCategory string    `json:"subcategory"`
}

type Status struct {
	Membership  bool   `json:"membership"`
	Ticket      bool   `json:"ticket"`
	StatusName  string `json:"status_name"`
	StatusColor string `json:"status_color"`
	IsSpecial   bool   `json:"is_special"`
}

type AccountRes struct {
	MessageAndSuccess
	Data Account `json:"data"`
}

type SpecialRes struct {
	MessageAndSuccess
	Data []Special `json:"data"`
}

type PaymentRes struct {
	MessageAndSuccess
	Data Payment `json:"data"`
}

type MultiplePaymentRes struct {
	MessageAndSuccess
	Data []Payment `json:"data"`
}

type OrdersRes struct {
	MessageAndSuccess
	Data []Order `json:"data"`
}

type OrderRes struct {
	MessageAndSuccess
	Data Order `json:"data"`
}
