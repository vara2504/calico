// Copyright (c) 2016-2024 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conversion

import (
	"fmt"
	"math/bits"
	"sort"
	"strings"

	"github.com/google/uuid"
	apiv3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/projectcalico/api/pkg/lib/numorstring"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	kapiv1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	adminpolicy "sigs.k8s.io/network-policy-api/apis/v1alpha1"

	"github.com/projectcalico/calico/libcalico-go/lib/backend/model"
	cerrors "github.com/projectcalico/calico/libcalico-go/lib/errors"
	"github.com/projectcalico/calico/libcalico-go/lib/names"
	cnet "github.com/projectcalico/calico/libcalico-go/lib/net"
)

var protoTCP = kapiv1.ProtocolTCP

type selectorType int8

const (
	SelectorNamespace selectorType = iota
	SelectorPod
)

type Converter interface {
	WorkloadEndpointConverter
	ParseWorkloadEndpointName(workloadName string) (names.WorkloadEndpointIdentifiers, error)
	NamespaceToProfile(ns *kapiv1.Namespace) (*model.KVPair, error)
	IsValidCalicoWorkloadEndpoint(pod *kapiv1.Pod) bool
	IsReadyCalicoPod(pod *kapiv1.Pod) bool
	IsScheduled(pod *kapiv1.Pod) bool
	IsHostNetworked(pod *kapiv1.Pod) bool
	HasIPAddress(pod *kapiv1.Pod) bool
	StagedKubernetesNetworkPolicyToStagedName(stagedK8sName string) string
	K8sNetworkPolicyToCalico(np *networkingv1.NetworkPolicy) (*model.KVPair, error)
	K8sAdminNetworkPolicyToCalico(anp *adminpolicy.AdminNetworkPolicy) (*model.KVPair, error)
	K8sBaselineAdminNetworkPolicyToCalico(banp *adminpolicy.BaselineAdminNetworkPolicy) (*model.KVPair, error)
	EndpointSliceToKVP(svc *discovery.EndpointSlice) (*model.KVPair, error)
	ServiceToKVP(service *kapiv1.Service) (*model.KVPair, error)
	ProfileNameToNamespace(profileName string) (string, error)
	ServiceAccountToProfile(sa *kapiv1.ServiceAccount) (*model.KVPair, error)
	ProfileNameToServiceAccount(profileName string) (ns, sa string, err error)
	JoinProfileRevisions(nsRev, saRev string) string
	SplitProfileRevision(rev string) (nsRev string, saRev string, err error)
}

type converter struct {
	WorkloadEndpointConverter
}

func NewConverter() Converter {
	return &converter{
		WorkloadEndpointConverter: NewWorkloadEndpointConverter(),
	}
}

// ParseWorkloadName extracts the Node name, Orchestrator, Pod name and endpoint from the
// given WorkloadEndpoint name.
// The expected format for k8s is <node>-k8s-<pod>-<endpoint>
func (c converter) ParseWorkloadEndpointName(workloadName string) (names.WorkloadEndpointIdentifiers, error) {
	return names.ParseWorkloadEndpointName(workloadName)
}

// NamespaceToProfile converts a Namespace to a Calico Profile.  The Profile stores
// labels from the Namespace which are inherited by the WorkloadEndpoints within
// the Profile. This Profile also has the default ingress and egress rules, which are both 'allow'.
func (c converter) NamespaceToProfile(ns *kapiv1.Namespace) (*model.KVPair, error) {
	// Generate the labels to apply to the profile, using a special prefix
	// to indicate that these are the labels from the parent Kubernetes Namespace.
	labels := map[string]string{}
	for k, v := range ns.Labels {
		labels[NamespaceLabelPrefix+k] = v
	}

	// Add a label for the namespace's name. This allows exact namespace matching
	// based on name within the namespaceSelector.
	labels[NamespaceLabelPrefix+NameLabel] = ns.Name

	uid, err := ConvertUID(ns.UID)
	if err != nil {
		return nil, err
	}

	// Create the profile object.
	name := NamespaceProfileNamePrefix + ns.Name
	profile := apiv3.NewProfile()
	profile.ObjectMeta = metav1.ObjectMeta{
		Name:              name,
		CreationTimestamp: ns.CreationTimestamp,
		UID:               uid,
	}
	profile.Spec = apiv3.ProfileSpec{
		Ingress:       []apiv3.Rule{{Action: apiv3.Allow}},
		Egress:        []apiv3.Rule{{Action: apiv3.Allow}},
		LabelsToApply: labels,
	}

	// Embed the profile in a KVPair.
	kvp := model.KVPair{
		Key: model.ResourceKey{
			Name: name,
			Kind: apiv3.KindProfile,
		},
		Value:    profile,
		Revision: c.JoinProfileRevisions(ns.ResourceVersion, ""),
	}
	return &kvp, nil
}

// IsValidCalicoWorkloadEndpoint returns true if the pod should be shown as a workloadEndpoint
// in the Calico API and false otherwise.  Note: since we completely ignore notifications for
// invalid Pods, it is important that pods can only transition from not-valid to valid and not
// the other way.  If they transition from valid to invalid, we'll fail to emit a deletion
// event in the watcher.
func (c converter) IsValidCalicoWorkloadEndpoint(pod *kapiv1.Pod) bool {
	if c.IsHostNetworked(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is host networked.")
		return false
	} else if !c.IsScheduled(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is not scheduled.")
		return false
	}
	return true
}

// IsReadyCalicoPod returns true if the pod is a valid Calico WorkloadEndpoint and has
// an IP address assigned (i.e. it's ready for Calico networking).
func (c converter) IsReadyCalicoPod(pod *kapiv1.Pod) bool {
	if !c.IsValidCalicoWorkloadEndpoint(pod) {
		return false
	} else if !c.HasIPAddress(pod) {
		log.WithField("pod", pod.Name).Debug("Pod does not have an IP address.")
		return false
	}
	return true
}

const (
	// Completed is documented but doesn't seem to be in the API, it should be safe to include.
	// Maybe it's in an older version of the API?
	podCompleted kapiv1.PodPhase = "Completed"
)

func IsFinished(pod *kapiv1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		// Pod is being deleted but it may still be in its termination grace period.  If Calico CNI
		// was used, then we use AnnotationPodIP to signal the moment that the pod actually loses its
		// IP by setting the annotation to "".  (Otherwise, just fall back on the status of the pod.)
		if ip, ok := pod.Annotations[AnnotationPodIP]; ok && ip == "" {
			// AnnotationPodIP is explicitly set to empty string, Calico CNI has removed the network
			// from the pod.
			log.Debug("Pod is being deleted and IPs have been removed by Calico CNI.")
			return true
		} else if ips, ok := pod.Annotations[AnnotationAWSPodIPs]; ok && ips == "" {
			// AnnotationAWSPodIPs is explicitly set to empty string, AWS CNI has removed the network
			// from the pod.
			log.Debug("Pod is being deleted and IPs have been removed by AWS CNI.")
			return true
		}
	}
	switch pod.Status.Phase {
	case kapiv1.PodFailed, kapiv1.PodSucceeded, podCompleted:
		log.Debug("Pod phase is failed/succeeded/completed.")
		return true
	}
	return false
}

func (c converter) IsScheduled(pod *kapiv1.Pod) bool {
	return pod.Spec.NodeName != ""
}

func (c converter) IsHostNetworked(pod *kapiv1.Pod) bool {
	return pod.Spec.HostNetwork
}

func (c converter) HasIPAddress(pod *kapiv1.Pod) bool {
	return pod.Status.PodIP != "" || pod.Annotations[AnnotationPodIP] != "" || pod.Annotations[AnnotationAWSPodIPs] != ""
	// Note: we don't need to check PodIPs and AnnotationPodIPs here, because those cannot be
	// non-empty if the corresponding singular field is empty.
}

// getPodIPs extracts the IP addresses from a Kubernetes Pod.  We support a single IPv4 address
// and/or a single IPv6.  getPodIPs loads the IPs either from the PodIPs and PodIP field, if
// present, or the calico podIP annotation.
func getPodIPs(pod *kapiv1.Pod) ([]*cnet.IPNet, error) {
	logc := log.WithFields(log.Fields{"pod": pod.Name, "namespace": pod.Namespace})
	var podIPs []string
	if ips := pod.Status.PodIPs; len(ips) != 0 {
		logc.WithField("ips", ips).Debug("PodIPs field filled in")
		for _, ip := range ips {
			podIPs = append(podIPs, ip.IP)
		}
	} else if ip := pod.Status.PodIP; ip != "" {
		logc.WithField("ip", ip).Debug("PodIP field filled in")
		podIPs = append(podIPs, ip)
	} else if ips := pod.Annotations[AnnotationPodIPs]; ips != "" {
		logc.WithField("ips", ips).Debug("No PodStatus IPs, use Calico plural annotation")
		podIPs = append(podIPs, strings.Split(ips, ",")...)
	} else if ip := pod.Annotations[AnnotationPodIP]; ip != "" {
		logc.WithField("ip", ip).Debug("No PodStatus IPs, use Calico singular annotation")
		podIPs = append(podIPs, ip)
	} else if ips := pod.Annotations[AnnotationAWSPodIPs]; ips != "" {
		logc.WithField("ips", ips).Debug("No PodStatus IPs, use AWS VPC annotation")
		podIPs = append(podIPs, strings.Split(ips, ",")...)
	} else {
		logc.Debug("Pod has no IP")
		return nil, nil
	}
	var podIPNets []*cnet.IPNet
	for _, ip := range podIPs {
		_, ipNet, err := cnet.ParseCIDROrIP(ip)
		if err != nil {
			logc.WithFields(log.Fields{"ip": ip}).WithError(err).Error("Failed to parse pod IP")
			return nil, err
		}
		podIPNets = append(podIPNets, ipNet)
	}
	return podIPNets, nil
}

// StagedKubernetesNetworkPolicyToStagedName converts a StagedKubernetesNetworkPolicy name into a StagedNetworkPolicy name
func (c converter) StagedKubernetesNetworkPolicyToStagedName(stagedK8sName string) string {
	return names.K8sNetworkPolicyNamePrefix + stagedK8sName
}

// EndpointSliceToKVP converts a k8s EndpointSlice to a model.KVPair.
func (c converter) EndpointSliceToKVP(slice *discovery.EndpointSlice) (*model.KVPair, error) {
	return &model.KVPair{
		Key: model.ResourceKey{
			Name:      slice.Name,
			Namespace: slice.Namespace,
			Kind:      model.KindKubernetesEndpointSlice,
		},
		Value:    slice.DeepCopy(),
		Revision: slice.ResourceVersion,
	}, nil
}

func (c converter) ServiceToKVP(service *kapiv1.Service) (*model.KVPair, error) {
	return &model.KVPair{
		Key: model.ResourceKey{
			Name:      service.Name,
			Namespace: service.Namespace,
			Kind:      model.KindKubernetesService,
		},
		Value:    service.DeepCopy(),
		Revision: service.ResourceVersion,
	}, nil
}

// K8sAdminNetworkPolicyToCalico converts a k8s AdminNetworkPolicy to a model.KVPair.
func (c converter) K8sAdminNetworkPolicyToCalico(anp *adminpolicy.AdminNetworkPolicy) (*model.KVPair, error) {
	// Pull out important fields.
	policyName := names.K8sAdminNetworkPolicyNamePrefix + anp.Name
	order := float64(anp.Spec.Priority)
	errorTracker := cerrors.ErrorAdminPolicyConversion{PolicyName: anp.Name}

	// Generate the ingress rules list.
	var ingressRules []apiv3.Rule
	for _, r := range anp.Spec.Ingress {
		rules, err := k8sANPIngressRuleToCalico(r)
		if err != nil {
			log.WithError(err).Warn("dropping k8s rule that couldn't be converted.")
			// Add rule to conversion error slice
			errorTracker.BadIngressRule(&r, fmt.Sprintf("k8s rule couldn't be converted: %s", err))
			failClosedRule := k8sANPHandleFailedRules(r.Action)
			if failClosedRule != nil {
				ingressRules = append(ingressRules, *failClosedRule)
			}
		} else {
			ingressRules = append(ingressRules, rules...)
		}
	}

	// Generate the egress rules list.
	var egressRules []apiv3.Rule
	for _, r := range anp.Spec.Egress {
		rules, err := k8sANPEgressRuleToCalico(r)
		if err != nil {
			log.WithError(err).Warn("dropping k8s rule that couldn't be converted.")
			// Add rule to conversion error slice
			errorTracker.BadEgressRule(&r, fmt.Sprintf("k8s rule couldn't be converted: %s", err))
			failClosedRule := k8sANPHandleFailedRules(r.Action)
			if failClosedRule != nil {
				egressRules = append(egressRules, *failClosedRule)
			}
		} else {
			egressRules = append(egressRules, rules...)
		}
	}

	// Either Namespaces or Pods is set. Use one of them to populate the selectors.
	var nsSelector, podSelector string
	if anp.Spec.Subject.Namespaces != nil {
		nsSelector = k8sSelectorToCalico(anp.Spec.Subject.Namespaces, SelectorNamespace)
		// Make sure projectcalico.org/orchestrator == 'k8s' label is added to exclude heps.
		podSelector = k8sSelectorToCalico(nil, SelectorPod)
	} else {
		nsSelector = k8sSelectorToCalico(&anp.Spec.Subject.Pods.NamespaceSelector, SelectorNamespace)
		podSelector = k8sSelectorToCalico(&anp.Spec.Subject.Pods.PodSelector, SelectorPod)
	}

	var uid types.UID
	var err error
	if anp.UID != "" {
		uid, err = ConvertUID(anp.UID)
		if err != nil {
			return nil, err
		}
	}

	gnp := apiv3.NewGlobalNetworkPolicy()
	gnp.ObjectMeta = metav1.ObjectMeta{
		Name:              policyName,
		CreationTimestamp: anp.CreationTimestamp,
		UID:               uid,
		ResourceVersion:   anp.ResourceVersion,
	}
	gnp.Spec = apiv3.GlobalNetworkPolicySpec{
		Tier:              names.AdminNetworkPolicyTierName,
		Order:             &order,
		NamespaceSelector: nsSelector,
		Selector:          podSelector,
		Ingress:           ingressRules,
		Egress:            egressRules,
		Types:             c.calculateANPPolicyTypes(ingressRules, egressRules),
	}

	// Build the KVPair.
	kvp := &model.KVPair{
		Key: model.ResourceKey{
			Name: policyName,
			Kind: apiv3.KindGlobalNetworkPolicy,
		},
		Value:    gnp,
		Revision: anp.ResourceVersion,
	}

	// Return the KVPair with conversion errors if applicable
	return kvp, errorTracker.GetError()
}

func k8sANPHandleFailedRules(action adminpolicy.AdminNetworkPolicyRuleAction) *apiv3.Rule {
	if action == adminpolicy.AdminNetworkPolicyRuleActionDeny ||
		action == adminpolicy.AdminNetworkPolicyRuleActionPass {
		logrus.Warn("replacing failed rule with a deny-all one.")
		return &apiv3.Rule{
			Action: apiv3.Deny,
		}
	}
	return nil
}

func k8sANPIngressRuleToCalico(rule adminpolicy.AdminNetworkPolicyIngressRule) ([]apiv3.Rule, error) {
	action, err := K8sAdminNetworkPolicyActionToCalico(rule.Action)
	if err != nil {
		return nil, err
	}
	return combinePortsWithANPIngressPeers(rule.Ports, rule.From, rule.Name, action)
}

func k8sANPEgressRuleToCalico(rule adminpolicy.AdminNetworkPolicyEgressRule) ([]apiv3.Rule, error) {
	action, err := K8sAdminNetworkPolicyActionToCalico(rule.Action)
	if err != nil {
		return nil, err
	}
	return combinePortsWithANPEgressPeers(rule.Ports, rule.To, rule.Name, action)
}

// K8sBaselineAdminNetworkPolicyToCalico converts a k8s BaselineAdminNetworkPolicy to a model.KVPair.
func (c converter) K8sBaselineAdminNetworkPolicyToCalico(anp *adminpolicy.BaselineAdminNetworkPolicy) (*model.KVPair, error) {
	// Pull out important fields.
	policyName := names.K8sBaselineAdminNetworkPolicyNamePrefix + anp.Name
	order := float64(1000)
	errorTracker := cerrors.ErrorAdminPolicyConversion{PolicyName: anp.Name}

	// Generate the ingress rules list.
	var ingressRules []apiv3.Rule
	for _, r := range anp.Spec.Ingress {
		rules, err := k8sBANPIngressRuleToCalico(r)
		if err != nil {
			log.WithError(err).Warn("dropping k8s rule that couldn't be converted.")
			// Add rule to conversion error slice
			errorTracker.BadIngressRule(&r, fmt.Sprintf("k8s rule couldn't be converted: %s", err))
			failClosedRule := k8sBANPHandleFailedRules(r.Action)
			if failClosedRule != nil {
				ingressRules = append(ingressRules, *failClosedRule)
			}
		} else {
			ingressRules = append(ingressRules, rules...)
		}
	}

	// Generate the egress rules list.
	var egressRules []apiv3.Rule
	for _, r := range anp.Spec.Egress {
		rules, err := k8sBANPEgressRuleToCalico(r)
		if err != nil {
			log.WithError(err).Warn("dropping k8s rule that couldn't be converted.")
			// Add rule to conversion error slice
			errorTracker.BadEgressRule(&r, fmt.Sprintf("k8s rule couldn't be converted: %s", err))
			failClosedRule := k8sBANPHandleFailedRules(r.Action)
			if failClosedRule != nil {
				egressRules = append(egressRules, *failClosedRule)
			}
		} else {
			egressRules = append(egressRules, rules...)
		}
	}

	// Either Namespaces or Pods is set. Use one of them to populate the selectors.
	var nsSelector, podSelector string
	if anp.Spec.Subject.Namespaces != nil {
		nsSelector = k8sSelectorToCalico(anp.Spec.Subject.Namespaces, SelectorNamespace)
		// Make sure projectcalico.org/orchestrator == 'k8s' label is added to exclude heps.
		podSelector = k8sSelectorToCalico(nil, SelectorPod)
	} else {
		nsSelector = k8sSelectorToCalico(&anp.Spec.Subject.Pods.NamespaceSelector, SelectorNamespace)
		podSelector = k8sSelectorToCalico(&anp.Spec.Subject.Pods.PodSelector, SelectorPod)
	}

	var uid types.UID
	var err error
	if anp.UID != "" {
		uid, err = ConvertUID(anp.UID)
		if err != nil {
			return nil, err
		}
	}

	gnp := apiv3.NewGlobalNetworkPolicy()
	gnp.ObjectMeta = metav1.ObjectMeta{
		Name:              policyName,
		CreationTimestamp: anp.CreationTimestamp,
		UID:               uid,
		ResourceVersion:   anp.ResourceVersion,
	}
	gnp.Spec = apiv3.GlobalNetworkPolicySpec{
		Tier:              names.BaselineAdminNetworkPolicyTierName,
		Order:             &order,
		NamespaceSelector: nsSelector,
		Selector:          podSelector,
		Ingress:           ingressRules,
		Egress:            egressRules,
		Types:             c.calculateANPPolicyTypes(ingressRules, egressRules),
	}

	// Build the KVPair.
	kvp := &model.KVPair{
		Key: model.ResourceKey{
			Name: policyName,
			Kind: apiv3.KindGlobalNetworkPolicy,
		},
		Value:    gnp,
		Revision: anp.ResourceVersion,
	}

	// Return the KVPair with conversion errors if applicable
	return kvp, errorTracker.GetError()
}

func (c converter) calculateANPPolicyTypes(ingressRules []apiv3.Rule, egressRules []apiv3.Rule) []apiv3.PolicyType {
	// Calculate Types setting. The ANP Tiers are default-Pass so the only
	// reason to enable a policy type is if we have rules.
	var policyTypes []apiv3.PolicyType
	if len(ingressRules) != 0 {
		policyTypes = append(policyTypes, apiv3.PolicyTypeIngress)
	}
	if len(egressRules) != 0 {
		policyTypes = append(policyTypes, apiv3.PolicyTypeEgress)
	}
	return policyTypes
}

func k8sBANPHandleFailedRules(action adminpolicy.BaselineAdminNetworkPolicyRuleAction) *apiv3.Rule {
	if action == adminpolicy.BaselineAdminNetworkPolicyRuleActionDeny {
		logrus.Warn("replacing failed rule with a deny-all one.")
		return &apiv3.Rule{
			Action: apiv3.Deny,
		}
	}
	return nil
}

func k8sBANPIngressRuleToCalico(rule adminpolicy.BaselineAdminNetworkPolicyIngressRule) (rules []apiv3.Rule, err error) {
	action, err := K8sBaselineAdminNetworkPolicyActionToCalico(rule.Action)
	if err != nil {
		return nil, err
	}
	return combinePortsWithANPIngressPeers(rule.Ports, rule.From, rule.Name, action)
}

func combinePortsWithANPIngressPeers(
	anpPorts *[]adminpolicy.AdminNetworkPolicyPort,
	anpPeers []adminpolicy.AdminNetworkPolicyIngressPeer,
	ruleName string,
	action apiv3.Action,
) (rules []apiv3.Rule, err error) {
	protocolPorts, sortedProtocols, err := unpackANPPorts(anpPorts)
	if err != nil {
		return nil, err
	}

	// Combine destinations with sources to generate rules. We generate one rule per protocol,
	// with each rule containing all the allowed ports.
	for _, protocolStr := range sortedProtocols {
		calicoPorts := protocolPorts[protocolStr]
		calicoPorts = SimplifyPorts(calicoPorts)

		var protocol *numorstring.Protocol
		if protocolStr != "" {
			p := numorstring.ProtocolFromString(protocolStr)
			protocol = &p
		}

		// Based on specifications at least one Peer is set.
		var selector, nsSelector string
		for _, peer := range anpPeers {
			var found bool
			if peer.Namespaces != nil {
				selector = ""
				nsSelector = k8sSelectorToCalico(peer.Namespaces, SelectorNamespace)
				found = true
			}
			if peer.Pods != nil {
				selector = k8sSelectorToCalico(&peer.Pods.PodSelector, SelectorPod)
				nsSelector = k8sSelectorToCalico(&peer.Pods.NamespaceSelector, SelectorNamespace)
				found = true
			}
			if !found {
				return nil, fmt.Errorf("none of supported fields in 'From' is set.")
			}

			// Build inbound rule and append to list.
			rules = append(rules, apiv3.Rule{
				Metadata: k8sAdminNetworkPolicyToCalicoMetadata(ruleName),
				Action:   action,
				Protocol: protocol,
				Source: apiv3.EntityRule{
					Selector:          selector,
					NamespaceSelector: nsSelector,
				},
				Destination: apiv3.EntityRule{
					Ports: calicoPorts,
				},
			})
		}
	}
	return rules, nil
}

func unpackANPPorts(k8sPorts *[]adminpolicy.AdminNetworkPolicyPort) (map[string][]numorstring.Port, []string, error) {
	// If there are no ports, represent that as zero struct.
	ports := []adminpolicy.AdminNetworkPolicyPort{{}}
	if k8sPorts != nil && len(*k8sPorts) != 0 {
		ports = *k8sPorts
	}

	protocolPorts := map[string][]numorstring.Port{}

	for _, port := range ports {
		protocol, calicoPort, err := k8sAdminPolicyPortToCalicoFields(&port)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse k8s port: %s", err)
		}

		if protocol == nil && calicoPort == nil {
			// If nil, no ports were specified, or an empty port struct was provided, which we translate to allowing all.
			// We want to use a nil protocol and a nil list of ports, which will allow any destination (for ingress).
			// Given we're gonna allow all, we may as well break here and keep only this rule
			protocolPorts = map[string][]numorstring.Port{"": nil}
			break
		}

		pStr := protocol.String()
		// treat nil as 'all ports'
		if calicoPort == nil {
			protocolPorts[pStr] = nil
		} else if _, ok := protocolPorts[pStr]; !ok || len(protocolPorts[pStr]) > 0 {
			// don't overwrite a nil (allow all ports) if present; if no ports yet for this protocol
			// or 1+ ports which aren't 'all ports', then add the present ports
			protocolPorts[pStr] = append(protocolPorts[pStr], *calicoPort)
		}
	}

	protocols := make([]string, 0, len(protocolPorts))
	for k := range protocolPorts {
		protocols = append(protocols, k)
	}
	// Ensure deterministic output
	sort.Strings(protocols)
	return protocolPorts, protocols, nil
}

func k8sBANPEgressRuleToCalico(rule adminpolicy.BaselineAdminNetworkPolicyEgressRule) ([]apiv3.Rule, error) {
	action, err := K8sBaselineAdminNetworkPolicyActionToCalico(rule.Action)
	if err != nil {
		return nil, err
	}
	return combinePortsWithANPEgressPeers(rule.Ports, rule.To, rule.Name, action)
}

func combinePortsWithANPEgressPeers(
	rulePorts *[]adminpolicy.AdminNetworkPolicyPort,
	rulePeers []adminpolicy.AdminNetworkPolicyEgressPeer,
	ruleName string,
	action apiv3.Action,
) ([]apiv3.Rule, error) {
	var rules []apiv3.Rule

	// If there no ports, represent that as zero struct.
	ports := []adminpolicy.AdminNetworkPolicyPort{{}}
	if rulePorts != nil && len(*rulePorts) != 0 {
		ports = *rulePorts
	}

	protocolPorts := map[string][]numorstring.Port{}

	for _, port := range ports {
		protocol, calicoPort, err := k8sAdminPolicyPortToCalicoFields(&port)
		if err != nil {
			return nil, fmt.Errorf("failed to parse k8s port: %s", err)
		}

		if protocol == nil && calicoPort == nil {
			// If nil, no ports were specified, or an empty port struct was provided, which we translate to allowing all.
			// We want to use a nil protocol and a nil list of ports, which will allow any destination (for ingress).
			// Given we're gonna allow all, we may as well break here and keep only this rule
			protocolPorts = map[string][]numorstring.Port{"": nil}
			break
		}

		pStr := protocol.String()
		// treat nil as 'all ports'
		if calicoPort == nil {
			protocolPorts[pStr] = nil
		} else if _, ok := protocolPorts[pStr]; !ok || len(protocolPorts[pStr]) > 0 {
			// don't overwrite a nil (allow all ports) if present; if no ports yet for this protocol
			// or 1+ ports which aren't 'all ports', then add the present ports
			protocolPorts[pStr] = append(protocolPorts[pStr], *calicoPort)
		}
	}

	protocols := make([]string, 0, len(protocolPorts))
	for k := range protocolPorts {
		protocols = append(protocols, k)
	}
	// Ensure deterministic output
	sort.Strings(protocols)

	// Combine destinations with sources to generate rules. We generate one rule per protocol,
	// with each rule containing all the allowed ports.
	for _, protocolStr := range protocols {
		calicoPorts := protocolPorts[protocolStr]
		calicoPorts = SimplifyPorts(calicoPorts)

		var protocol *numorstring.Protocol
		if protocolStr != "" {
			p := numorstring.ProtocolFromString(protocolStr)
			protocol = &p
		}

		// Based on specifications at least one Peer is set.
		for _, peer := range rulePeers {
			var selector, nsSelector string
			var nets []string
			// One and only one of the following fields is set (based on specification).
			var found bool
			if peer.Namespaces != nil {
				nsSelector = k8sSelectorToCalico(peer.Namespaces, SelectorNamespace)
				found = true
			}
			if peer.Pods != nil {
				selector = k8sSelectorToCalico(&peer.Pods.PodSelector, SelectorPod)
				nsSelector = k8sSelectorToCalico(&peer.Pods.NamespaceSelector, SelectorNamespace)
				found = true
			}
			if len(peer.Networks) != 0 {
				for _, n := range peer.Networks {
					_, ipNet, err := cnet.ParseCIDR(string(n))
					if err != nil {
						return nil, fmt.Errorf("invalid CIDR in ANP rule: %w", err)
					}
					nets = append(nets, ipNet.String())
				}
				found = true
			}
			if !found {
				return nil, fmt.Errorf("none of supported fields in 'To' is set.")
			}

			// Build outbound rule and append to list.
			rules = append(rules, apiv3.Rule{
				Metadata: k8sAdminNetworkPolicyToCalicoMetadata(ruleName),
				Action:   action,
				Protocol: protocol,
				Destination: apiv3.EntityRule{
					Ports:             calicoPorts,
					Selector:          selector,
					NamespaceSelector: nsSelector,
					Nets:              nets,
				},
			})
		}
	}

	return rules, nil
}

func K8sAdminNetworkPolicyActionToCalico(action adminpolicy.AdminNetworkPolicyRuleAction) (apiv3.Action, error) {
	switch action {
	case adminpolicy.AdminNetworkPolicyRuleActionAllow,
		adminpolicy.AdminNetworkPolicyRuleActionDeny,
		adminpolicy.AdminNetworkPolicyRuleActionPass:
		return apiv3.Action(action), nil
	default:
		return "", fmt.Errorf("unsupported admin network policy action %v", action)
	}
}

func K8sBaselineAdminNetworkPolicyActionToCalico(action adminpolicy.BaselineAdminNetworkPolicyRuleAction) (apiv3.Action, error) {
	switch action {
	case adminpolicy.BaselineAdminNetworkPolicyRuleActionAllow,
		adminpolicy.BaselineAdminNetworkPolicyRuleActionDeny:
		return apiv3.Action(action), nil
	default:
		return "", fmt.Errorf("unsupported admin network policy action %v", action)
	}
}

func k8sAdminNetworkPolicyToCalicoMetadata(ruleName string) *apiv3.RuleMetadata {
	if ruleName == "" {
		return nil
	}
	return &apiv3.RuleMetadata{
		Annotations: map[string]string{
			AdminPolicyRuleNameLabel: ruleName,
		},
	}
}

func k8sAdminPolicyPortToCalicoFields(port *adminpolicy.AdminNetworkPolicyPort) (
	protocol *numorstring.Protocol,
	dstPort *numorstring.Port,
	err error,
) {
	// If no port info, return zero values for all fields (protocol, dstPorts).
	if port == nil {
		return
	}
	// Only one of the PortNumber or PortRange is set.
	if port.PortNumber != nil {
		dstPort = k8sAdminPolicyPortToCalico(port.PortNumber)
		proto := ensureProtocol(port.PortNumber.Protocol)
		protocol = k8sProtocolToCalico(&proto)
		return
	}
	if port.PortRange != nil {
		dstPort, err = k8sAdminPolicyPortRangeToCalico(port.PortRange)
		if err != nil {
			return
		}
		proto := ensureProtocol(port.PortRange.Protocol)
		protocol = k8sProtocolToCalico(&proto)
		return
	}
	// TODO: Add support for NamedPorts
	return
}

func ensureProtocol(proto kapiv1.Protocol) kapiv1.Protocol {
	if proto != "" {
		return proto
	}
	return kapiv1.ProtocolTCP
}

func k8sAdminPolicyPortToCalico(port *adminpolicy.Port) *numorstring.Port {
	if port == nil {
		return nil
	}
	p := numorstring.SinglePort(uint16(port.Port))
	return &p
}

func k8sAdminPolicyPortRangeToCalico(port *adminpolicy.PortRange) (*numorstring.Port, error) {
	if port == nil {
		return nil, nil
	}
	p, err := numorstring.PortFromRange(uint16(port.Start), uint16(port.End))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// K8sNetworkPolicyToCalico converts a k8s NetworkPolicy to a model.KVPair.
func (c converter) K8sNetworkPolicyToCalico(np *networkingv1.NetworkPolicy) (*model.KVPair, error) {
	// Pull out important fields.
	policyName := names.K8sNetworkPolicyNamePrefix + np.Name

	// We insert all the NetworkPolicy Policies at order 1000.0 after conversion.
	// This order might change in future.
	order := float64(1000.0)

	errorTracker := cerrors.ErrorPolicyConversion{PolicyName: np.Name}

	// Generate the ingress rules list.
	var ingressRules []apiv3.Rule
	for _, r := range np.Spec.Ingress {
		rules, err := c.k8sRuleToCalico(r.From, r.Ports, true)
		if err != nil {
			log.WithError(err).Warn("dropping k8s rule that couldn't be converted.")
			// Add rule to conversion error slice
			errorTracker.BadIngressRule(&r, fmt.Sprintf("k8s rule couldn't be converted: %s", err))
		} else {
			ingressRules = append(ingressRules, rules...)
		}
	}

	// Generate the egress rules list.
	var egressRules []apiv3.Rule
	for _, r := range np.Spec.Egress {
		rules, err := c.k8sRuleToCalico(r.To, r.Ports, false)
		if err != nil {
			log.WithError(err).Warn("dropping k8s rule that couldn't be converted")
			// Add rule to conversion error slice
			errorTracker.BadEgressRule(&r, fmt.Sprintf("k8s rule couldn't be converted: %s", err))
		} else {
			egressRules = append(egressRules, rules...)
		}
	}

	// Calculate Types setting.
	ingress := false
	egress := false
	for _, policyType := range np.Spec.PolicyTypes {
		switch policyType {
		case networkingv1.PolicyTypeIngress:
			ingress = true
		case networkingv1.PolicyTypeEgress:
			egress = true
		}
	}
	policyTypes := []apiv3.PolicyType{}
	if ingress {
		policyTypes = append(policyTypes, apiv3.PolicyTypeIngress)
	}
	if egress {
		policyTypes = append(policyTypes, apiv3.PolicyTypeEgress)
	} else if len(egressRules) > 0 {
		// Egress was introduced at the same time as policyTypes.  It shouldn't be possible to
		// receive a NetworkPolicy with an egress rule but without "Egress" specified in its types,
		// but we'll warn about it anyway.
		log.Warn("K8s PolicyTypes don't include 'egress', but NetworkPolicy has egress rules.")
	}

	// If no types were specified in the policy, then we're running on a cluster that doesn't
	// include support for that field in the API.  In that case, the correct behavior is for the policy
	// to apply to only ingress traffic.
	if len(policyTypes) == 0 {
		policyTypes = append(policyTypes, apiv3.PolicyTypeIngress)
	}

	var uid types.UID
	var err error
	if np.UID != "" {
		uid, err = ConvertUID(np.UID)
		if err != nil {
			return nil, err
		}
	}

	// Create the NetworkPolicy.
	policy := apiv3.NewNetworkPolicy()
	policy.ObjectMeta = metav1.ObjectMeta{
		Name:              policyName,
		Namespace:         np.Namespace,
		CreationTimestamp: np.CreationTimestamp,
		UID:               uid,
		ResourceVersion:   np.ResourceVersion,
	}
	policy.Spec = apiv3.NetworkPolicySpec{
		Order:    &order,
		Selector: k8sSelectorToCalico(&np.Spec.PodSelector, SelectorPod),
		Ingress:  ingressRules,
		Egress:   egressRules,
		Types:    policyTypes,
	}

	// Build the KVPair.
	kvp := &model.KVPair{
		Key: model.ResourceKey{
			Name:      policyName,
			Namespace: np.Namespace,
			Kind:      apiv3.KindNetworkPolicy,
		},
		Value:    policy,
		Revision: np.ResourceVersion,
	}

	// Return the KVPair with conversion errors if applicable
	return kvp, errorTracker.GetError()
}

// k8sSelectorToCalico takes a namespaced k8s label selector and returns the Calico
// equivalent.
func k8sSelectorToCalico(s *metav1.LabelSelector, selectorType selectorType) string {
	// Only prefix pod selectors - this won't work for namespace selectors.
	selectors := []string{}
	if selectorType == SelectorPod {
		selectors = append(selectors, fmt.Sprintf("%s == 'k8s'", apiv3.LabelOrchestrator))
	}

	if s == nil {
		return strings.Join(selectors, " && ")
	}

	// For namespace selectors, if they are present but have no terms, it means "select all
	// namespaces". We use empty string to represent the nil namespace selector, so use all() to
	// represent all namespaces.
	if selectorType == SelectorNamespace && len(s.MatchLabels) == 0 && len(s.MatchExpressions) == 0 {
		return "all()"
	}

	// matchLabels is a map key => value, it means match if (label[key] ==
	// value) for all keys.
	keys := make([]string, 0, len(s.MatchLabels))
	for k := range s.MatchLabels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := s.MatchLabels[k]
		selectors = append(selectors, fmt.Sprintf("%s == '%s'", k, v))
	}

	// matchExpressions is a list of in/notin/exists/doesnotexist tests.
	for _, e := range s.MatchExpressions {
		valueList := strings.Join(e.Values, "', '")

		// Each selector is formatted differently based on the operator.
		switch e.Operator {
		case metav1.LabelSelectorOpIn:
			selectors = append(selectors, fmt.Sprintf("%s in { '%s' }", e.Key, valueList))
		case metav1.LabelSelectorOpNotIn:
			selectors = append(selectors, fmt.Sprintf("%s not in { '%s' }", e.Key, valueList))
		case metav1.LabelSelectorOpExists:
			selectors = append(selectors, fmt.Sprintf("has(%s)", e.Key))
		case metav1.LabelSelectorOpDoesNotExist:
			selectors = append(selectors, fmt.Sprintf("! has(%s)", e.Key))
		}
	}

	return strings.Join(selectors, " && ")
}

func (c converter) k8sRuleToCalico(
	rPeers []networkingv1.NetworkPolicyPeer,
	rPorts []networkingv1.NetworkPolicyPort,
	ingress bool,
) ([]apiv3.Rule, error) {
	rules := []apiv3.Rule{}
	peers := []*networkingv1.NetworkPolicyPeer{}
	ports := []*networkingv1.NetworkPolicyPort{}

	// Built up a list of the sources and a list of the destinations.
	for _, f := range rPeers {
		// We need to add a copy of the peer so all the rules don't
		// point to the same location.
		peers = append(peers, &networkingv1.NetworkPolicyPeer{
			NamespaceSelector: f.NamespaceSelector,
			PodSelector:       f.PodSelector,
			IPBlock:           f.IPBlock,
		})
	}
	for _, p := range rPorts {
		// We need to add a copy of the port so all the rules don't
		// point to the same location.
		port := networkingv1.NetworkPolicyPort{}
		if p.Port != nil {
			portval := intstr.FromString(p.Port.String())
			port.Port = &portval
		}
		if p.Protocol != nil {
			protval := kapiv1.Protocol(fmt.Sprintf("%s", *p.Protocol))
			port.Protocol = &protval
		} else {
			// TCP is the implicit default (as per the definition of NetworkPolicyPort).
			// Make the default explicit here because our data-model always requires
			// the protocol to be specified if we're doing a port match.
			port.Protocol = &protoTCP
		}

		if p.EndPort != nil {
			port.EndPort = p.EndPort
		}
		ports = append(ports, &port)
	}

	// If there no peers, or no ports, represent that as nil.
	if len(peers) == 0 {
		peers = []*networkingv1.NetworkPolicyPeer{nil}
	}
	if len(ports) == 0 {
		ports = []*networkingv1.NetworkPolicyPort{nil}
	}

	protocolPorts := map[string][]numorstring.Port{}

	for _, port := range ports {
		protocol, calicoPorts, err := c.k8sPortToCalicoFields(port)
		if err != nil {
			return nil, fmt.Errorf("failed to parse k8s port: %s", err)
		}

		if protocol == nil && calicoPorts == nil {
			// If nil, no ports were specified, or an empty port struct was provided, which we translate to allowing all.
			// We want to use a nil protocol and a nil list of ports, which will allow any destination (for ingress).
			// Given we're gonna allow all, we may as well break here and keep only this rule
			protocolPorts = map[string][]numorstring.Port{"": nil}
			break
		}

		pStr := protocol.String()
		// treat nil as 'all ports'
		if calicoPorts == nil {
			protocolPorts[pStr] = nil
		} else if _, ok := protocolPorts[pStr]; !ok || len(protocolPorts[pStr]) > 0 {
			// don't overwrite a nil (allow all ports) if present; if no ports yet for this protocol
			// or 1+ ports which aren't 'all ports', then add the present ports
			protocolPorts[pStr] = append(protocolPorts[pStr], calicoPorts...)
		}
	}

	protocols := make([]string, 0, len(protocolPorts))
	for k := range protocolPorts {
		protocols = append(protocols, k)
	}
	// Ensure deterministic output
	sort.Strings(protocols)

	// Combine destinations with sources to generate rules. We generate one rule per protocol,
	// with each rule containing all the allowed ports.
	for _, protocolStr := range protocols {
		calicoPorts := protocolPorts[protocolStr]
		calicoPorts = SimplifyPorts(calicoPorts)

		var protocol *numorstring.Protocol
		if protocolStr != "" {
			p := numorstring.ProtocolFromString(protocolStr)
			protocol = &p
		}

		for _, peer := range peers {
			selector, nsSelector, nets, notNets := c.k8sPeerToCalicoFields(peer)
			if ingress {
				// Build inbound rule and append to list.
				rules = append(rules, apiv3.Rule{
					Action:   "Allow",
					Protocol: protocol,
					Source: apiv3.EntityRule{
						Selector:          selector,
						NamespaceSelector: nsSelector,
						Nets:              nets,
						NotNets:           notNets,
					},
					Destination: apiv3.EntityRule{
						Ports: calicoPorts,
					},
				})
			} else {
				// Build outbound rule and append to list.
				rules = append(rules, apiv3.Rule{
					Action:   "Allow",
					Protocol: protocol,
					Destination: apiv3.EntityRule{
						Ports:             calicoPorts,
						Selector:          selector,
						NamespaceSelector: nsSelector,
						Nets:              nets,
						NotNets:           notNets,
					},
				})
			}
		}
	}
	return rules, nil
}

// SimplifyPorts calculates a minimum set of port ranges that cover the given set of ports.
// For example, if the input was [80, 81, 82, 9090, "foo"] the output would consist of
// [80-82, 9090, "foo"] in some order.
func SimplifyPorts(ports []numorstring.Port) []numorstring.Port {
	if len(ports) <= 1 {
		return ports
	}
	var numericPorts []int
	var outputPorts []numorstring.Port
	for _, p := range ports {
		if p.PortName != "" {
			// Pass named ports through immediately, there's nothing to be done for them.
			outputPorts = append(outputPorts, p)
		} else {
			// Work with ints to avoid overflow with the uint16 port type.
			// In practice, we currently only get single ports here so this
			// loop should run exactly once.
			for i := int(p.MinPort); i <= int(p.MaxPort); i++ {
				numericPorts = append(numericPorts, i)
			}
		}
	}

	if len(numericPorts) <= 1 {
		// We have nothing to combine, short-circuit.
		return ports
	}

	// Sort the ports so it will be easy to find ranges.
	sort.Ints(numericPorts)

	// Each pass around this outer loop extracts one port range from the sorted slice
	// and it moves the slice along to the start of the next range.
	for len(numericPorts) > 0 {
		// Initialise the next range to the contain only the first port in the slice.
		firstPortInRange := numericPorts[0]
		lastPortInRange := firstPortInRange

		// Scan ahead, looking for ports that can be combined into this range.
		numericPorts = numericPorts[1:]
		for len(numericPorts) > 0 {
			nextPort := numericPorts[0]
			if nextPort > lastPortInRange+1 {
				// This port can't be coalesced with the existing range, break out so
				// that we record the range; then we'll loop again and pick up this
				// port as the start of a new range.
				break
			}
			// The next port is either equal to the last port (due to a duplicate port
			// in the input) or it is exactly one greater.  Extend the range to include
			// it.
			lastPortInRange = nextPort
			numericPorts = numericPorts[1:]
		}

		// Record the port.
		outputPorts = appendPortRange(outputPorts, firstPortInRange, lastPortInRange)
	}

	return outputPorts
}

func appendPortRange(ports []numorstring.Port, first, last int) []numorstring.Port {
	portRange, err := numorstring.PortFromRange(uint16(first), uint16(last))
	if err != nil {
		log.WithError(err).Panic("Failed to make port range from ports that should have been pre-validated.")
	}
	return append(ports, portRange)
}

func (c converter) k8sPortToCalicoFields(port *networkingv1.NetworkPolicyPort) (protocol *numorstring.Protocol, dstPorts []numorstring.Port, err error) {
	// If no port info, return zero values for all fields (protocol, dstPorts).
	if port == nil {
		return
	}
	// Port information available.
	dstPorts, err = c.k8sPortToCalico(*port)
	if err != nil {
		return
	}
	protocol = k8sProtocolToCalico(port.Protocol)
	return
}

func k8sProtocolToCalico(protocol *kapiv1.Protocol) *numorstring.Protocol {
	if protocol != nil {
		p := numorstring.ProtocolFromString(string(*protocol))
		return &p
	}
	return nil
}

func (c converter) k8sPeerToCalicoFields(peer *networkingv1.NetworkPolicyPeer) (selector, nsSelector string, nets []string, notNets []string) {
	// If no peer, return zero values for all fields (selector, nets and !nets).
	if peer == nil {
		return
	}
	// Peer information available.
	// Determine the source selector for the rule.
	if peer.IPBlock != nil {
		// Convert the CIDR to include.
		_, ipNet, err := cnet.ParseCIDR(peer.IPBlock.CIDR)
		if err != nil {
			log.WithField("cidr", peer.IPBlock.CIDR).WithError(err).Error("Failed to parse CIDR")
			return
		}
		nets = []string{ipNet.String()}

		// Convert the CIDRs to exclude.
		for _, exception := range peer.IPBlock.Except {
			_, ipNet, err = cnet.ParseCIDR(exception)
			if err != nil {
				log.WithField("cidr", exception).WithError(err).Error("Failed to parse CIDR")
				return
			}
			notNets = append(notNets, ipNet.String())
		}
		// If IPBlock is set, then PodSelector and NamespaceSelector cannot be.
		return
	}

	// IPBlock is not set to get here.
	// Note that k8sSelectorToCalico() accepts nil values of the selector.
	selector = k8sSelectorToCalico(peer.PodSelector, SelectorPod)
	nsSelector = k8sSelectorToCalico(peer.NamespaceSelector, SelectorNamespace)
	return
}

func (c converter) k8sPortToCalico(port networkingv1.NetworkPolicyPort) ([]numorstring.Port, error) {
	var portList []numorstring.Port
	if port.Port != nil {
		calicoPort := port.Port.String()
		if port.EndPort != nil {
			calicoPort = fmt.Sprintf("%s:%d", calicoPort, *port.EndPort)
		}
		p, err := numorstring.PortFromString(calicoPort)
		if err != nil {
			return nil, fmt.Errorf("invalid port %+v: %s", calicoPort, err)
		}
		return append(portList, p), nil
	}

	// No ports - return empty list.
	return portList, nil
}

// ProfileNameToNamespace extracts the Namespace name from the given Profile name.
func (c converter) ProfileNameToNamespace(profileName string) (string, error) {
	// Profile objects backed by Namespaces have form "kns.<ns_name>"
	if !strings.HasPrefix(profileName, NamespaceProfileNamePrefix) {
		// This is not backed by a Kubernetes Namespace.
		return "", fmt.Errorf("Profile %s not backed by a Namespace", profileName)
	}

	return strings.TrimPrefix(profileName, NamespaceProfileNamePrefix), nil
}

// serviceAccountNameToProfileName creates a profile name that is a join
// of 'ksa.' + namespace + "." + serviceaccount name.
func serviceAccountNameToProfileName(sa, namespace string) string {
	// Need to incorporate the namespace into the name of the sa based profile
	// to make them globally unique
	if namespace == "" {
		namespace = "default"
	}
	return ServiceAccountProfileNamePrefix + namespace + "." + sa
}

// ServiceAccountToProfile converts a ServiceAccount to a Calico Profile.  The Profile stores
// labels from the ServiceAccount which are inherited by the WorkloadEndpoints within
// the Profile.
func (c converter) ServiceAccountToProfile(sa *kapiv1.ServiceAccount) (*model.KVPair, error) {
	// Generate the labels to apply to the profile, using a special prefix
	// to indicate that these are the labels from the parent Kubernetes ServiceAccount.
	labels := map[string]string{}
	for k, v := range sa.ObjectMeta.Labels {
		labels[ServiceAccountLabelPrefix+k] = v
	}

	// Add a label for the serviceaccount's name. This allows exact namespace matching
	// based on name within the serviceAccountSelector.
	labels[ServiceAccountLabelPrefix+NameLabel] = sa.Name

	uid, err := ConvertUID(sa.UID)
	if err != nil {
		return nil, err
	}

	name := serviceAccountNameToProfileName(sa.Name, sa.Namespace)
	profile := apiv3.NewProfile()
	profile.ObjectMeta = metav1.ObjectMeta{
		Name:              name,
		CreationTimestamp: sa.CreationTimestamp,
		UID:               uid,
	}
	profile.Spec.LabelsToApply = labels

	// Embed the profile in a KVPair.
	kvp := model.KVPair{
		Key: model.ResourceKey{
			Name: name,
			Kind: apiv3.KindProfile,
		},
		Value:    profile,
		Revision: c.JoinProfileRevisions("", sa.ResourceVersion),
	}
	return &kvp, nil
}

// ProfileNameToServiceAccount extracts the ServiceAccount name from the given Profile name.
func (c converter) ProfileNameToServiceAccount(profileName string) (ns, sa string, err error) {
	// Profile objects backed by ServiceAccounts have form "ksa.<namespace>.<sa_name>"
	if !strings.HasPrefix(profileName, ServiceAccountProfileNamePrefix) {
		// This is not backed by a Kubernetes ServiceAccount.
		err = fmt.Errorf("Profile %s not backed by a ServiceAccount", profileName)
		return
	}

	names := strings.SplitN(profileName, ".", 3)
	if len(names) != 3 {
		err = fmt.Errorf("Profile %s is not formatted correctly", profileName)
		return
	}

	ns = names[1]
	sa = names[2]
	return
}

// JoinProfileRevisions constructs the revision from the individual namespace and serviceaccount
// revisions.
// This is conditional on the feature flag for serviceaccount set or not.
func (c converter) JoinProfileRevisions(nsRev, saRev string) string {
	return nsRev + "/" + saRev
}

// SplitProfileRevision extracts the namespace and serviceaccount revisions from the combined
// revision returned on the KDD service account based profile.
// This is conditional on the feature flag for serviceaccount set or not.
func (c converter) SplitProfileRevision(rev string) (nsRev string, saRev string, err error) {
	if rev == "" || rev == "0" {
		return
	}

	revs := strings.Split(rev, "/")
	if len(revs) != 2 {
		err = fmt.Errorf("ResourceVersion is not valid: %s", rev)
		return
	}
	nsRev = revs[0]
	saRev = revs[1]
	return
}

func stringsToIPNets(ipStrings []string) ([]*cnet.IPNet, error) {
	var podIPNets []*cnet.IPNet
	for _, ip := range ipStrings {
		_, ipNet, err := cnet.ParseCIDROrIP(ip)
		if err != nil {
			return nil, err
		}
		podIPNets = append(podIPNets, ipNet)
	}
	return podIPNets, nil
}

// ConvertUID converts a UID to a new UID in a deterministic way. This is useful when we want to generate a new UID
// for a resource that is derived from another resource, but we don't want to use the same UID in order to
// ensure that the new resource is treated as unique. This is important, as two objects with the same UID causes
// confusion in the Kubernetes garbage collection logic.
func ConvertUID(uid types.UID) (types.UID, error) {
	parsed, err := uuid.Parse(string(uid))
	if err != nil {
		return "", fmt.Errorf("failed to parse UID for resource: %s", err)
	}
	reversed, err := reverseUID(parsed)
	if err != nil {
		return "", fmt.Errorf("failed to reverse UID for resource: %s", err)
	}
	return types.UID(reversed.String()), nil
}

func reverseUID(uid uuid.UUID) (uuid.UUID, error) {
	// v4 UUIDs used by Kubernetes use bits in the 7th byte to indicate the version and
	// bits in the 9th byte to indicate the variant. Reverse the bits in the surrounding bytes but leave these intact.
	nuid := make([]byte, len(uid))
	copy(nuid, uid[:])

	// Reverse the bits in the first 6 bytes.
	for ii := range uid[:6] {
		nuid[ii] = byte(bits.Reverse(uint(uid[ii])) >> 56)
	}

	// Reverse the bits in the 8th byte.
	nuid[7] = byte(bits.Reverse(uint(uid[7])) >> 56)

	// Reverse the bits in the remaining bytes.
	for ii := range uid[9:] {
		nuid[ii+9] = byte(bits.Reverse(uint(uid[ii+9])) >> 56)
	}
	return uuid.FromBytes(nuid)
}
