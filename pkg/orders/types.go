package orders

import "time"

type MessageAndSuccess struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type APIError struct {
	Error string `json:"error"`
}

type Order struct {
	ID           int       `json:"ID"`
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
	Amount        int       `json:"Amount"`
	DebitCurrency string    `json:"DebitCurrency"`
	CCNumber      string    `json:"CCNumber"`
	PaymentStatus string    `json:"PaymentStatus"`
	PaymentType   string    `json:"PaymentType"`
	CreatedAt     time.Time `json:"created_at"`
	Status        string    `json:"Status"`
}

type Special struct {
	Email       string `json:"email"`
	Category    string `json:"category"`
	SubCategory string `json:"subcategory"`
}

type Status struct {
	Membership  bool   `json:"membership"`
	Ticket      bool   `json:"ticket"`
	StatusName  string `json:"status_name"`
	StatusColor string `json:"status_color"`
	IsSpecial   bool   `json:"is_special"`
}

type SpecialRes struct {
	MessageAndSuccess
	Data Special `json:"data"`
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
