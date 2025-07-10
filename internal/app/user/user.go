package user

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"webapi/config"
	user "webapi/internal/dto"
	"webapi/internal/helper/utils"
	"webapi/internal/http/requests"
	internal_minio "webapi/pkg/minio"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"webapi/internal/db/model"
	"webapi/internal/repository"
	"webapi/pkg/exception"
)

type UserApp interface {
	Login(ctx context.Context, dti requests.AuthLoginRequest) (*user.GetUserDTO, error)
	GetUsers(ctx context.Context) ([]*user.GetUserDTO, error)
	GetUsersWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) (*user.DataWithPaginationDTO, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.GetUserDTO, error)
	CreateUser(ctx context.Context, input model.User) (*user.GetUserDTO, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input model.User) (*user.GetUserDTO, error)
	DeleteUser(ctx context.Context, id uuid.UUID) (bool, error)
	ChangePassword(ctx context.Context, id uuid.UUID, input requests.ChangePasswordRequest) (*user.GetUserDTO, error)
	ChangePhone(ctx context.Context, id uuid.UUID, phone string) (*user.GetUserDTO, error)
	ChangeEmail(ctx context.Context, id uuid.UUID, email string) (*user.GetUserDTO, error)
	ChangeUserName(ctx context.Context, id uuid.UUID, username string) (*user.GetUserDTO, error)
	GetUserByUsername(ctx context.Context, username string) (*user.GetUserDTO, error)
	GetUserByEmail(ctx context.Context, email string) (*user.GetUserDTO, error)
	GetUserByPhone(ctx context.Context, phone string) (*user.GetUserDTO, error)
	UploadAvatar(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (minio.UploadInfo, error)
}

type userApp struct {
	Repo *repository.Repository
}

func NewUserApp(repo *repository.Repository) UserApp {
	return &userApp{
		Repo: repo,
	}
}

func (s *userApp) Login(ctx context.Context, input requests.AuthLoginRequest) (*user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByUsername(ctx, input.UserName)
	if err != nil {
		return nil, err
	}
	authLogin := utils.ComparePassword(userRepo.Password, input.Password)

	if userRepo.ID == uuid.Nil { // Check for zero-value UUID
		return nil, exception.DataNotFoundError
	}
	if authLogin {
		return &user.GetUserDTO{
			ID:        userRepo.ID,
			UserName:  userRepo.UserName,
			Email:     userRepo.Email,
			Phone:     userRepo.Phone,
			CreatedAt: userRepo.CreatedAt,
			UpdatedAt: userRepo.UpdatedAt,
		}, nil

	}
	return nil, exception.InvalidCredentialsError

}

func (s *userApp) GetUsers(ctx context.Context) ([]*user.GetUserDTO, error) {
	users, err := s.Repo.User.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	var usersDTO []*user.GetUserDTO
	for _, userRepo := range users {
		usersDTO = append(usersDTO, &user.GetUserDTO{
			ID:       userRepo.ID,
			UserName: userRepo.UserName,
			Email:    userRepo.Email,
			Phone:    userRepo.Phone,
		})
	}

	return usersDTO, nil
}

func (s *userApp) GetUsersWithPagination(ctx context.Context, input requests.DataWithPaginationRequest) (*user.DataWithPaginationDTO, error) {
	responseUser, err := s.Repo.User.GetUsersWithPagination(ctx, input)
	if err != nil {
		return nil, err
	}

	return &responseUser, nil
}

func (s *userApp) GetUserByID(ctx context.Context, id uuid.UUID) (*user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if userRepo.ID == uuid.Nil { // Check for zero-value UUID
		return nil, exception.DataNotFoundError
	}
	return &user.GetUserDTO{
		ID:       userRepo.ID,
		UserName: userRepo.UserName,
		Email:    userRepo.Email,
		Phone:    userRepo.Phone,
	}, nil
}

func (s *userApp) CreateUser(ctx context.Context, input model.User) (*user.GetUserDTO, error) {
	// Ensure email is not already taken
	isUserEmailExist, err := s.Repo.User.IsUserEmailExist(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if isUserEmailExist {
		return nil, exception.UserEmailAlreadyTakenError
	}
	isUserPhoneExist, err := s.Repo.User.IsUserPhoneExist(ctx, input.Phone)
	if err != nil {
		return nil, err
	}
	if isUserPhoneExist {
		return nil, exception.UserPhoneAlreadyTakenError
	}
	phone, err := utils.FormatPhoneToInternational(input.Phone, "ID")
	if err != nil {
		return nil, exception.UserPhoneAlreadyTakenError
	}

	userRepo, err := s.Repo.User.AddUser(ctx, model.User{
		UserName: input.UserName,
		Email:    input.Email,
		Phone:    phone,
		Password: utils.GeneratePassword(input.Password),
	})
	if err != nil {
		return nil, err
	}

	return &user.GetUserDTO{
		ID:        userRepo.ID,
		UserName:  userRepo.UserName,
		Email:     userRepo.Email,
		Phone:     userRepo.Phone,
		CreatedAt: userRepo.CreatedAt,
		UpdatedAt: userRepo.UpdatedAt,
	}, nil
}

func (s *userApp) UpdateUser(ctx context.Context, id uuid.UUID, input model.User) (*user.GetUserDTO, error) {

	userRepo, err := s.Repo.User.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	userRepo, err = s.Repo.User.UpdateUser(ctx, model.User{
		ID:       userRepo.ID,
		UserName: input.UserName,
		Email:    input.Email,
		Phone:    input.Phone,
	})
	if err != nil {
		return nil, err
	}

	if userRepo.ID != uuid.Nil { // Check if the user was successfully updated
		return &user.GetUserDTO{
			ID:        userRepo.ID,
			UserName:  userRepo.UserName,
			Email:     userRepo.Email,
			Phone:     userRepo.Phone,
			CreatedAt: userRepo.CreatedAt,
			UpdatedAt: userRepo.UpdatedAt,
		}, nil
	}

	return nil, nil
}

func (s *userApp) DeleteUser(ctx context.Context, id uuid.UUID) (bool, error) {
	userRepo, err := s.Repo.User.GetUserByID(ctx, id)
	if err != nil {
		return false, err
	}
	deleteUser, err := s.Repo.User.DeleteUser(ctx, userRepo.ID)
	if err != nil {
		return false, err
	}

	return deleteUser, nil
}
func (s *userApp) ChangePassword(ctx context.Context, id uuid.UUID, input requests.ChangePasswordRequest) (*user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	authLogin := utils.ComparePassword(userRepo.Password, input.OldPassword)
	if !authLogin {
		return nil, exception.InvalidCredentialsError
	}
	userRepo, err = s.Repo.User.UpdateUser(ctx, model.User{
		ID:       userRepo.ID,
		UserName: userRepo.UserName,
		Email:    userRepo.Email,
		Phone:    userRepo.Phone,
		Password: utils.GeneratePassword(input.NewPassword),
	})
	if err != nil {
		return nil, err
	}

	return &user.GetUserDTO{
		ID:        userRepo.ID,
		UserName:  userRepo.UserName,
		Email:     userRepo.Email,
		Phone:     userRepo.Phone,
		CreatedAt: userRepo.CreatedAt,
		UpdatedAt: userRepo.UpdatedAt,
	}, nil
}
func (s *userApp) ChangePhone(ctx context.Context, id uuid.UUID, phone string) (*user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.UpdateUser(ctx, model.User{
		ID:    id,
		Phone: phone,
	})
	if err != nil {
		return nil, err
	}

	return &user.GetUserDTO{
		ID:        userRepo.ID,
		UserName:  userRepo.UserName,
		Email:     userRepo.Email,
		Phone:     userRepo.Phone,
		CreatedAt: userRepo.CreatedAt,
		UpdatedAt: userRepo.UpdatedAt,
	}, nil
}
func (s *userApp) ChangeEmail(ctx context.Context, id uuid.UUID, email string) (*user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.UpdateUser(ctx, model.User{
		ID:    id,
		Email: email,
	})
	if err != nil {
		return nil, err
	}

	return &user.GetUserDTO{
		ID:        userRepo.ID,
		UserName:  userRepo.UserName,
		Email:     userRepo.Email,
		Phone:     userRepo.Phone,
		CreatedAt: userRepo.CreatedAt,
		UpdatedAt: userRepo.UpdatedAt,
	}, nil
}
func (s *userApp) ChangeUserName(ctx context.Context, id uuid.UUID, username string) (*user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.UpdateUser(ctx, model.User{
		ID:       id,
		UserName: username,
	})
	if err != nil {
		return nil, err
	}

	return &user.GetUserDTO{
		ID:        userRepo.ID,
		UserName:  userRepo.UserName,
		Email:     userRepo.Email,
		Phone:     userRepo.Phone,
		CreatedAt: userRepo.CreatedAt,
		UpdatedAt: userRepo.UpdatedAt,
	}, nil
}
func (s *userApp) GetUserByUsername(ctx context.Context, username string) (*user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if userRepo.ID == uuid.Nil { // Check for zero-value UUID
		return nil, exception.DataNotFoundError
	}
	return &user.GetUserDTO{
		ID:       userRepo.ID,
		UserName: userRepo.UserName,
		Email:    userRepo.Email,
		Phone:    userRepo.Phone,
	}, nil
}

func (s *userApp) GetUserByEmail(ctx context.Context, email string) (*user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if userRepo.ID == uuid.Nil { // Check for zero-value UUID
		return nil, exception.DataNotFoundError
	}
	return &user.GetUserDTO{
		ID:       userRepo.ID,
		UserName: userRepo.UserName,
		Email:    userRepo.Email,
		Phone:    userRepo.Phone,
	}, nil
}

func (s *userApp) GetUserByPhone(ctx context.Context, phone string) (*user.GetUserDTO, error) {
	userRepo, err := s.Repo.User.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if userRepo.ID == uuid.Nil { // Check for zero-value UUID
		return nil, exception.DataNotFoundError
	}
	return &user.GetUserDTO{
		ID:       userRepo.ID,
		UserName: userRepo.UserName,
		Email:    userRepo.Email,
		Phone:    userRepo.Phone,
	}, nil
}

func (s *userApp) UploadAvatar(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (minio.UploadInfo, error) {
	fileContent, err := file.Open()
	if err != nil {
		return minio.UploadInfo{}, err
	}
	defer func(fileContent multipart.File) {
		err := fileContent.Close()
		if err != nil {

		}
	}(fileContent)

	conf := config.GetConfig().Minio
	objectName := file.Filename
	bucketName := conf.BucketName
	contentType := file.Header.Get("Content-Type")

	userRepo, err := s.GetUserByID(ctx, id)

	minioClient := internal_minio.GetMinio()
	uploadInfo, err := minioClient.PutObject(context.Background(), bucketName, objectName, fileContent, file.Size, minio.PutObjectOptions{ContentType: contentType})

	if err != nil {
		return minio.UploadInfo{}, err
	}
	atribute, _ := json.Marshal(uploadInfo)
	media, err := s.Repo.Media.CreateMedia(ctx, model.Media{
		Name:             objectName,
		FileName:         file.Filename,
		Size:             file.Size,
		MimeType:         contentType,
		Hash:             uploadInfo.ChecksumSHA256,
		CustomAttributes: string(atribute),
	})
	if err != nil {
		return minio.UploadInfo{}, err
	}
	err = s.Repo.Media.AttachMedia(ctx, user.MediaRelation{
		MediaID:      media.ID,
		MediableType: "user",
		MediableId:   userRepo.ID,
		Group:        "avatar",
	})

	return uploadInfo, nil
}
