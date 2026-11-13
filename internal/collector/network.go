package collector

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"
)

type NetworkCollector struct {
	clientset *kubernetes.Clientset
	namespace string
}

func NewNetworkCollector(clientset *kubernetes.Clientset, namespace string) *NetworkCollector {
	return &NetworkCollector{
		clientset: clientset,
		namespace: namespace,
	}
}

type ServiceInfo struct {
	Name            string
	Namespace       string
	Spec            corev1.ServiceSpec
	Status          corev1.ServiceStatus
	Endpoints       *EndpointsInfo
	Selector        map[string]string
	ClusterIP       string
	ClusterIPs      []string
	Type            corev1.ServiceType
	Ports           []corev1.ServicePort
}

type EndpointsInfo struct {
	Name       string
	Namespace  string
	Subsets    []corev1.EndpointSubset
	Addresses  []corev1.EndpointAddress
	NotReadyAddresses []corev1.EndpointAddress
}

type IngressInfo struct {
	Name        string
	Namespace   string
	Spec        networkingv1.IngressSpec
	Status      networkingv1.IngressStatus
	Rules       []networkingv1.IngressRule
	TLS         []networkingv1.IngressTLS
	LoadBalancer *networkingv1.IngressLoadBalancerStatus
}

type NetworkPolicyInfo struct {
	Name       string
	Namespace  string
	Spec       networkingv1.NetworkPolicySpec
	PodSelector map[string]string
	PolicyTypes []networkingv1.PolicyType
	Ingress    []networkingv1.NetworkPolicyIngressRule
	Egress     []networkingv1.NetworkPolicyEgressRule
}

func (c *NetworkCollector) GetService(ctx context.Context, serviceName string) (*ServiceInfo, error) {
	svc, err := c.clientset.CoreV1().Services(c.namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s: %w", serviceName, err)
	}

	info := &ServiceInfo{
		Name:       svc.Name,
		Namespace:  svc.Namespace,
		Spec:       svc.Spec,
		Status:     svc.Status,
		Selector:   svc.Spec.Selector,
		ClusterIP:  svc.Spec.ClusterIP,
		ClusterIPs: svc.Spec.ClusterIPs,
		Type:       svc.Spec.Type,
		Ports:      svc.Spec.Ports,
	}

	endpoints, err := c.clientset.CoreV1().Endpoints(c.namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err == nil {
		info.Endpoints = &EndpointsInfo{
			Name:            endpoints.Name,
			Namespace:       endpoints.Namespace,
			Subsets:         endpoints.Subsets,
			Addresses:       endpoints.Subsets[0].Addresses,
			NotReadyAddresses: endpoints.Subsets[0].NotReadyAddresses,
		}
	}

	return info, nil
}

func (c *NetworkCollector) ListServices(ctx context.Context) ([]ServiceInfo, error) {
	svcs, err := c.clientset.CoreV1().Services(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var infos []ServiceInfo
	for _, svc := range svcs.Items {
		info := ServiceInfo{
			Name:       svc.Name,
			Namespace:  svc.Namespace,
			Spec:       svc.Spec,
			Status:     svc.Status,
			Selector:   svc.Spec.Selector,
			ClusterIP:  svc.Spec.ClusterIP,
			ClusterIPs: svc.Spec.ClusterIPs,
			Type:       svc.Spec.Type,
			Ports:      svc.Spec.Ports,
		}
		infos = append(infos, info)
	}

	return infos, nil
}

func (c *NetworkCollector) GetEndpoints(ctx context.Context, endpointsName string) (*EndpointsInfo, error) {
	ep, err := c.clientset.CoreV1().Endpoints(c.namespace).Get(ctx, endpointsName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints %s: %w", endpointsName, err)
	}

	var addresses []corev1.EndpointAddress
	var notReadyAddresses []corev1.EndpointAddress
	for _, subset := range ep.Subsets {
		addresses = append(addresses, subset.Addresses...)
		notReadyAddresses = append(notReadyAddresses, subset.NotReadyAddresses...)
	}

	return &EndpointsInfo{
		Name:              ep.Name,
		Namespace:         ep.Namespace,
		Subsets:           ep.Subsets,
		Addresses:         addresses,
		NotReadyAddresses: notReadyAddresses,
	}, nil
}

func (c *NetworkCollector) GetIngress(ctx context.Context, ingressName string) (*IngressInfo, error) {
	ing, err := c.clientset.NetworkingV1().Ingresses(c.namespace).Get(ctx, ingressName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress %s: %w", ingressName, err)
	}

	return &IngressInfo{
		Name:       ing.Name,
		Namespace:  ing.Namespace,
		Spec:       ing.Spec,
		Status:     ing.Status,
		Rules:      ing.Spec.Rules,
		TLS:        ing.Spec.TLS,
		LoadBalancer: &ing.Status.LoadBalancer,
	}, nil
}

func (c *NetworkCollector) ListIngresses(ctx context.Context) ([]IngressInfo, error) {
	ings, err := c.clientset.NetworkingV1().Ingresses(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var infos []IngressInfo
	for _, ing := range ings.Items {
		info := IngressInfo{
			Name:       ing.Name,
			Namespace:  ing.Namespace,
			Spec:       ing.Spec,
			Status:     ing.Status,
			Rules:      ing.Spec.Rules,
			TLS:        ing.Spec.TLS,
			LoadBalancer: &ing.Status.LoadBalancer,
		}
		infos = append(infos, info)
	}

	return infos, nil
}

func (c *NetworkCollector) GetNetworkPolicy(ctx context.Context, policyName string) (*NetworkPolicyInfo, error) {
	np, err := c.clientset.NetworkingV1().NetworkPolicies(c.namespace).Get(ctx, policyName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get network policy %s: %w", policyName, err)
	}

	return &NetworkPolicyInfo{
		Name:        np.Name,
		Namespace:   np.Namespace,
		Spec:        np.Spec,
		PodSelector: np.Spec.PodSelector.MatchLabels,
		PolicyTypes: np.Spec.PolicyTypes,
		Ingress:     np.Spec.Ingress,
		Egress:      np.Spec.Egress,
	}, nil
}

func (c *NetworkCollector) ListNetworkPolicies(ctx context.Context) ([]NetworkPolicyInfo, error) {
	nps, err := c.clientset.NetworkingV1().NetworkPolicies(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var infos []NetworkPolicyInfo
	for _, np := range nps.Items {
		info := NetworkPolicyInfo{
			Name:        np.Name,
			Namespace:   np.Namespace,
			Spec:        np.Spec,
			PodSelector: np.Spec.PodSelector.MatchLabels,
			PolicyTypes: np.Spec.PolicyTypes,
			Ingress:     np.Spec.Ingress,
			Egress:      np.Spec.Egress,
		}
		infos = append(infos, info)
	}

	return infos, nil
}

func (c *NetworkCollector) GetServicesForPod(ctx context.Context, podLabels map[string]string) ([]ServiceInfo, error) {
	svcs, err := c.ListServices(ctx)
	if err != nil {
		return nil, err
	}

	var matching []ServiceInfo
	for _, svc := range svcs {
		if c.labelsMatch(svc.Selector, podLabels) {
			matching = append(matching, svc)
		}
	}

	return matching, nil
}

func (c *NetworkCollector) GetNetworkPoliciesForPod(ctx context.Context, podLabels map[string]string) ([]NetworkPolicyInfo, error) {
	nps, err := c.ListNetworkPolicies(ctx)
	if err != nil {
		return nil, err
	}

	var matching []NetworkPolicyInfo
	for _, np := range nps {
		if c.labelsMatch(np.PodSelector, podLabels) {
			matching = append(matching, np)
		}
	}

	return matching, nil
}

func (c *NetworkCollector) labelsMatch(selector, labels map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for key, value := range selector {
		if labels[key] != value {
			return false
		}
	}
	return true
}

func (c *NetworkCollector) HasEndpoints(svc *ServiceInfo) bool {
	return svc.Endpoints != nil && len(svc.Endpoints.Addresses) > 0
}

func (c *NetworkCollector) GetReadyEndpointsCount(svc *ServiceInfo) int {
	if svc.Endpoints == nil {
		return 0
	}
	return len(svc.Endpoints.Addresses)
}

func (c *NetworkCollector) GetNotReadyEndpointsCount(svc *ServiceInfo) int {
	if svc.Endpoints == nil {
		return 0
	}
	return len(svc.Endpoints.NotReadyAddresses)
}