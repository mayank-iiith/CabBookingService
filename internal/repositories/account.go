package repositories

import (
	"context"

	"CabBookingService/internal/db"
	"CabBookingService/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountRepository interface {
	GetByUsername(ctx context.Context, username string) (*models.Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error)

	// Create We don't necessarily need a separate Create here as AuthService handles it transactionally,
	// but it's good practice to have.
	Create(ctx context.Context, account *models.Account) error
	AddRole(ctx context.Context, accountID uuid.UUID, role models.Role) error
}

type gormAccountRepository struct {
	db *gorm.DB
}

func NewGormAccountRepository(db *gorm.DB) AccountRepository {
	return &gormAccountRepository{db: db}
}

func (r *gormAccountRepository) Create(ctx context.Context, account *models.Account) error {
	tx := db.NewGormTx(ctx, r.db)
	return tx.Create(account).Error
}

func (r *gormAccountRepository) GetByUsername(ctx context.Context, username string) (*models.Account, error) {
	tx := db.NewGormTx(ctx, r.db)

	var account models.Account
	// Critical: Preload "Roles" so we know if they are a Passenger/Driver/Admin
	err := tx.Where("username = ?", username).
		Preload("Roles").
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *gormAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	tx := db.NewGormTx(ctx, r.db)

	var account models.Account
	// Critical: Preload "Roles" so we know if they are a Passenger/Driver/Admin
	err := tx.Where("id = ?", id).
		Preload("Roles").
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *gormAccountRepository) AddRole(ctx context.Context, accountID uuid.UUID, role models.Role) error {
	tx := db.NewGormTx(ctx, r.db)

	// GORM association mode handles the many-to-many insert into 'account_roles'
	// We only need the ID to update the association. GORM will use this ID to insert a row into the 'account_roles' join table.
	var account models.Account
	account.ID = accountID

	return tx.Model(&account).Association("Roles").Append(&role)
}
