package controller

import (
	"time"

	"github.com/golang/glog"
	"github.com/xing/kubernetes-oom-event-generator/src/util"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

const (
	// informerSyncMinute defines how often the cache is synced from Kubernetes
	informerSyncMinute = 3
)

// Controller is a controller that listens on Pod changes and create Kubernetes Events
// when a container reports it was previously killed
type Controller struct {
	Stop       chan struct{}
	k8sFactory informers.SharedInformerFactory
	recorder   record.EventRecorder
	startTime  time.Time
	stopCh     chan struct{}
	eventCh    chan *core.Event
}

// NewController returns an instance of the Controller
func NewController(stop chan struct{}) *Controller {
	k8sClient := util.Clientset()
	k8sFactory := informers.NewSharedInformerFactory(k8sClient, time.Minute*informerSyncMinute)

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})

	controller := &Controller{
		stopCh:     make(chan struct{}),
		Stop:       stop,
		k8sFactory: k8sFactory,
		eventCh:    make(chan *core.Event),
		recorder:   eventBroadcaster.NewRecorder(scheme.Scheme, core.EventSource{Component: "oom-event-generator"}),
		startTime:  time.Now(),
	}

	informers.SharedInformerFactory(k8sFactory).Core().V1().Pods().Informer()
	eventsInformer := informers.SharedInformerFactory(k8sFactory).Core().V1().Events().Informer()
	eventsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.eventAdded,
	})

	return controller
}

func (c *Controller) eventAdded(obj interface{}) {
	event := obj.(*core.Event)
	glog.V(2).Infof("got event %s/%s", event.ObjectMeta.Namespace, event.ObjectMeta.Name)
	c.eventCh <- event
}

// Run is the main loop that processes Kubernetes Pod changes
func (c *Controller) Run() error {
	c.k8sFactory.Start(c.stopCh)
	c.k8sFactory.WaitForCacheSync(nil)

	for {
		select {
		case event := <-c.eventCh:
			c.evaluateEvent(event)
		case <-c.Stop:
			glog.Info("Stopping")
			return nil
		}
	}
}

const startedEvent = "Started"
const podKind = "Pod"

func isContainerStartedEvent(event *core.Event) bool {
	return (event.Reason == startedEvent &&
		event.InvolvedObject.Kind == podKind)
}

func (c *Controller) evaluateEvent(event *core.Event) {
	if !isContainerStartedEvent(event) {
		return
	}
	pod, err := c.k8sFactory.Core().V1().Pods().Lister().Pods(event.InvolvedObject.Namespace).Get(event.InvolvedObject.Name)
	if err != nil {
		glog.Errorf("Failed to retrieve pod %s/%s, due to: %v", event.InvolvedObject.Namespace, event.InvolvedObject.Name, err)
		return
	}
	c.evaluatePodStatus(pod)
}

func (c *Controller) evaluatePodStatus(pod *core.Pod) {
	// Look for OOMKilled containers
	for _, s := range pod.Status.ContainerStatuses {
		if s.LastTerminationState.Terminated == nil || s.LastTerminationState.Terminated.Reason != "OOMKilled" {
			ProcessedContainerUpdates.WithLabelValues("not_oomkilled").Inc()
			continue
		}

		if s.LastTerminationState.Terminated.FinishedAt.Time.Before(c.startTime) {
			glog.V(1).Infof("The container '%s' in '%s/%s' was terminated before this controller started", s.Name, pod.Namespace, pod.Name)
			ProcessedContainerUpdates.WithLabelValues("oomkilled_termination_too_old").Inc()
			continue
		}

		c.recorder.Eventf(pod, core.EventTypeWarning, "PreviousContainerWasOOMKilled", "The previous instance of the container '%s' (%s) was OOMKilled", s.Name, s.ContainerID)
		ProcessedContainerUpdates.WithLabelValues("oomkilled_event_sent").Inc()
	}
}
