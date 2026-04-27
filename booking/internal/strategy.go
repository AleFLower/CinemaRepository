package internal

// SeatStrategy definisce l'interfaccia comune per gli algoritmi di selezione

// SeatStrategy definisce l'interfaccia comune
type SeatStrategy interface {
	Validate(room *RoomState, seat int32) error
}
