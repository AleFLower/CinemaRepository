package internal

type SeatStrategy interface {
	Validate(room *RoomState, seat int32) error
}
