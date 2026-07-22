package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ListingSpuResolutionTestSuite struct {
	suite.Suite
}

func TestListingSpuResolutionSuite(t *testing.T) {
	suite.Run(t, new(ListingSpuResolutionTestSuite))
}

func (s *ListingSpuResolutionTestSuite) TestResolveListingSpuID_SpuEvent_UsesAggregateID() {
	envelope := listingEnvelope{AggregateType: "spu", AggregateID: 42}

	spuID, ok := resolveListingSpuID(envelope)

	s.Require().True(ok)
	s.Equal(uint64(42), spuID)
}

func (s *ListingSpuResolutionTestSuite) TestResolveListingSpuID_SkuEvent_ReadsNestedSpuID() {
	data, marshalErr := json.Marshal(listingSkuData{SkuID: 7, SpuID: 42})
	s.Require().NoError(marshalErr)
	envelope := listingEnvelope{AggregateType: "sku", AggregateID: 7, Data: data}

	spuID, ok := resolveListingSpuID(envelope)

	s.Require().True(ok)
	s.Equal(uint64(42), spuID)
}

func (s *ListingSpuResolutionTestSuite) TestResolveListingSpuID_UnknownAggregate_ReturnsFalse() {
	envelope := listingEnvelope{AggregateType: "other", AggregateID: 1}

	_, ok := resolveListingSpuID(envelope)

	s.Require().False(ok)
}
