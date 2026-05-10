package handlers

import (
	"fmt"

	"github.com/farildzaky/user-service/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// =============================================================================
// UUID helpers
// =============================================================================

func uuidToString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		id.Bytes[0:4], id.Bytes[4:6], id.Bytes[6:8], id.Bytes[8:10], id.Bytes[10:16])
}

func stringToUUID(s string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}

// =============================================================================
// Row → DTO converters
// =============================================================================

func toUserProfile(row db.CreateUserRow) UserProfile {
	p := UserProfile{
		ID:        uuidToString(row.ID),
		FirstName: row.FirstName,
		LastName:  row.LastName,
		Phone:     row.Phone,
		Email:     row.Email,
		NIK:       row.Nik,
		Role:      string(row.Role),
		IsActive:  row.IsActive,
	}
	if row.Address.Valid {
		p.Address = row.Address.String
	}
	if row.BirthDate.Valid {
		p.BirthDate = row.BirthDate.Time.Format("01/02/2006")
	}
	if row.Gender.Valid {
		p.Gender = string(row.Gender.GenderType)
	}
	return p
}

func toUserProfileFromUser(u db.User) UserProfile {
	p := UserProfile{
		ID:        uuidToString(u.ID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		Email:     u.Email,
		NIK:       u.Nik,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
	}
	if u.Address.Valid {
		p.Address = u.Address.String
	}
	if u.BirthDate.Valid {
		p.BirthDate = u.BirthDate.Time.Format("01/02/2006")
	}
	if u.Gender.Valid {
		p.Gender = string(u.Gender.GenderType)
	}
	return p
}

func toUserProfileFromAuthRow(u db.GetUserByEmailForAuthRow) UserProfile {
	p := UserProfile{
		ID:        uuidToString(u.ID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		Email:     u.Email,
		NIK:       u.Nik,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
	}
	if u.Address.Valid {
		p.Address = u.Address.String
	}
	if u.BirthDate.Valid {
		p.BirthDate = u.BirthDate.Time.Format("01/02/2006")
	}
	if u.Gender.Valid {
		p.Gender = string(u.Gender.GenderType)
	}
	return p
}

func toUserProfileFromIDRow(u db.GetUserByIDRow) UserProfile {
	p := UserProfile{
		ID:        uuidToString(u.ID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		Email:     u.Email,
		NIK:       u.Nik,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
	}
	if u.Address.Valid {
		p.Address = u.Address.String
	}
	if u.BirthDate.Valid {
		p.BirthDate = u.BirthDate.Time.Format("01/02/2006")
	}
	if u.Gender.Valid {
		p.Gender = string(u.Gender.GenderType)
	}
	return p
}

func toCreateRowFromUpdateRow(row db.UpdateUserProfileRow) db.CreateUserRow {
	return db.CreateUserRow{
		ID:        row.ID,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		Phone:     row.Phone,
		Email:     row.Email,
		Nik:       row.Nik,
		Address:   row.Address,
		BirthDate: row.BirthDate,
		Gender:    row.Gender,
		Role:      row.Role,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
