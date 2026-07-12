package service

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
)

func TestCountAutoScalingGroupCapacityCountsMixedInstancesOnce(t *testing.T) {
	group := &types.AutoScalingGroup{
		Instances: []types.Instance{
			{
				InstanceId:     strPtr("i-on-demand-1"),
				LifecycleState: types.LifecycleStateInService,
			},
			{
				InstanceId:     strPtr("i-spot-1"),
				LifecycleState: types.LifecycleStateInService,
			},
			{
				InstanceId:     strPtr("i-spot-1"),
				LifecycleState: types.LifecycleStateInService,
			},
			{
				InstanceId:     strPtr("i-spot-2"),
				LifecycleState: types.LifecycleStatePending,
			},
		},
	}

	if got, want := countAutoScalingGroupCapacity(group), int32(3); got != want {
		t.Fatalf("countAutoScalingGroupCapacity() = %d, want %d", got, want)
	}
}

func TestCountAutoScalingGroupCapacityIgnoresNonServingInstances(t *testing.T) {
	group := &types.AutoScalingGroup{
		Instances: []types.Instance{
			{
				InstanceId:     strPtr("i-in-service"),
				LifecycleState: types.LifecycleStateInService,
			},
			{
				InstanceId:     strPtr("i-standby"),
				LifecycleState: types.LifecycleStateStandby,
			},
			{
				InstanceId:     strPtr("i-terminating"),
				LifecycleState: types.LifecycleStateTerminating,
			},
		},
	}

	if got, want := countAutoScalingGroupCapacity(group), int32(1); got != want {
		t.Fatalf("countAutoScalingGroupCapacity() = %d, want %d", got, want)
	}
}

func strPtr(value string) *string {
	return &value
}
