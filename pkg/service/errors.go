package service

import (
	"fmt"
)

type ScalingFailureError struct {
	ServiceName  Service
	IdentifierId string
	Reason       string
}

func (s *ScalingFailureError) Error() string {
	return fmt.Sprintf("Scaling failed for service %s with identifier %s. Reason: %s", s.ServiceName, s.IdentifierId, s.Reason)
}
