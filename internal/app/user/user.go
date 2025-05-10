package user

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"mime/multipart"
	"webapi/config"
	user "webapi/internal/dto"
	password "webapi/internal/helper/utils"
	"webapi/internal/http/requests"
	internal_minio "webapi/pkg/minio"

	"webapi/internal/db/model"
	"webapi/internal/repository"
	"webapi/pkg/exception"
)

type UserApp interface {
	Login(ctx context.Context, dti requests.AuthLoginRequest) (user.GetUserDTO, error)
	GetUsers(ctx context.Context) ([]user.GetUserDTO, error)
	GetUsersWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) (user.DataWithPaginationDTO, error)
	GetUserByID(ctx context.Context, input requests.GetUserIdRequest) (user.GetUserDTO, error)
	CreateUser(ctx context.Context, input requests.CreateUserRequest) (user.GetUserDTO, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input requests.UpdateUserRequest) (user.GetUserDTO, error)
	DeleteUser(ctx context.Context, input requests.GetUserIdRequest) (bool, error)
	ChangePassword(ctx context.Context, id uuid.UUID, input requests.ChangePasswordRequest) (user.GetUserDTO, error)
	ChangePhone(ctx context.Context, id uuid.UUID, input requests.ChangePhoneRequest) (user.GetUserDTO, error)
	ChangeEmail(ctx context.Context, id uuid.UUID, input requests.ChangeEmailRequest) (user.GetUserDTO, error)
	ChangeUserName(ctx context.Context, id uuid.UUID, input requests.ChangeUserNameRequest) (user.GetUserDTO, error)
	GetUserByUsername(ctx context.Context, input requests.GetUserNameRequest) (user.GetUserDTO, error)
	GetUserByEmail(ctx context.Context, input requests.ChangeEmailRequest) (user.GetUserDTO, error)
	GetUserByPhone(ctx context.Context, input requests.ChangePhoneRequest) (user.GetUserDTO, error)
	UploadAvatar(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (user.GetUserDTO, error)
}

type userApp struct {
	Repo *repository.Repository
}

func NewUserApp(repo *repository.Repository) UserApp {
	return &userApp{
		Repo: repo,
	}
}

func (s *userApp) Login(ctx context.Context, input requests.AuthLoginRequest) (user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByUsername(ctx, input.UserName)
	if err != nil {
		return user.GetUserDTO{}, err
	}
	fmt.Println(userRepo)
	authLogin := password.ComparePassword(userRepo.Password, input.Password)

	if userRepo.ID == uuid.Nil { // Check for zero-value UUID
		return user.GetUserDTO{}, exception.DataNotFoundError
	}
	if authLogin {
		return user.GetUserDTO{
			ID:        userRepo.ID,
			UserName:  userRepo.UserName,
			Email:     userRepo.Email,
			Phone:     userRepo.Phone,
			CreatedAt: userRepo.CreatedAt,
			UpdatedAt: userRepo.UpdatedAt,
		}, nil

	}
	return user.GetUserDTO{}, exception.InvalidCredentialsError

}

func (s *userApp) GetUsers(ctx context.Context) ([]user.GetUserDTO, error) {
	users, err := s.Repo.User.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	var usersDTO []user.GetUserDTO
	for _, userRepo := range users {
		usersDTO = append(usersDTO, user.GetUserDTO{
			ID:       userRepo.ID,
			UserName: userRepo.UserName,
			Email:    userRepo.Email,
			Phone:    userRepo.Phone,
		})
	}

	return usersDTO, nil
}

func (s *userApp) GetUsersWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) (user.DataWithPaginationDTO, error) {
	responseUser, err := s.Repo.User.GetUsersWithPagination(ctx, input)
	if err != nil {
		return user.DataWithPaginationDTO{}, err
	}

	return responseUser, nil
}

func (s *userApp) GetUserByID(ctx context.Context, input requests.GetUserIdRequest) (user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByID(ctx, input.ID)
	if err != nil {
		return user.GetUserDTO{}, err
	}
	if userRepo.ID == uuid.Nil { // Check for zero-value UUID
		return user.GetUserDTO{}, exception.DataNotFoundError
	}
	return user.GetUserDTO{
		ID:       userRepo.ID,
		UserName: userRepo.UserName,
		Email:    userRepo.Email,
		Phone:    userRepo.Phone,
	}, nil
}

func (s *userApp) CreateUser(ctx context.Context, input requests.CreateUserRequest) (user.GetUserDTO, error) {
	// Ensure email is not already taken
	isUserEmailExist, err := s.Repo.User.IsUserEmailExist(ctx, input.Email)
	if err != nil {
		return user.GetUserDTO{}, err
	}
	if isUserEmailExist {
		return user.GetUserDTO{}, exception.UserEmailAlreadyTakenError
	}
	isUserPhoneExist, err := s.Repo.User.IsUserPhoneExist(ctx, input.Phone)
	if err != nil {
		return user.GetUserDTO{}, err
	}
	if isUserPhoneExist {
		return user.GetUserDTO{}, exception.UserPhoneAlreadyTakenError
	}

	userRepo, err := s.Repo.User.AddUser(ctx, model.User{
		UserName: input.UserName,
		Email:    input.Email,
		Phone:    input.Phone,
		Password: password.GeneratePassword(input.Password),
	})
	if err != nil {
		return user.GetUserDTO{}, err
	}

	return user.GetUserDTO{
		ID:        userRepo.ID,
		UserName:  userRepo.UserName,
		Email:     userRepo.Email,
		Phone:     userRepo.Phone,
		CreatedAt: userRepo.CreatedAt,
		UpdatedAt: userRepo.UpdatedAt,
	}, nil
}

func (s *userApp) UpdateUser(ctx context.Context, id uuid.UUID, input requests.UpdateUserRequest) (user.GetUserDTO, error) {

	userRepo, err := s.Repo.User.GetUserByID(ctx, id)
	if err != nil {
		return user.GetUserDTO{}, err
	}

	userRepo, err = s.Repo.User.UpdateUser(ctx, model.User{
		ID:       userRepo.ID,
		UserName: input.UserName,
		Email:    input.Email,
		Phone:    input.Phone,
	})
	if err != nil {
		return user.GetUserDTO{}, err
	}

	if userRepo.ID != uuid.Nil { // Check if the user was successfully updated
		return user.GetUserDTO{
			ID:        userRepo.ID,
			UserName:  userRepo.UserName,
			Email:     userRepo.Email,
			Phone:     userRepo.Phone,
			CreatedAt: userRepo.CreatedAt,
			UpdatedAt: userRepo.UpdatedAt,
		}, nil
	}

	return user.GetUserDTO{}, nil
}

func (s *userApp) DeleteUser(ctx context.Context, input requests.GetUserIdRequest) (bool, error) {
	userRepo, err := s.Repo.User.GetUserByID(ctx, input.ID)
	if err != nil {
		return false, err
	}
	deleteUser, err := s.Repo.User.DeleteUser(ctx, userRepo.ID)
	if err != nil {
		return false, err
	}

	return deleteUser, nil
}
func (s *userApp) ChangePassword(ctx context.Context, id uuid.UUID, input requests.ChangePasswordRequest) (user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByID(ctx, id)
	if err != nil {
		return user.GetUserDTO{}, err
	}
	authLogin := password.ComparePassword(userRepo.Password, input.OldPassword)
	if !authLogin {
		return user.GetUserDTO{}, exception.InvalidCredentialsError
	}
	userRepo, err = s.Repo.User.UpdateUser(ctx, model.User{
		ID:       userRepo.ID,
		UserName: userRepo.UserName,
		Email:    userRepo.Email,
		Phone:    userRepo.Phone,
		Password: password.GeneratePassword(input.NewPassword),
	})
	if err != nil {
		return user.GetUserDTO{}, err
	}

	return user.GetUserDTO{
		ID:        userRepo.ID,
		UserName:  userRepo.UserName,
		Email:     userRepo.Email,
		Phone:     userRepo.Phone,
		CreatedAt: userRepo.CreatedAt,
		UpdatedAt: userRepo.UpdatedAt,
	}, nil
}
func (s *userApp) ChangePhone(ctx context.Context, id uuid.UUID, input requests.ChangePhoneRequest) (user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.UpdateUser(ctx, model.User{
		ID:    id,
		Phone: input.Phone,
	})
	if err != nil {
		return user.GetUserDTO{}, err
	}

	return user.GetUserDTO{
		ID:        userRepo.ID,
		UserName:  userRepo.UserName,
		Email:     userRepo.Email,
		Phone:     userRepo.Phone,
		CreatedAt: userRepo.CreatedAt,
		UpdatedAt: userRepo.UpdatedAt,
	}, nil
}
func (s *userApp) ChangeEmail(ctx context.Context, id uuid.UUID, input requests.ChangeEmailRequest) (user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.UpdateUser(ctx, model.User{
		ID:    id,
		Email: input.Email,
	})
	if err != nil {
		return user.GetUserDTO{}, err
	}

	return user.GetUserDTO{
		ID:        userRepo.ID,
		UserName:  userRepo.UserName,
		Email:     userRepo.Email,
		Phone:     userRepo.Phone,
		CreatedAt: userRepo.CreatedAt,
		UpdatedAt: userRepo.UpdatedAt,
	}, nil
}
func (s *userApp) ChangeUserName(ctx context.Context, id uuid.UUID, input requests.ChangeUserNameRequest) (user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.UpdateUser(ctx, model.User{
		ID:       id,
		UserName: input.UserName,
	})
	if err != nil {
		return user.GetUserDTO{}, err
	}

	return user.GetUserDTO{
		ID:        userRepo.ID,
		UserName:  userRepo.UserName,
		Email:     userRepo.Email,
		Phone:     userRepo.Phone,
		CreatedAt: userRepo.CreatedAt,
		UpdatedAt: userRepo.UpdatedAt,
	}, nil
}
func (s *userApp) GetUserByUsername(ctx context.Context, input requests.GetUserNameRequest) (user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByUsername(ctx, input.UserName)
	if err != nil {
		return user.GetUserDTO{}, err
	}
	if userRepo.ID == uuid.Nil { // Check for zero-value UUID
		return user.GetUserDTO{}, exception.DataNotFoundError
	}
	return user.GetUserDTO{
		ID:       userRepo.ID,
		UserName: userRepo.UserName,
		Email:    userRepo.Email,
		Phone:    userRepo.Phone,
	}, nil
}

func (s *userApp) GetUserByEmail(ctx context.Context, input requests.ChangeEmailRequest) (user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return user.GetUserDTO{}, err
	}
	if userRepo.ID == uuid.Nil { // Check for zero-value UUID
		return user.GetUserDTO{}, exception.DataNotFoundError
	}
	return user.GetUserDTO{
		ID:       userRepo.ID,
		UserName: userRepo.UserName,
		Email:    userRepo.Email,
		Phone:    userRepo.Phone,
	}, nil
}

func (s *userApp) GetUserByPhone(ctx context.Context, input requests.ChangePhoneRequest) (user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByPhone(ctx, input.Phone)
	if err != nil {
		return user.GetUserDTO{}, err
	}
	if userRepo.ID == uuid.Nil { // Check for zero-value UUID
		return user.GetUserDTO{}, exception.DataNotFoundError
	}
	return user.GetUserDTO{
		ID:       userRepo.ID,
		UserName: userRepo.UserName,
		Email:    userRepo.Email,
		Phone:    userRepo.Phone,
	}, nil
}

func (s *userApp) UploadAvatar(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (user.GetUserDTO, error) {
	fileContent, err := file.Open()
	if err != nil {
		return user.GetUserDTO{}, err
	}
	defer fileContent.Close()

	conf := config.GetConfig().Minio
	objectName := file.Filename
	bucketName := conf.BucketName
	contentType := file.Header.Get("Content-Type")

	userRepo, err := s.GetUserByID(ctx, requests.GetUserIdRequest{ID: id})

	_, err = s.Repo.Media.CreateMedia(ctx, model.Media{
		Name:     objectName,
		FileName: file.Filename,
		Size:     file.Size,
		MimeType: contentType,
	})

	minioClient := internal_minio.GetMinio()
	if _, err = minioClient.PutObject(context.Background(), bucketName, objectName, fileContent, file.Size, minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return user.GetUserDTO{}, err
	}

	return userRepo, nil
}
