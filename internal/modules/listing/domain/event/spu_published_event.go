package event

const (
	SpuPublishedEventType = "listing.spu.published"
)

type SpuPublishedEvent struct {
	DomainEvent
	spuID uint64
}

func NewSpuPublishedEvent(spuID uint64) SpuPublishedEvent {
	return SpuPublishedEvent{
		DomainEvent: newDomainEvent(),
		spuID:       spuID,
	}
}

func (e SpuPublishedEvent) SpuID() uint64 {
	return e.spuID
}
