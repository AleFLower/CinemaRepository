package internal

import "errors"

type NoGapStrategy struct{}

func (s *NoGapStrategy) Validate(room *RoomState, seat int32) error {
	if room.Seats[seat] {
		return errors.New("seat already taken")
	}

	maxSeat := int32(len(room.Seats))

	// LEFT SIDE
	if seat > 1 {
		leftFree := !room.Seats[seat-1]

		leftLeftOccupied := true
		if seat-2 > 0 {
			leftLeftOccupied = room.Seats[seat-2]
		}

		if leftFree && leftLeftOccupied {
			return errors.New("you would leave an isolated seat on the left")
		}
	}

	// RIGHT SIDE
	if seat < maxSeat {
		rightFree := !room.Seats[seat+1]

		rightRightOccupied := true
		if seat+2 <= maxSeat {
			rightRightOccupied = room.Seats[seat+2]
		}

		if rightFree && rightRightOccupied {
			return errors.New("you would leave an isolated seat on the right")
		}
	}

	return nil
}
