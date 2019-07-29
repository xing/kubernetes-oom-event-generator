package controller

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	listersV1 "k8s.io/client-go/listers/core/v1"
)

func init() {
	if os.Getenv("V") == "" {
		flag.Set("stderrthreshold", "5")
	}
}

func TestEvaluatingUninterestingEvent(t *testing.T) {
	c := controller()
	recorder := dummyRecorder()
	c.recorder = recorder
	event := event("Scaled", "ReplicaSet", "my-namespace", "rs-247484f")
	c.evaluateEvent(event)

	assert.Equal(t, len(recorder.Events), 0)
}

func TestEvaluatingInterestingEvent(t *testing.T) {
	ns := "my-namespace"
	podName := "rs-247484f"
	c := controller()
	p := pod("OOMKilled", 1, c.startTime.Add(120))
	podLister := &dummyPodLister{
		MethodCalls: []methodCall{},
		Errors:      []error{},
		PodCollection: map[string]map[string]*core.Pod{
			ns: map[string]*core.Pod{
				podName: p,
			},
		},
	}
	recorder := dummyRecorder()
	c.recorder = recorder
	c.podLister = podLister
	event := event(startedEvent, podKind, ns, podName)
	c.evaluateEvent(event)

	assert.Equal(t, podLister.MethodCalls, []methodCall{
		methodCall{
			Method:   "Pods",
			Argument: ns,
		},
		methodCall{
			Method:   "Get",
			Argument: podName,
		},
	})
	assert.Equal(t, []dummyEvent{
		dummyEvent{
			Obj:       p,
			EventType: core.EventTypeWarning,
			Reason:    "PreviousContainerWasOOMKilled",
			Message:   "The previous instance of the container 'our-container' (our-container-1234) was OOMKilled",
		},
	}, recorder.Events)
}

func TestEvaluatingPodStatusOnNotOOMKilled(t *testing.T) {
	c := controller()
	recorder := dummyRecorder()
	c.recorder = recorder
	p := pod("", 1, c.startTime.Add(120))
	c.evaluatePodStatus(p)

	assert.Equal(t, len(recorder.Events), 0)
}

func TestEvaluatingPodStatusOnOOMKilled(t *testing.T) {
	c := controller()
	recorder := dummyRecorder()
	c.recorder = recorder
	p := pod("OOMKilled", 1, c.startTime.Add(120))
	c.evaluatePodStatus(p)

	assert.Equal(t, []dummyEvent{
		dummyEvent{
			Obj:       p,
			EventType: core.EventTypeWarning,
			Reason:    "PreviousContainerWasOOMKilled",
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

func event(reason, kind, namespace, name string) *core.Event {
	return &core.Event{
		Reason: reason,
		InvolvedObject: core.ObjectReference{
			Kind:      kind,
			Namespace: namespace,
			Name:      name,
		},
	}
}

func pod(terminationReason string, restartCount int32, finishedAt time.Time) *core.Pod {
	terminatedState := core.ContainerStateTerminated{
		Reason:     terminationReason,
		FinishedAt: metav1.NewTime(finishedAt),
	}
	containerStatus := core.ContainerStatus{
		ContainerID: "our-container-1234",
		Name:        "our-container",
		LastTerminationState: core.ContainerState{
			Terminated: &terminatedState,
		},
		RestartCount: restartCount,
	}
	objMeta := metav1.ObjectMeta{
		Name:      "our-pod",
		Namespace: "our-pod-namespace",
	}
	return &core.Pod{
		ObjectMeta: objMeta,
		Status: core.PodStatus{
			ContainerStatuses: []core.ContainerStatus{
				containerStatus,
			},
		},
	}
}

type dummyPodLister struct {
	MethodCalls   []methodCall
	Errors        []error
	PodCollection map[string]map[string]*core.Pod
}

type methodCall struct {
	Method   string
	Argument string
}

func (d *dummyPodLister) Pods(namespace string) listersV1.PodNamespaceLister {
	d.MethodCalls = append(d.MethodCalls, methodCall{
		Method:   "Pods",
		Argument: namespace,
	})
	return d
}

func (d *dummyPodLister) Get(name string) (*core.Pod, error) {
	d.MethodCalls = append(d.MethodCalls, methodCall{
		Method:   "Get",
		Argument: name,
	})
	if len(d.MethodCalls)-2 < 0 {
		return nil, fmt.Errorf("did not call Pods(namespace string) first")
	}
	call := d.MethodCalls[len(d.MethodCalls)-2]
	if ns, ok := d.PodCollection[call.Argument]; ok {
		return ns[name], nil
	}
	return nil, d.popError()
}

func (d *dummyPodLister) List(selector labels.Selector) (ret []*v1.Pod, err error) {
	return
}

func (d *dummyPodLister) popError() (err error) {
	if len(d.Errors) != 0 {
		err, d.Errors = d.Errors[0], d.Errors[1:]
	}
	return
}
