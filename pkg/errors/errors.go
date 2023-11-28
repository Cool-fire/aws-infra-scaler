package errors

import (
	"fmt"
	"github.com/Cool-fire/aws-infra-scaler/pkg/service"
)

type ScalingFailureError struct {
	ServiceName  service.Service
	IdentifierId string
	Reason       string
}

func (s *ScalingFailureError) Error() string {
	return fmt.Sprintf("Scaling failed for service %s with identifier %s. Reason: %s", s.ServiceName, s.IdentifierId, s.Reason)
}
