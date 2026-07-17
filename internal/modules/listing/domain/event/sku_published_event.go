package event

const (
	SkuPublishedEventType = "listing.sku.published"
)

type SkuPublishedEvent struct {
	DomainEvent
	skuID        uint64
	spuID        uint64
	priceInCents uint64
	quantity     uint64
}

func NewSkuPublishedEvent(skuID, spuID, priceInCents, quantity uint64) SkuPublishedEvent {
	return SkuPublishedEvent{
		DomainEvent:  newDomainEvent(),
		skuID:        skuID,
		spuID:        spuID,
		priceInCents: priceInCents,
		quantity:     quantity,
	}
}

func (e SkuPublishedEvent) SkuID() uint64 {
	return e.skuID
}

func (e SkuPublishedEvent) SpuID() uint64 {
	return e.spuID
}

func (e SkuPublishedEvent) PriceInCents() uint64 {
	return e.priceInCents
}

func (e SkuPublishedEvent) Quantity() uint64 {
	return e.quantity
}
