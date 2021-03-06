package cmd

import (
	"fmt"

	"github.com/devonmoss/ckube/util"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"text/tabwriter"
	"os"
)

// nodesCmd represents the nodes command
var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Aliases: []string{"node"},
	Short: "Lists pods grouped by the node",
	Long: `Lists pods grouped by node`,
	Run: func(cmd *cobra.Command, args []string) {
		printNodeView()
	},
}

func printNodeView() {
	nodeMap := nodeMap()
	for node, pods := range nodeMap {
		fmt.Println(node)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.StripEscape)
		headerLine := fmt.Sprintf("\t%v\t%v\t%v\t%v\t%v\t", "NAME", "READY", "STATUS", "RESTARTS", "AGE")
		fmt.Fprintln(w, headerLine)
		for _, pod := range pods {
			ps := NewPodStatus(pod)
			statusLine := fmt.Sprintf("\t%v\t%v/%v\t%v\t%v\t%v\t", pod.Name, ps.ready, ps.total, pod.Status.Phase, ps.restarts, pod.Status.StartTime)
			fmt.Fprintln(w, statusLine)
		}
		w.Flush()
		fmt.Println()
	}
}

type PodStatus struct {
	total    int
	ready    int
	restarts int32
}

func NewPodStatus(pod v1.Pod) PodStatus {
	total := len(pod.Status.ContainerStatuses)
	var ready int
	var restarts int32
	for _, c := range pod.Status.ContainerStatuses {
		if c.Ready {
			ready++
		}
		restarts += c.RestartCount
	}
	return PodStatus{total: total, ready: ready, restarts: restarts}
}

func nodeMap() map[string][]v1.Pod {
	clientset := util.GetClientset(kubeconfig)

	podList, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(fmt.Errorf("error listing pods: %v", err))
	}
	nodeMap := make(map[string][]v1.Pod)
	for _, pod := range podList.Items {
		if _, ok := nodeMap[pod.Spec.NodeName]; ok {
			nodeMap[pod.Spec.NodeName] = append(nodeMap[pod.Spec.NodeName], pod)
		} else {
			nodeMap[pod.Spec.NodeName] = []v1.Pod{pod}
		}
	}
	return nodeMap
}

func init() {
	RootCmd.AddCommand(nodesCmd)
}
