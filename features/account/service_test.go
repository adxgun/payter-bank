package account

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"payter-bank/internal/database/models"
	databasemocks "payter-bank/internal/database/models/mocks"
	platformerrors "payter-bank/internal/errors"
	"payter-bank/pkg/generator"
	generatormocks "payter-bank/pkg/generator/mocks"
	"testing"
)

func TestService_InitialiseAdmin(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	email := "admin@example.com"
	password := "securepassword"
	ctx := context.Background()

	tests := []struct {
		name        string
		setupMocks  func(db *databasemocks.MockQuerier, mockGenerator *generatormocks.MockNumberGenerator)
		expectedErr error
	}{
		{
			name: "success - admin user created",
			setupMocks: func(mockDB *databasemocks.MockQuerier, mockGenerator *generatormocks.MockNumberGenerator) {
				mockDB.EXPECT().GetUserByEmail(gomock.Any(), email).Return(models.User{}, sql.ErrNoRows)
				mockDB.EXPECT().SaveUser(gomock.Any(), models.SaveUserParams{Email: email, Password: password}).
					Return(&models.User{ID: userID, Email: email}, nil)
				mockGenerator.EXPECT().Generate().Return("ADMIN123")
				mockDB.EXPECT().SaveAccount(gomock.Any(), models.SaveAccountParams{
					UserID:        userID,
					FirstName:     "Admin",
					LastName:      "Admin",
					AccountType:   models.AccountTypeADMIN,
					Status:        models.StatusACTIVE,
					AccountNumber: "ADMIN123",
				}).Return(&models.Account{}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "failure - GetUserByEmail returns non-NoRows error",
			setupMocks: func(mockDB *databasemocks.MockQuerier, mockGenerator *generatormocks.MockNumberGenerator) {
				mockDB.EXPECT().GetUserByEmail(gomock.Any(), email).Return(models.User{}, errors.New("database error"))
			},
			expectedErr: platformerrors.ErrInternal,
		},
		{
			name: "failure - admin user already exists",
			setupMocks: func(mockDB *databasemocks.MockQuerier, mockGenerator *generatormocks.MockNumberGenerator) {
				mockDB.EXPECT().GetUserByEmail(gomock.Any(), email).Return(models.User{Email: email}, nil)
			},
			expectedErr: platformerrors.MakeApiError(400, "admin user already exists"),
		},
		{
			name: "failure - SaveUser fails",
			setupMocks: func(mockDB *databasemocks.MockQuerier, mockGenerator *generatormocks.MockNumberGenerator) {
				mockDB.EXPECT().GetUserByEmail(gomock.Any(), email).Return(models.User{}, sql.ErrNoRows)
				mockDB.EXPECT().SaveUser(gomock.Any(), models.SaveUserParams{Email: email, Password: password}).Return(nil, errors.New("failed to save user"))
			},
			expectedErr: platformerrors.ErrInternal,
		},
		{
			name: "failure - SaveAccount fails",
			setupMocks: func(mockDB *databasemocks.MockQuerier, mockGenerator *generatormocks.MockNumberGenerator) {
				mockDB.EXPECT().GetUserByEmail(gomock.Any(), email).Return(models.User{}, sql.ErrNoRows)
				mockDB.EXPECT().SaveUser(gomock.Any(), models.SaveUserParams{Email: email, Password: password}).
					Return(&models.User{ID: userID, Email: email}, nil)
				mockGenerator.EXPECT().Generate().Return("ADMIN123")
				mockDB.EXPECT().SaveAccount(gomock.Any(), models.SaveAccountParams{
					UserID:        userID,
					FirstName:     "Admin",
					LastName:      "Admin",
					AccountType:   models.AccountTypeADMIN,
					Status:        models.StatusACTIVE,
					AccountNumber: "ADMIN123",
				}).Return(nil, errors.New("failed to save account"))
			},
			expectedErr: platformerrors.ErrInternal,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := databasemocks.NewMockQuerier(ctrl)
			mockGenerator := generatormocks.NewMockNumberGenerator(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockDB, mockGenerator)
			}

			generator.DefaultNumberGenerator = mockGenerator
			s := service{
				db: mockDB,
			}

			err := s.InitialiseAdmin(ctx, email, password)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
