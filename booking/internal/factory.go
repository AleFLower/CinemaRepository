package internal

import "fmt"

func SeatStrategyFactory(strategyName string) (SeatStrategy, error) {
	switch strategyName {
	case "no-gap":
		return &NoGapStrategy{}, nil
	case "social":
		return &SocialDistancingStrategy{}, nil
	case "default", "":
		return &DefaultStrategy{}, nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategyName)
	}
}
