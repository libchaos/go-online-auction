package repository

import (
	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/infra/sqlcgen"
)

func toNullListingStatus(status *enum.ListingStatusEnum) sqlcgen.NullListingStatus {
	if status == nil {
		return sqlcgen.NullListingStatus{}
	}
	return sqlcgen.NullListingStatus{
		ListingStatus: sqlcgen.ListingStatus(status.String()),
		Valid:         true,
	}
}

func toNullableInt64Filter(v *uint64) *int64 {
	if v == nil {
		return nil
	}
	i := int64(*v)
	return &i
}
