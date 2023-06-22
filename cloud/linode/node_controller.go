package linode

import (
	"time"

	"github.com/appscode/go/wait"
	v1 "k8s.io/api/core/v1"
	v1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const nodeControllerRetryInterval = time.Minute * 1

type nodeController struct {
	//nodePoolLabels *nodePoolLabels
	//nodePoolTaints *nodePoolTaints
	informer v1informers.NodeInformer

	queue workqueue.DelayingInterface
}

func newNodeController(
	// We might not pass a nodePoolLabels, nodePoolTaints into the controller, but rather fetch them in the controller at an interval.
	// nodePoolLabels *nplabels, nodePoolTaints *nptaints,
	informer v1informers.NodeInformer) *nodeController {
	return &nodeController{
		//nodePoolLabels: nplabels,
		//nodePoolTaints: nptaints,
		informer: informer,
		queue:    workqueue.NewDelayingQueue(),
	}
}

func (n *nodeController) Run(stopCh <-chan struct{}) {
	n.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Unlike the service controller, we don't want to delete anything.
		// DeleteFunc: func(obj interface{}) {
		// },
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			// oldNode
			_, ok := oldObj.(*v1.Node)
			if !ok {
				return
			}
			newNode, ok := newObj.(*v1.Node)
			if !ok {
				return
			}

			// * needs more checks on node objects, comparison of oldNode with newNode.
			klog.Infof("NodeController will handle update of node %s", newNode.Name)
			n.queue.Add(newObj)
		},
		AddFunc: func(obj interface{}) {
			node, ok := obj.(*v1.Node)
			if !ok {
				return
			}

			// * needs more checks on node
			klog.Infof("NodeController will handle addition of node %s", node.Name)
			n.queue.Add(obj)
		},
	})

	go wait.Until(n.worker, time.Second, stopCh)
	// * add another worker that keeps our nodepoollabel/nodepooltaints cached?
	n.informer.Informer().Run(stopCh)
}

// worker thread updates the queued nodes with the labels and taints contained in the node's associated nodePool spec.
func (n *nodeController) worker() {
	for n.processNextNode() {
	}
}

func (n *nodeController) processNextNode() bool {
	key, quit := n.queue.Get()
	if quit {
		return false
	}
	defer n.queue.Done(key)

	node, ok := key.(*v1.Node)
	if !ok {
		klog.Errorf("expected dequeued key to be of type *v1.Node but got %T", node)
		return true
	}

	err := n.handleNodeUpdate(node)
	if err != nil {
		// * figure out whether to retry or give up
	}
	return true

}

func (n *nodeController) handleNodeUpdate(node *v1.Node) error {
	klog.Infof("NodeController handling label and taint update for node %s", node.Name)
	// read node label "lke.linode.com/pool-id": "555"
	// * we don't know the cluster name from the node through the label, but we do through the name of the node (in all cases?)
	//
	// get pool via pool-id
	//
	// check for labels, and taints, and update the node
	//
	return nil
}
