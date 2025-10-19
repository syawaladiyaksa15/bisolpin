package domain

type User struct {
	ID        uint64  `json:"id"`
	TutorID   *uint64 `json:"tutor_id"`
	PesertaID *uint64 `json:"peserta_id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Password  string  `json:"-"`
	Role      string  `json:"role"`
	IsActive  int     `json:"is_active"`
}
