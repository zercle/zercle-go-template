package usecase

import (
	"github.com/zercle/zercle-go-template/internal/feature/user"
)

// domainToDTO converts a domain User entity to a UserDTO.
func domainToDTO(u *user.User) *user.UserDTO {
	if u == nil {
		return nil
	}
	return &user.UserDTO{
		ID:        string(u.ID),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// createDTOToDomain converts a CreateUserDTO to a domain User entity.
// Note: This only creates the domain entity structure, password hashing
// should be done by the caller before persisting.
func createDTOToDomain(dto *user.CreateUserDTO) (*user.User, error) {
	if dto == nil {
		return nil, nil
	}
	return user.NewUser(dto.Email, dto.Password, dto.FirstName, dto.LastName)
}

// listParamsToRepo converts ListParamsDTO to repository ListParams.
func listParamsToRepo(dto user.ListParamsDTO) user.ListParams {
	params := user.ListParams{
		Email:  dto.Email,
		Limit:  dto.Limit,
		Offset: dto.Offset,
	}

	if dto.Status != nil {
		status := user.UserStatus(*dto.Status)
		params.Status = &status
	}

	return params
}

// listResultToDTO converts a repository ListResult to a ListResultDTO.
func listResultToDTO(result *user.ListResult) *user.ListResultDTO {
	if result == nil {
		return nil
	}

	users := make([]*user.UserDTO, 0, len(result.Users))
	for _, u := range result.Users {
		users = append(users, domainToDTO(u))
	}

	return &user.ListResultDTO{
		Users:  users,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}
}
