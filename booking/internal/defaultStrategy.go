package internal

import "fmt"

type DefaultStrategy struct{}

func (s *DefaultStrategy) Validate(room *RoomState, seat int32) error {
	if room.Seats[seat] {
		return fmt.Errorf("seat %d already taken", seat)
	}
	return nil
}
