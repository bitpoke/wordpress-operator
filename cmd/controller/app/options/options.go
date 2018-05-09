package options

import (
	"time"

	"github.com/spf13/pflag"
)

const (
	defaultKubeconfig        = ""
	defaultKubeconfigContext = ""

	defaultRuntimeImage = "quay.io/presslabs/runtime:latest"

	defaultLeaderElect                 = true
	defaultLeaderElectionNamespace     = "kube-system"
	defaultLeaderElectionLeaseDuration = 15 * time.Second
	defaultLeaderElectionRenewDeadline = 10 * time.Second
	defaultLeaderElectionRetryPeriod   = 2 * time.Second
)

type ControllerManagerOptions struct {
	Kubeconfig        string
	KubeconfigContext string
	InstallCRDs       bool

	RuntimeImage string

	LeaderElect                 bool
	LeaderElectionNamespace     string
	LeaderElectionLeaseDuration time.Duration
	LeaderElectionRenewDeadline time.Duration
	LeaderElectionRetryPeriod   time.Duration
}

func NewControllerManagerOptions() *ControllerManagerOptions {
	return &ControllerManagerOptions{
		Kubeconfig:        defaultKubeconfig,
		KubeconfigContext: defaultKubeconfigContext,

		LeaderElect:                 defaultLeaderElect,
		LeaderElectionNamespace:     defaultLeaderElectionNamespace,
		LeaderElectionLeaseDuration: defaultLeaderElectionLeaseDuration,
		LeaderElectionRenewDeadline: defaultLeaderElectionRenewDeadline,
		LeaderElectionRetryPeriod:   defaultLeaderElectionRetryPeriod,
	}
}

func (o *ControllerManagerOptions) Validate() error {
	return nil
}

func (o *ControllerManagerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Kubeconfig, "kubeconfig", defaultKubeconfig, "Path for kubeconfig file")
	fs.StringVar(&o.KubeconfigContext, "context", defaultKubeconfigContext, "Name of the kubeconfig context to use")
	fs.BoolVar(&o.InstallCRDs, "install-crds", true, "If true, the "+
		"wordpress-controller will install the Custom Resource Definitions "+
		"for Wordpress. The controller needs the appropriate permissions in "+
		"order to do so.")

	fs.StringVar(&o.RuntimeImage, "runtime", defaultRuntimeImage, "Runtime image to use by default")

	fs.BoolVar(&o.LeaderElect, "leader-elect", true, ""+
		"If true, wordpress-controller will perform leader election between instances to ensure no more "+
		"than one instance of wordpress-controller operates at a time")
	fs.StringVar(&o.LeaderElectionNamespace, "leader-election-namespace", defaultLeaderElectionNamespace, ""+
		"Namespace used to perform leader election. Only used if leader election is enabled")
	fs.DurationVar(&o.LeaderElectionLeaseDuration, "leader-election-lease-duration", defaultLeaderElectionLeaseDuration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	fs.DurationVar(&o.LeaderElectionRenewDeadline, "leader-election-renew-deadline", defaultLeaderElectionRenewDeadline, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	fs.DurationVar(&o.LeaderElectionRetryPeriod, "leader-election-retry-period", defaultLeaderElectionRetryPeriod, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
}
