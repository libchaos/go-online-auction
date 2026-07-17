package event

const (
	SkuOffShelfEventType = "listing.sku.off_shelf"
)

type SkuOffShelfEvent struct {
	DomainEvent
	skuID uint64
	spuID uint64
}

func NewSkuOffShelfEvent(skuID, spuID uint64) SkuOffShelfEvent {
	return SkuOffShelfEvent{
		DomainEvent: newDomainEvent(),
		skuID:       skuID,
		spuID:       spuID,
	}
}

func (e SkuOffShelfEvent) SkuID() uint64 {
	return e.skuID
}

func (e SkuOffShelfEvent) SpuID() uint64 {
	return e.spuID
}
