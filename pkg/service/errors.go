package service

import (
	"fmt"
)

type ScalingError struct {
	Region       string
	ServiceName  string
	IdentifierId string
	Err          error
}

func (s *ScalingError) Error() string {
	return fmt.Sprintf("Scaling failed for service %s with identifier %s. Reason: %s", s.ServiceName, s.IdentifierId, s.Err.Error())
}
