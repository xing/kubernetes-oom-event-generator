package controller

import (
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	corev1_api "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"github.com/xing/kubernetes-oom-event-generator/src/util"
)

const (
	// informerSyncMinute defines how often the cache is synced from Kubernetes
	informerSyncMinute = 3
)

// Controller is a controller that listens on Pod changes and create Kubernetes Events
// when a container reports it was previously killed
type Controller struct {
	Stop       chan struct{}
	k8sClient  kubernetes.Interface
	k8sFactory informers.SharedInformerFactory
	podState   map[string]map[string]int32
	recorder   record.EventRecorder
	startTime  time.Time
	stopCh     chan struct{}
	updateCh   chan *corev1_api.Pod
}

// NewController returns an instance of the Controller
func NewController(stop chan struct{}) *Controller {
	k8sClient := util.Clientset()
	k8sFactory := informers.NewSharedInformerFactory(k8sClient, time.Minute*informerSyncMinute)

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})

	controller := &Controller{
		k8sClient:  k8sClient,
		stopCh:     make(chan struct{}),
		Stop:       stop,
		k8sFactory: k8sFactory,
		updateCh:   make(chan *corev1_api.Pod),
		recorder:   eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "oom-event-generator"}),
		podState:   make(map[string]map[string]int32),
		startTime:  time.Now(),
	}

	podInformer := informers.SharedInformerFactory(k8sFactory).Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: controller.podUpdated,
		DeleteFunc: controller.podDeleted,
	})
	return controller
}

func (c *Controller) podUpdated(oldObj, newObj interface{}) {
	c.updateCh <- newObj.(*corev1_api.Pod)
}

func (c *Controller) podDeleted(obj interface{}) {
	c.deletePodState(string(obj.(*corev1_api.Pod).UID))
}

// Run is the main loop that processes Kubernetes Pod changes
func (c *Controller) Run() error {
	c.k8sFactory.Start(c.stopCh)

	for {
		select {
		case pod := <-c.updateCh:
			c.evaluatePodStatus(pod)
		case <-c.Stop:
			glog.Info("Stopping")
			return nil
		}
	}
}

func (c *Controller) evaluatePodStatus(pod *corev1_api.Pod) {
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

		if c.setRestartCount(string(pod.UID), s.ContainerID, s.RestartCount) {
			glog.V(1).Infof("The container '%s' in '%s/%s' was restarted for the %d time", s.Name, pod.Namespace, pod.Name, s.RestartCount)
			c.recorder.Eventf(pod, v1.EventTypeWarning, "PreviousPodWasOOMKilled", "The previous instance of the container '%s' (%s) was OOMKilled", s.Name, s.ContainerID)
			ProcessedContainerUpdates.WithLabelValues("oomkilled_event_sent").Inc()
		} else {
			glog.V(1).Infof("Restart count hasn't changed for '%s' in '%s/%s'", s.Name, pod.Namespace, pod.Name)
			ProcessedContainerUpdates.WithLabelValues("oomkilled_restart_count_unchanged").Inc()
		}
	}
}

// setRestartCount stores the number of restart for each Container of each Pod
func (c *Controller) setRestartCount(podUID, containerID string, restartCount int32) bool {
	pod, ok := c.podState[podUID]

	// If the Pod is not known yet, save its state
	if !ok {
		pod = map[string]int32{}
		c.podState[podUID] = pod
	}

	// if the container is not known yet, or the restartCount has changed, update it
	cachedRestartedCount, ok := pod[containerID]
	if !ok || restartCount > cachedRestartedCount {
		pod[containerID] = restartCount
		return true
	}

	return false
}

// deletePodState removes a Pod from the local state if it exists
func (c *Controller) deletePodState(podUID string) {
	_, ok := c.podState[podUID]
	if ok {
		delete(c.podState, podUID)
	}
}
