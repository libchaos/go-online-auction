package event

const (
	SpuOffShelfEventType = "listing.spu.off_shelf"
)

type SpuOffShelfEvent struct {
	DomainEvent
	spuID uint64
}

func NewSpuOffShelfEvent(spuID uint64) SpuOffShelfEvent {
	return SpuOffShelfEvent{
		DomainEvent: newDomainEvent(),
		spuID:       spuID,
	}
}

func (e SpuOffShelfEvent) SpuID() uint64 {
	return e.spuID
}
