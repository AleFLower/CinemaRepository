package internal

import "errors"

type SocialDistancingStrategy struct{}

func (s *SocialDistancingStrategy) Validate(room *RoomState, seat int32) error {
	if room.Seats[seat] {
		return errors.New("posto già occupato")
	}

	maxSeat := int32(len(room.Seats))

	if (seat > 1 && room.Seats[seat-1]) ||
		(seat < maxSeat && room.Seats[seat+1]) {
		return errors.New("devi lasciare almeno un posto vuoto tra le persone")
	}

	return nil
}
