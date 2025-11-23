package user

type UserDTO struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func ToUserDTO(u *User) *UserDTO {
	return &UserDTO{
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}
