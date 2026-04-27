package internal

import "errors"

type NoGapStrategy struct{}

func (s *NoGapStrategy) Validate(room *RoomState, seat int32) error {
	if room.Seats[seat] {
		return errors.New("posto già occupato")
	}

	maxSeat := int32(len(room.Seats))

	// SINISTRA
	if seat > 1 {
		leftFree := !room.Seats[seat-1]

		leftLeftOccupied := true
		if seat-2 > 0 {
			leftLeftOccupied = room.Seats[seat-2]
		}

		if leftFree && leftLeftOccupied {
			return errors.New("lasceresti un posto isolato a sinistra")
		}
	}

	// DESTRA
	if seat < maxSeat {
		rightFree := !room.Seats[seat+1]

		rightRightOccupied := true
		if seat+2 <= maxSeat {
			rightRightOccupied = room.Seats[seat+2]
		}

		if rightFree && rightRightOccupied {
			return errors.New("lasceresti un posto isolato a destra")
		}
	}

	return nil
}
