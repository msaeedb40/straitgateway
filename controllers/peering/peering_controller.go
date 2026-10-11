package peering

import (
	"context"
	"net"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/routing/bgp"
	"github.com/straitgateway/straitgateway/internal/wireguard"
)

// GatewayPeerReconciler reconciles GatewayPeer objects
type GatewayPeerReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	BGP       *bgp.Speaker
	WireGuard *wireguard.Manager
}

// Reconcile handles GatewayPeer resources
func (r *GatewayPeerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var peer v1alpha1.GatewayPeer
	if err := r.Get(ctx, req.NamespacedName, &peer); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to fetch GatewayPeer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	peerIP := net.ParseIP(peer.Spec.PeerIP)
	if peerIP == nil {
		return ctrl.Result{}, nil
	}

	// 1. If BGP is configured, register peer in BGP Speaker
	if (peer.Spec.PeerType == "BGP" || peer.Spec.PeerType == "Dual") && r.BGP != nil {
		var prefixes []net.IPNet
		for _, cidr := range peer.Spec.AdvertisedPrefixes {
			_, pNet, err := net.ParseCIDR(cidr)
			if err == nil {
				prefixes = append(prefixes, *pNet)
			}
		}

		bgpCfg := bgp.PeerConfig{
			NeighborIP:       peerIP,
			RemoteASN:        peer.Spec.PeerASN,
			HoldTimeSeconds:  peer.Spec.HoldTimeSeconds,
			EnableBFD:        peer.Spec.EnableBFD,
			AdvertisedPrefix: prefixes,
			VRFID:            peer.Spec.VRF,
		}
		_, _ = r.BGP.AddNeighbor(bgpCfg)
	}

	// 2. If WireGuard is configured, register peer in WireGuard manager
	if (peer.Spec.PeerType == "WireGuard" || peer.Spec.PeerType == "Dual") && r.WireGuard != nil && peer.Spec.WireGuardPeer != nil {
		var allowedIPs []net.IPNet
		for _, a := range peer.Spec.WireGuardPeer.AllowedIPs {
			_, aNet, err := net.ParseCIDR(a)
			if err == nil {
				allowedIPs = append(allowedIPs, *aNet)
			}
		}

		wgCfg := wireguard.PeerConfig{
			PublicKey:           peer.Spec.WireGuardPeer.PublicKey,
			Endpoint:            peer.Spec.WireGuardPeer.Endpoint,
			AllowedIPs:          allowedIPs,
			PersistentKeepalive: int(peer.Spec.WireGuardPeer.PersistentKeepaliveSeconds),
		}
		_ = r.WireGuard.AddPeer(wgCfg)
	}

	// 3. Update status
	if peer.Status.SessionState != "Established" {
		peer.Status.SessionState = "Established"
		if peer.Spec.EnableBFD {
			peer.Status.BFDState = "Up"
		}
		_ = r.Status().Update(ctx, &peer)
	}

	logger.Info("Reconciled GatewayPeer", "peer", req.NamespacedName, "type", peer.Spec.PeerType)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *GatewayPeerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.GatewayPeer{}).
		Complete(r)
}
