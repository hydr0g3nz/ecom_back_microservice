package grpcctl

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/controller/grpc/proto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// UserServer implements the gRPC UserService interface
type UserServer struct {
	pb.UnimplementedUserServiceServer
	authUsecase usecase.AuthUsecase
	userUsecase usecase.UserUsecase
	logger      logger.Logger
}

// NewUserServer creates a new UserServer instance
func NewUserServer(authUsecase usecase.AuthUsecase, userUsecase usecase.UserUsecase, logger logger.Logger) *UserServer {
	return &UserServer{
		authUsecase: authUsecase,
		userUsecase: userUsecase,
		logger:      logger,
	}
}

// CreateUser creates a new user
func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	s.logger.Info("gRPC CreateUser request received", "email", req.Email)

	user := entity.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	createdUser, err := s.userUsecase.CreateUser(ctx, &user, req.Password)
	if err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, handleError(err)
	}

	return convertUserToProto(createdUser), nil
}

// GetUser gets a user by ID
func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	s.logger.Info("gRPC GetUser request received", "id", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	user, err := s.userUsecase.GetUserByID(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err)
		return nil, handleError(err)
	}

	return convertUserToProto(user), nil
}

// UpdateUser updates an existing user
func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	s.logger.Info("gRPC UpdateUser request received", "id", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	user := entity.User{
		ID:        req.Id,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	updatedUser, err := s.userUsecase.UpdateUser(ctx, req.Id, user)
	if err != nil {
		s.logger.Error("Failed to update user", "error", err)
		return nil, handleError(err)
	}

	return convertUserToProto(updatedUser), nil
}

// DeleteUser deletes a user by ID
func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	s.logger.Info("gRPC DeleteUser request received", "id", req.Id)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	err := s.userUsecase.DeleteUser(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to delete user", "error", err)
		return nil, handleError(err)
	}

	return &pb.DeleteUserResponse{Success: true}, nil
}

// Login authenticates a user and returns tokens
func (s *UserServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	s.logger.Info("gRPC Login request received", "email", req.Email)

	tokenPair, err := s.authUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		s.logger.Error("Failed to login", "error", err)
		return nil, handleError(err)
	}

	return &pb.LoginResponse{
		TokenPair: &pb.TokenPairResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
		},
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *UserServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.TokenPairResponse, error) {
	s.logger.Info("gRPC RefreshToken request received")

	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	tokenPair, err := s.authUsecase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Error("Failed to refresh token", "error", err)
		return nil, handleError(err)
	}

	return &pb.TokenPairResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// Helper function to convert domain user entity to protobuf user response
func convertUserToProto(user *entity.User) *pb.UserResponse {
	return &pb.UserResponse{
		Id:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

// handleError maps domain errors to appropriate gRPC status errors
func handleError(err error) error {
	var statusCode codes.Code
	var message string

	switch {
	case errors.Is(err, entity.ErrUserNotFound):
		statusCode = codes.NotFound
		message = "User not found"
	case errors.Is(err, entity.ErrUserAlreadyExists) || errors.Is(err, entity.ErrUserExists) || errors.Is(err, entity.ErrUserAlreadyExists):
		statusCode = codes.AlreadyExists
		message = "User already exists"
	case errors.Is(err, entity.ErrInvalidCredentials):
		statusCode = codes.Unauthenticated
		message = "Invalid credentials"
	case errors.Is(err, entity.ErrInvalidToken) || errors.Is(err, entity.ErrTokenHasBeenRevoked) || errors.Is(err, entity.ErrInvalidToken):
		statusCode = codes.Unauthenticated
		message = "Invalid or revoked token"
	case errors.Is(err, entity.ErrInternalServerError):
		statusCode = codes.Internal
		message = "Internal server error"
	default:
		statusCode = codes.Internal
		message = "Something went wrong"
	}

	return status.Error(statusCode, message)
}
