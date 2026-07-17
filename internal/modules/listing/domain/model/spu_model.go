package model

import (
	"net/url"
	"strings"
	"time"

	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/errs"
)

// SpuModel is the Standard Product Unit: the listing shell (title, category,
// brand, images) shared by its SKUs. Lifecycle: draft -> published -> off_shelf,
// with re-publish allowed from off_shelf.
type SpuModel struct {
	id          uint64
	title       string
	description string
	categoryID  uint64
	brand       *string
	images      []string
	status      enum.ListingStatusEnum
	version     uint64
	createdAt   time.Time
	updatedAt   time.Time
}

func NewSpuModel(
	title, description string,
	categoryID uint64,
	brand *string,
	images []string,
) (SpuModel, error) {
	if err := validateSpu(title, categoryID, images); err != nil {
		return SpuModel{}, err
	}

	status, err := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	if err != nil {
		return SpuModel{}, err
	}

	now := time.Now().UTC()
	return SpuModel{
		title:       strings.TrimSpace(title),
		description: description,
		categoryID:  categoryID,
		brand:       brand,
		images:      images,
		status:      status,
		version:     1,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func RestoreSpuModel(
	id uint64,
	title, description string,
	categoryID uint64,
	brand *string,
	images []string,
	status enum.ListingStatusEnum,
	version uint64,
	createdAt, updatedAt time.Time,
) (SpuModel, error) {
	if id == 0 {
		return SpuModel{}, errs.ErrSpuIDRequired
	}

	if err := validateSpu(title, categoryID, images); err != nil {
		return SpuModel{}, err
	}

	return SpuModel{
		id:          id,
		title:       title,
		description: description,
		categoryID:  categoryID,
		brand:       brand,
		images:      images,
		status:      status,
		version:     version,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

func (s *SpuModel) ID() uint64 {
	return s.id
}

func (s *SpuModel) Title() string {
	return s.title
}

func (s *SpuModel) Description() string {
	return s.description
}

func (s *SpuModel) CategoryID() uint64 {
	return s.categoryID
}

func (s *SpuModel) Brand() *string {
	return s.brand
}

func (s *SpuModel) Images() []string {
	return s.images
}

func (s *SpuModel) Status() enum.ListingStatusEnum {
	return s.status
}

func (s *SpuModel) Version() uint64 {
	return s.version
}

func (s *SpuModel) CreatedAt() time.Time {
	return s.createdAt
}

func (s *SpuModel) UpdatedAt() time.Time {
	return s.updatedAt
}

// Publish makes the SPU visible; allowed from draft and off_shelf
func (s *SpuModel) Publish() error {
	if s.status.IsPublished() {
		return errs.ErrSpuAlreadyPublished
	}

	status, err := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	if err != nil {
		return err
	}

	s.status = status
	s.touch()
	return nil
}

// OffShelf takes the SPU off shelf; allowed only from published
func (s *SpuModel) OffShelf() error {
	if !s.status.IsPublished() {
		return errs.ErrSpuNotPublished
	}

	status, err := enum.NewListingStatusEnum(enum.EnumListingStatusOffShelf)
	if err != nil {
		return err
	}

	s.status = status
	s.touch()
	return nil
}

// Update changes the SPU attributes; allowed only in draft and off_shelf
func (s *SpuModel) Update(
	title, description string,
	categoryID uint64,
	brand *string,
	images []string,
) error {
	if s.status.IsPublished() {
		return errs.ErrSpuNotEditable
	}

	if err := validateSpu(title, categoryID, images); err != nil {
		return err
	}

	s.title = strings.TrimSpace(title)
	s.description = description
	s.categoryID = categoryID
	s.brand = brand
	s.images = images
	s.touch()
	return nil
}

func (s *SpuModel) touch() {
	s.version++
	s.updatedAt = time.Now().UTC()
}

func validateSpu(title string, categoryID uint64, images []string) error {
	if strings.TrimSpace(title) == "" {
		return errs.ErrSpuTitleRequired
	}

	if categoryID == 0 {
		return errs.ErrSpuCategoryRequired
	}

	for _, image := range images {
		if err := validateImageURL(image); err != nil {
			return err
		}
	}

	return nil
}

func validateImageURL(rawURL string) error {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return errs.ErrImageURLInvalid
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errs.ErrImageURLInvalid
	}

	return nil
}
