package usecase

import (
	"context"
	"strings"
	"unicode/utf8"

	"food-rescue/internal/entity"
	"food-rescue/internal/repository"
)

type PartnerUsecase struct {
	partnerRepo *repository.PartnerRepository
}

func NewPartnerUsecase(
	partnerRepo *repository.PartnerRepository,
) *PartnerUsecase {
	return &PartnerUsecase{partnerRepo: partnerRepo}
}

// PartnerInput dipakai untuk membuat dan mengubah mitra.
// IsActive bersifat opsional: nil berarti "default aktif" saat membuat
// dan "tidak diubah" saat memperbarui.
type PartnerInput struct {
	Name     string
	Type     string
	Address  string
	Phone    string
	IsActive *bool
}

func (in *PartnerInput) normalize() {
	in.Name = strings.TrimSpace(in.Name)
	in.Type = strings.ToLower(strings.TrimSpace(in.Type))
	in.Address = strings.TrimSpace(in.Address)
	in.Phone = strings.TrimSpace(in.Phone)
}

func validatePartnerInput(in PartnerInput) error {
	if in.Name == "" {
		return entity.NewDomainError(entity.ErrInvalid, "nama mitra wajib diisi")
	}

	if utf8.RuneCountInString(in.Name) > 150 {
		return entity.NewDomainError(entity.ErrInvalid, "nama mitra maksimal 150 karakter")
	}

	if in.Type == "" {
		return entity.NewDomainError(entity.ErrInvalid, "jenis mitra wajib diisi (contoh: panti_asuhan, panti_jompo, komunitas)")
	}

	if utf8.RuneCountInString(in.Type) > 50 {
		return entity.NewDomainError(entity.ErrInvalid, "jenis mitra maksimal 50 karakter")
	}

	if in.Address == "" {
		return entity.NewDomainError(entity.ErrInvalid, "alamat mitra wajib diisi")
	}

	if utf8.RuneCountInString(in.Phone) > 20 {
		return entity.NewDomainError(entity.ErrInvalid, "nomor telepon maksimal 20 karakter")
	}

	return nil
}

func (u *PartnerUsecase) Create(ctx context.Context, input PartnerInput) (*entity.Partner, error) {
	input.normalize()

	if err := validatePartnerInput(input); err != nil {
		return nil, err
	}

	partner := &entity.Partner{
		Name:     input.Name,
		Type:     input.Type,
		Address:  input.Address,
		Phone:    input.Phone,
		IsActive: true,
	}

	if input.IsActive != nil {
		partner.IsActive = *input.IsActive
	}

	if err := u.partnerRepo.Create(ctx, partner); err != nil {
		return nil, err
	}

	return partner, nil
}

func (u *PartnerUsecase) List(ctx context.Context, activeOnly bool) ([]entity.Partner, error) {
	return u.partnerRepo.FindAll(ctx, activeOnly)
}

func (u *PartnerUsecase) GetByID(ctx context.Context, id int) (*entity.Partner, error) {
	if id <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID mitra tidak valid")
	}

	return u.partnerRepo.FindByID(ctx, id)
}

// Update mengganti data mitra. Menonaktifkan mitra (is_active = false) membuat
// mitra tidak bisa dipilih untuk alokasi baru, tanpa menghapus riwayat distribusinya.
func (u *PartnerUsecase) Update(ctx context.Context, id int, input PartnerInput) (*entity.Partner, error) {
	if id <= 0 {
		return nil, entity.NewDomainError(entity.ErrInvalid, "ID mitra tidak valid")
	}

	input.normalize()

	if err := validatePartnerInput(input); err != nil {
		return nil, err
	}

	partner, err := u.partnerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	partner.Name = input.Name
	partner.Type = input.Type
	partner.Address = input.Address
	partner.Phone = input.Phone

	if input.IsActive != nil {
		partner.IsActive = *input.IsActive
	}

	if err := u.partnerRepo.Update(ctx, partner); err != nil {
		return nil, err
	}

	return partner, nil
}
