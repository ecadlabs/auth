package users

func (u *User) CanUpdateUser(target *User) bool {
	var r Role
	if u != nil {
		r = u.Role
	} else {
		r = RoleRegular
	}

	return r >= target.Role
}

func (u *User) CanCreateUser(target *User) bool {
	return u.CanUpdateUser(target)
}

func (u *User) CanViewUser(target *User) bool {
	return u.Role >= target.Role
}

func (u *User) CanListUsers() bool {
	return u.Role == RoleAdmin
}
