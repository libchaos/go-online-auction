package errs

import "errors"

var (
	ErrCategoryNotFound       = errors.New("category not found")
	ErrCategoryIDRequired     = errors.New("category id must be greater than zero")
	ErrCategoryNameRequired   = errors.New("category name is required")
	ErrCategoryParentNotFound = errors.New("parent category not found")
	ErrCategoryHasChildren    = errors.New("category has child categories and cannot be deleted")
	ErrCategoryInUse          = errors.New("category is referenced by SPUs and cannot be deleted")

	ErrSpuNotFound         = errors.New("spu not found")
	ErrSpuIDRequired       = errors.New("spu id must be greater than zero")
	ErrSpuTitleRequired    = errors.New("spu title is required")
	ErrSpuCategoryRequired = errors.New("spu category id must be greater than zero")
	ErrSpuNotEditable      = errors.New("spu can only be edited in draft or off_shelf status")
	ErrSpuAlreadyPublished = errors.New("spu is already published")
	ErrSpuNotPublished     = errors.New("spu can only be taken off shelf from published status")
	ErrSpuMustBePublished  = errors.New("parent spu must be published")

	ErrSkuNotFound           = errors.New("sku not found")
	ErrSkuIDRequired         = errors.New("sku id must be greater than zero")
	ErrSkuPriceInvalid       = errors.New("sku price must be greater than zero")
	ErrSkuSpecValuesRequired = errors.New("sku spec values are required")
	ErrSkuNotEditable        = errors.New("sku can only be edited in draft or off_shelf status")
	ErrSkuAlreadyPublished   = errors.New("sku is already published")
	ErrSkuNotPublished       = errors.New("sku can only be taken off shelf from published status")

	ErrInvalidListingStatus = errors.New("invalid listing status")
	ErrImageURLInvalid      = errors.New("image url must be a valid http or https url")
	ErrConcurrencyConflict  = errors.New("concurrency conflict: resource was modified")
	ErrTransactionFailed    = errors.New("transaction failed")
)
