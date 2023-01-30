package models

type UserStore interface {
	GetAll(exclude []string, searchInput string, page, pageSize int, descending bool) ([]User, error)
	Count() (int64, error)
	GetById(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(user *User) error
	DeleteById(id string) error
	DeleteByEmail(email string) error

	GetCashLog(user *User, searchInput string, page, pageSize int, oldestFirst bool) ([]CashLogEntry, error)
	CashLogEntryCount(user *User) (int64, error)
	GetLastCashLogEntry(user *User) (*CashLogEntry, error)
	GetCashLogEntryById(user *User, id string) (*CashLogEntry, error)
	AddCashLogEntry(user *User, entry *CashLogEntry) error
}

type User struct {
	Base
	Name                    string
	Email                   string `gorm:"unique"`
	PubliclyVisible         bool   `gorm:"default:true"`
	DontSendInvitationEmail bool
	CashLog                 []CashLogEntry
	GroupMemberships        []GroupMembership
	GroupInvitations        []GroupInvitation
}

type CashLogEntry struct {
	Base
	ChangeTitle       string
	ChangeDescription string
	TotalAmount       int
	ChangeDifference  int

	Ct1  int
	Ct2  int
	Ct5  int
	Ct10 int
	Ct20 int
	Ct50 int

	Eur1   int
	Eur2   int
	Eur5   int
	Eur10  int
	Eur20  int
	Eur50  int
	Eur100 int
	Eur200 int
	Eur500 int

	UserId string
}
