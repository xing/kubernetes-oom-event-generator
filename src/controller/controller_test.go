package controller

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	corev1_api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func init() {
	if os.Getenv("V") == "" {
		flag.Set("stderrthreshold", "5")
	}
}

func TestSetNewContainerRestart(t *testing.T) {
	c := controller()

	changed := c.setRestartCount("pod-1", "container-1", 1)

	assert.Equal(t, len(c.podState), 1)
	assert.Equal(t, changed, true)
}

func TestSetExistingContainerRestart(t *testing.T) {
	c := controller()

	c.setRestartCount("pod-1", "container-1", 1)
	changed := c.setRestartCount("pod-1", "container-1", 1)

	assert.Equal(t, len(c.podState), 1)
	assert.Equal(t, changed, false)
}

func TestSetMultipleNewContainerRestart(t *testing.T) {
	c := controller()

	changed1 := c.setRestartCount("pod-1", "container-1", 1)
	changed2 := c.setRestartCount("pod-1", "container-1", 2)
	changed3 := c.setRestartCount("pod-1", "container-1", 1)

	assert.Equal(t, len(c.podState), 1)
	assert.Equal(t, changed1, true)
	assert.Equal(t, changed2, true)
	assert.Equal(t, changed3, false)
}

func TestSetTwoContainerRestart(t *testing.T) {
	c := controller()

	changed1 := c.setRestartCount("pod-1", "container-1", 3)
	changed2 := c.setRestartCount("pod-1", "container-2", 9)

	assert.Equal(t, c.podState, map[string]map[string]int32(map[string]map[string]int32{"pod-1": map[string]int32{"container-2": 9, "container-1": 3}}))
	assert.Equal(t, changed1, true)
	assert.Equal(t, changed2, true)
}

func TestDeletePodState(t *testing.T) {
	c := controller()

	c.setRestartCount("pod-1", "container-1", 1)
	assert.Equal(t, len(c.podState), 1)

	c.deletePodState("pod-1")
	assert.Equal(t, len(c.podState), 0)
}

func TestEvaluatingPodStatusOnNotOOMKilled(t *testing.T) {
	c := controller()
	p := pod("", 1, c.startTime.Add(120))
	c.evaluatePodStatus(p)

	assert.Equal(t, len(c.podState), 0)
}

func TestEvaluatingPodStatusOnOOMKilled(t *testing.T) {
	c := controller()
	recorder := dummyRecorder()
	c.recorder = recorder
	p := pod("OOMKilled", 1, c.startTime.Add(120))
	c.evaluatePodStatus(p)

	assert.Equal(t, 1, len(c.podState))
	assert.Equal(t, []dummyEvent{
		dummyEvent{
			Obj:       p,
			EventType: v1.EventTypeWarning,
			Reason:    "PreviousPodWasOOMKilled",
			Message:   "The previous instance of the container 'our-container' (our-container-1234) was OOMKilled",
		},
	}, recorder.Events)
}

func controller() *Controller {
	stopChan := make(chan struct{})
	controller := NewController(stopChan)
	return controller
}

type dummyEvent struct {
	Obj       runtime.Object
	EventType string
	Reason    string
	Message   string
}
type dummyEventRecorder struct {
	Events []dummyEvent
}

func (r *dummyEventRecorder) Event(_object runtime.Object, _eventtype, _reason, _message string) {}

func (r *dummyEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	r.Events = append(r.Events, dummyEvent{
		Obj:       object,
		EventType: eventtype,
		Reason:    reason,
		Message:   fmt.Sprintf(messageFmt, args...),
	})
	return
}

func (r *dummyEventRecorder) PastEventf(_object runtime.Object, _timestamp metav1.Time, _eventtype, _reason, _messageFmt string, _args ...interface{}) {
}

func (r *dummyEventRecorder) AnnotatedEventf(_object runtime.Object, _annotations map[string]string, _eventtype, _reason, _messageFmt string, _args ...interface{}) {
}

func dummyRecorder() *dummyEventRecorder {
	return &dummyEventRecorder{}
}

func pod(terminationReason string, restartCount int32, finishedAt time.Time) *corev1_api.Pod {
	terminatedState := corev1_api.ContainerStateTerminated{
		Reason:     terminationReason,
		FinishedAt: metav1.NewTime(finishedAt),
	}
	containerStatus := corev1_api.ContainerStatus{
		ContainerID: "our-container-1234",
		Name:        "our-container",
		LastTerminationState: corev1_api.ContainerState{
			Terminated: &terminatedState,
		},
		RestartCount: restartCount,
	}
	objMeta := metav1.ObjectMeta{
		Name:      "our-pod",
		Namespace: "our-pod-namespace",
	}
	return &corev1_api.Pod{
		ObjectMeta: objMeta,
		Status: corev1_api.PodStatus{
			ContainerStatuses: []corev1_api.ContainerStatus{
				containerStatus,
			},
		},
	}
}
