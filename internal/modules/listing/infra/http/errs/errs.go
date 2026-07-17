package errs

import (
	"errors"
	"net/http"

	domainerrs "auction/internal/modules/listing/domain/errs"
	"auction/pkg/errs"
)

var (
	ErrCategoryNotFound     = errs.New("LISTING_01", "Category not found", http.StatusNotFound, nil)
	ErrCategoryNameRequired = errs.New(
		"LISTING_02",
		"Category name is required",
		http.StatusBadRequest,
		nil,
	)
	ErrCategoryParentNotFound = errs.New(
		"LISTING_03",
		"Parent category not found",
		http.StatusBadRequest,
		nil,
	)
	ErrCategoryHasChildren = errs.New(
		"LISTING_04",
		"Category has child categories and cannot be deleted",
		http.StatusConflict,
		nil,
	)
	ErrCategoryInUse = errs.New(
		"LISTING_05",
		"Category is referenced by SPUs and cannot be deleted",
		http.StatusConflict,
		nil,
	)
	ErrSpuNotFound         = errs.New("LISTING_06", "SPU not found", http.StatusNotFound, nil)
	ErrSpuTitleRequired    = errs.New("LISTING_07", "SPU title is required", http.StatusBadRequest, nil)
	ErrSpuCategoryRequired = errs.New(
		"LISTING_08",
		"SPU category ID must be greater than zero",
		http.StatusBadRequest,
		nil,
	)
	ErrSpuNotEditable = errs.New(
		"LISTING_09",
		"SPU can only be edited in draft or off_shelf status",
		http.StatusBadRequest,
		nil,
	)
	ErrSpuAlreadyPublished = errs.New("LISTING_10", "SPU is already published", http.StatusBadRequest, nil)
	ErrSpuNotPublished     = errs.New(
		"LISTING_11",
		"SPU can only be taken off shelf from published status",
		http.StatusBadRequest,
		nil,
	)
	ErrSpuMustBePublished = errs.New(
		"LISTING_12",
		"Parent SPU must be published",
		http.StatusBadRequest,
		nil,
	)
	ErrSkuNotFound           = errs.New("LISTING_13", "SKU not found", http.StatusNotFound, nil)
	ErrSkuPriceInvalid       = errs.New("LISTING_14", "SKU price must be greater than zero", http.StatusBadRequest, nil)
	ErrSkuSpecValuesRequired = errs.New(
		"LISTING_15",
		"SKU spec values are required",
		http.StatusBadRequest,
		nil,
	)
	ErrSkuNotEditable = errs.New(
		"LISTING_16",
		"SKU can only be edited in draft or off_shelf status",
		http.StatusBadRequest,
		nil,
	)
	ErrSkuAlreadyPublished = errs.New("LISTING_17", "SKU is already published", http.StatusBadRequest, nil)
	ErrSkuNotPublished     = errs.New(
		"LISTING_18",
		"SKU can only be taken off shelf from published status",
		http.StatusBadRequest,
		nil,
	)
	ErrInvalidListingStatus = errs.New("LISTING_19", "Invalid listing status", http.StatusBadRequest, nil)
	ErrImageURLInvalid      = errs.New(
		"LISTING_20",
		"Image URL must be a valid http or https URL",
		http.StatusBadRequest,
		nil,
	)
	ErrOptimisticLockFailed = errs.New(
		"LISTING_21",
		"Resource was modified by another transaction",
		http.StatusConflict,
		nil,
	)
	ErrTransactionFailed = errs.New("LISTING_22", "Transaction failed", http.StatusInternalServerError, nil)
	ErrInvalidRequest    = errs.New("LISTING_23", "Invalid request body", http.StatusBadRequest, nil)
	ErrInvalidID         = errs.New("LISTING_24", "Invalid ID", http.StatusBadRequest, nil)
)

var domainToHTTPErrorMap = []struct {
	domainError error
	httpError   error
}{
	{domainerrs.ErrCategoryNotFound, ErrCategoryNotFound},
	{domainerrs.ErrCategoryNameRequired, ErrCategoryNameRequired},
	{domainerrs.ErrCategoryParentNotFound, ErrCategoryParentNotFound},
	{domainerrs.ErrCategoryHasChildren, ErrCategoryHasChildren},
	{domainerrs.ErrCategoryInUse, ErrCategoryInUse},
	{domainerrs.ErrSpuNotFound, ErrSpuNotFound},
	{domainerrs.ErrSpuTitleRequired, ErrSpuTitleRequired},
	{domainerrs.ErrSpuCategoryRequired, ErrSpuCategoryRequired},
	{domainerrs.ErrSpuNotEditable, ErrSpuNotEditable},
	{domainerrs.ErrSpuAlreadyPublished, ErrSpuAlreadyPublished},
	{domainerrs.ErrSpuNotPublished, ErrSpuNotPublished},
	{domainerrs.ErrSpuMustBePublished, ErrSpuMustBePublished},
	{domainerrs.ErrSkuNotFound, ErrSkuNotFound},
	{domainerrs.ErrSkuPriceInvalid, ErrSkuPriceInvalid},
	{domainerrs.ErrSkuSpecValuesRequired, ErrSkuSpecValuesRequired},
	{domainerrs.ErrSkuNotEditable, ErrSkuNotEditable},
	{domainerrs.ErrSkuAlreadyPublished, ErrSkuAlreadyPublished},
	{domainerrs.ErrSkuNotPublished, ErrSkuNotPublished},
	{domainerrs.ErrInvalidListingStatus, ErrInvalidListingStatus},
	{domainerrs.ErrImageURLInvalid, ErrImageURLInvalid},
	{domainerrs.ErrConcurrencyConflict, ErrOptimisticLockFailed},
	{domainerrs.ErrTransactionFailed, ErrTransactionFailed},
}

func MapDomainError(err error) error {
	for _, mapping := range domainToHTTPErrorMap {
		if errors.Is(err, mapping.domainError) {
			return mapping.httpError
		}
	}
	return err
}
