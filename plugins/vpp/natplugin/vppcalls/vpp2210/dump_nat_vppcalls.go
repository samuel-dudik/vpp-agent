//  Copyright (c) 2022 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp2210

import (
	"context"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"

	vpp_nat_ed "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/nat44_ed"
	vpp_nat_ei "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/nat44_ei"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/nat_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

// DNATs sorted by tags
type dnatMap map[string]*nat.DNat44

// static mappings sorted by tags
type stMappingMap map[string][]*nat.DNat44_StaticMapping

// identity mappings sorted by tags
type idMappingMap map[string][]*nat.DNat44_IdentityMapping

// WithLegacyStartupConf returns true if the loaded VPP NAT plugin is still using
// the legacy startup NAT configuration (this is the case for VPP <= 20.09).
func (h *NatVppHandler) WithLegacyStartupConf() bool {
	return false
}

func (h *NatVppHandler) DefaultNat44GlobalConfig() *nat.Nat44Global {
	return &nat.Nat44Global{
		Forwarding:          false,
		EndpointIndependent: false,
		NatInterfaces:       nil,
		AddressPool:         nil,
		VirtualReassembly:   nil, // VirtualReassembly is not part of NAT API in VPP 20.01+ anymore
	}
}

func (h *NatVppHandler) nat44EiGlobalConfigDump(dumpDeprecated bool) (cfg *nat.Nat44Global, err error) {
	cfg = &nat.Nat44Global{}
	req := &vpp_nat_ei.Nat44EiShowRunningConfig{}
	reply := &vpp_nat_ei.Nat44EiShowRunningConfigReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}
	cfg.EndpointIndependent = true
	cfg.Forwarding = reply.ForwardingEnabled
	cfg.VirtualReassembly, _, err = h.virtualReassemblyDump()
	if err != nil {
		return nil, err
	}
	if dumpDeprecated {
		cfg.NatInterfaces, err = h.nat44InterfaceDump()
		if err != nil {
			return nil, err
		}
		cfg.AddressPool, err = h.nat44AddressDump()
		if err != nil {
			return nil, err
		}
	}
	return
}

func (h *NatVppHandler) nat44EdGlobalConfigDump(dumpDeprecated bool) (cfg *nat.Nat44Global, err error) {
	cfg = &nat.Nat44Global{}
	req := &vpp_nat_ed.Nat44ShowRunningConfig{}
	reply := &vpp_nat_ed.Nat44ShowRunningConfigReply{}
	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}
	cfg.EndpointIndependent = false
	cfg.Forwarding = reply.ForwardingEnabled
	cfg.VirtualReassembly, _, err = h.virtualReassemblyDump()
	if err != nil {
		return nil, err
	}
	if dumpDeprecated {
		cfg.NatInterfaces, err = h.nat44InterfaceDump()
		if err != nil {
			return nil, err
		}
		cfg.AddressPool, err = h.nat44AddressDump()
		if err != nil {
			return nil, err
		}
	}
	return
}

// Nat44GlobalConfigDump dumps global NAT44 config in NB format.
func (h *NatVppHandler) Nat44GlobalConfigDump(dumpDeprecated bool) (cfg *nat.Nat44Global, err error) {
	if h.ed {
		return h.nat44EdGlobalConfigDump(dumpDeprecated)
	} else {
		return h.nat44EiGlobalConfigDump(dumpDeprecated)
	}

}

// DNat44Dump dumps all configured DNAT-44 configurations ordered by label.
func (h *NatVppHandler) DNat44Dump() (dnats []*nat.DNat44, err error) {
	dnatMap := make(dnatMap)

	// Static mappings
	natStMappings, err := h.nat44StaticMappingDump()
	if err != nil {
		return nil, fmt.Errorf("failed to dump NAT44 static mappings: %v", err)
	}
	for label, mappings := range natStMappings {
		dnat := getOrCreateDNAT(dnatMap, label)
		dnat.StMappings = append(dnat.StMappings, mappings...)
	}

	// Static mappings with load balancer
	natStLbMappings, err := h.nat44StaticMappingLbDump()
	if err != nil {
		return nil, fmt.Errorf("failed to dump NAT44 static mappings with load balancer: %v", err)
	}
	for label, mappings := range natStLbMappings {
		dnat := getOrCreateDNAT(dnatMap, label)
		dnat.StMappings = append(dnat.StMappings, mappings...)
	}

	// Identity mappings
	natIDMappings, err := h.nat44IdentityMappingDump()
	if err != nil {
		return nil, fmt.Errorf("failed to dump NAT44 identity mappings: %v", err)
	}
	for label, mappings := range natIDMappings {
		dnat := getOrCreateDNAT(dnatMap, label)
		dnat.IdMappings = append(dnat.IdMappings, mappings...)
	}

	// Convert map of DNAT configurations into a list.
	for _, dnat := range dnatMap {
		dnats = append(dnats, dnat)
	}

	// sort to simplify testing
	sort.Slice(dnats, func(i, j int) bool { return dnats[i].Label < dnats[j].Label })

	return dnats, nil
}

func (h *NatVppHandler) nat44EiInterfacesDump() ([]*nat.Nat44Interface, error) {
	var natIfs []*nat.Nat44Interface

	// dump non-output NAT interfaces
	req1 := &vpp_nat_ei.Nat44EiInterfaceDump{}
	reqContext := h.callsChannel.SendMultiRequest(req1)
	for {
		msg := &vpp_nat_ei.Nat44EiInterfaceDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface: %v", err)
		}
		if stop {
			break
		}
		ifName, _, found := h.ifIndexes.LookupBySwIfIndex(uint32(msg.SwIfIndex))
		if !found {
			h.log.Warnf("Interface with index %d not found in the mapping", msg.SwIfIndex)
			continue
		}
		flags := getNat44EiFlags(msg.Flags)
		natIf := &nat.Nat44Interface{
			Name:          ifName,
			NatInside:     flags.eiIfInside,
			NatOutside:    flags.eiIfOutside,
			OutputFeature: false,
		}
		natIfs = append(natIfs, natIf)
	}

	// dump output NAT interfaces
	var cursor uint32
	for {
		rpcServ, err := h.natEi.Nat44EiOutputInterfaceGet(context.Background(), &vpp_nat_ei.Nat44EiOutputInterfaceGet{
			Cursor: cursor,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface output feature: %v", err)
		}
	RecvLoop:
		for {
			details, reply, err := rpcServ.Recv()
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("failed to dump NAT44 interface output feature: %v", err)
			}
			if reply != nil {
				// TODO: Possible bug in VPP? When VPP contains no NAT interfaces this
				// reply often returns cursor value of 0 repeatedly even though it should return
				// ^uint32(0). So if the reply contains cursor that is the same as cursor
				// from previous reply, return.
				if reply.Cursor == cursor || reply.Cursor == ^uint32(0) {
					return natIfs, nil
				}
				cursor = reply.Cursor
				break RecvLoop
			}
			if details != nil {
				ifName, _, found := h.ifIndexes.LookupBySwIfIndex(uint32(details.SwIfIndex))
				if !found {
					h.log.Warnf("Interface with index %d not found in the mapping", details.SwIfIndex)
				}
				natIfs = append(natIfs, &nat.Nat44Interface{
					Name:          ifName,
					OutputFeature: true,
				})
			}
		}
	}
}

func (h *NatVppHandler) nat44EdInterfacesDump() ([]*nat.Nat44Interface, error) {
	var natIfs []*nat.Nat44Interface

	// dump non-output NAT interfaces
	req1 := &vpp_nat_ed.Nat44InterfaceDump{}
	reqContext := h.callsChannel.SendMultiRequest(req1)
	for {
		msg := &vpp_nat_ed.Nat44InterfaceDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface: %v", err)
		}
		if stop {
			break
		}
		ifName, _, found := h.ifIndexes.LookupBySwIfIndex(uint32(msg.SwIfIndex))
		if !found {
			h.log.Warnf("Interface with index %d not found in the mapping", msg.SwIfIndex)
			continue
		}
		flags := getNat44Flags(msg.Flags)
		natIf := &nat.Nat44Interface{
			Name:          ifName,
			NatInside:     flags.isInside,
			NatOutside:    flags.isOutside,
			OutputFeature: false,
		}
		natIfs = append(natIfs, natIf)
	}

	// dump output NAT interfaces
	var cursor uint32
	for {
		rpcServ, err := h.natEd.Nat44EdOutputInterfaceGet(context.Background(), &vpp_nat_ed.Nat44EdOutputInterfaceGet{
			Cursor: cursor,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface output feature: %v", err)
		}
	RecvLoop:
		for {
			details, reply, err := rpcServ.Recv()
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("failed to dump NAT44 interface output feature: %v", err)
			}
			if reply != nil {
				// TODO: Possible bug in VPP? When VPP contains no NAT interfaces this
				// reply often returns cursor value of 0 repeatedly even though it should return
				// ^uint32(0). So if the reply contains cursor that is the same as cursor
				// from previous reply, return.
				if reply.Cursor == cursor || reply.Cursor == ^uint32(0) {
					return natIfs, nil
				}
				cursor = reply.Cursor
				break RecvLoop
			}
			if details != nil {
				ifName, _, found := h.ifIndexes.LookupBySwIfIndex(uint32(details.SwIfIndex))
				if !found {
					h.log.Warnf("Interface with index %d not found in the mapping", details.SwIfIndex)
				}
				natIfs = append(natIfs, &nat.Nat44Interface{
					Name:          ifName,
					OutputFeature: true,
				})
			}
		}
	}
}

// Nat44InterfacesDump dumps NAT44 config of all NAT44-enabled interfaces.
func (h *NatVppHandler) Nat44InterfacesDump() (natIfs []*nat.Nat44Interface, err error) {
	if h.ed {
		return h.nat44EdInterfacesDump()
	} else {
		return h.nat44EiInterfacesDump()
	}
}

func (h *NatVppHandler) nat44EiAddressPoolsDump() (natPools []*nat.Nat44AddressPool, err error) {
	var curPool *nat.Nat44AddressPool
	var lastIP net.IP

	req := &vpp_nat_ei.Nat44EiAddressDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &vpp_nat_ei.Nat44EiAddressDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 Address pool: %v", err)
		}
		if stop {
			break
		}
		ip := net.IP(msg.IPAddress[:])
		// merge subsequent IPs into a single pool
		if curPool != nil && curPool.VrfId == msg.VrfID && ip.Equal(incIP(lastIP)) {
			// update current pool
			curPool.LastIp = ip.String()
		} else {
			// start a new pool
			pool := &nat.Nat44AddressPool{
				FirstIp:  ip.String(),
				VrfId:    msg.VrfID,
				TwiceNat: false,
			}
			curPool = pool
			natPools = append(natPools, pool)
		}
		lastIP = ip
	}
	return
}

func (h *NatVppHandler) nat44EdAddressPoolsDump() (natPools []*nat.Nat44AddressPool, err error) {
	var curPool *nat.Nat44AddressPool
	var lastIP net.IP

	req := &vpp_nat_ed.Nat44AddressDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &vpp_nat_ed.Nat44AddressDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 Address pool: %v", err)
		}
		if stop {
			break
		}
		ip := net.IP(msg.IPAddress[:])
		isTwiceNat := getNat44Flags(msg.Flags).isTwiceNat
		// merge subsequent IPs into a single pool
		if curPool != nil && curPool.VrfId == msg.VrfID && curPool.TwiceNat == isTwiceNat && ip.Equal(incIP(lastIP)) {
			// update current pool
			curPool.LastIp = ip.String()
		} else {
			// start a new pool
			pool := &nat.Nat44AddressPool{
				FirstIp:  ip.String(),
				VrfId:    msg.VrfID,
				TwiceNat: isTwiceNat,
			}
			curPool = pool
			natPools = append(natPools, pool)
		}
		lastIP = ip
	}
	return
}

// Nat44AddressPoolsDump dumps all configured NAT44 address pools.
func (h *NatVppHandler) Nat44AddressPoolsDump() (natPools []*nat.Nat44AddressPool, err error) {
	if h.ed {
		return h.nat44EdAddressPoolsDump()
	} else {
		return h.nat44EiAddressPoolsDump()
	}
}

func (h *NatVppHandler) nat44EiAddressDump() (addressPool []*nat.Nat44Global_Address, err error) {
	req := &vpp_nat_ei.Nat44EiAddressDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &vpp_nat_ei.Nat44EiAddressDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 Address pool: %v", err)
		}
		if stop {
			break
		}

		addressPool = append(addressPool, &nat.Nat44Global_Address{
			Address:  net.IP(msg.IPAddress[:]).String(),
			VrfId:    msg.VrfID,
			TwiceNat: false,
		})
	}

	return
}

func (h *NatVppHandler) nat44EdAddressDump() (addressPool []*nat.Nat44Global_Address, err error) {
	req := &vpp_nat_ed.Nat44AddressDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &vpp_nat_ed.Nat44AddressDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 Address pool: %v", err)
		}
		if stop {
			break
		}

		addressPool = append(addressPool, &nat.Nat44Global_Address{
			Address:  net.IP(msg.IPAddress[:]).String(),
			VrfId:    msg.VrfID,
			TwiceNat: getNat44Flags(msg.Flags).isTwiceNat,
		})
	}

	return
}

// nat44AddressDump returns NAT44 address pool configured in the VPP.
// Deprecated. Functionality moved to Nat44AddressPoolsDump. Kept for backward compatibility.
func (h *NatVppHandler) nat44AddressDump() (addressPool []*nat.Nat44Global_Address, err error) {
	if h.ed {
		return h.nat44EdAddressDump()
	} else {
		return h.nat44EiAddressDump()
	}
}

// virtualReassemblyDump returns current NAT virtual-reassembly configuration.
func (h *NatVppHandler) virtualReassemblyDump() (vrIPv4 *nat.VirtualReassembly, vrIPv6 *nat.VirtualReassembly, err error) {
	/*ipv4vr, err := h.ip.IPReassemblyGet(context.TODO(), &vpp_ip.IPReassemblyGet{IsIP6: false})
	if err != nil {
		return nil, nil, fmt.Errorf("getting virtual reassembly IPv4 config failed: %w", err)
	}
	h.log.Debugf("IP Reassembly config IPv4: %+v\n", ipv4vr)
	ipv6vr, err := h.ip.IPReassemblyGet(context.TODO(), &vpp_ip.IPReassemblyGet{IsIP6: true})
	if err != nil {
		return nil, nil, fmt.Errorf("getting virtual reassembly IPv6 config failed: %w", err)
	}
	h.log.Debugf("IP Reassembly config IPv6: %+v\n", ipv6vr)*/

	// Virtual Reassembly has been removed from NAT API in VPP (moved to IP API)
	// TODO: define IPReassembly model in L3 plugin
	return nil, nil, nil
	/*vrIPv4 = &nat.VirtualReassembly{
		Timeout:         reply.IP4Timeout,
		MaxReassemblies: uint32(reply.IP4MaxReass),
		MaxFragments:    uint32(reply.IP4MaxFrag),
		DropFragments:   uintToBool(reply.IP4DropFrag),
	}
	vrIPv6 = &nat.VirtualReassembly{
		Timeout:         reply.IP6Timeout,
		MaxReassemblies: uint32(reply.IP6MaxReass),
		MaxFragments:    uint32(reply.IP6MaxFrag),
		DropFragments:   uintToBool(reply.IP6DropFrag),
	}
	return*/
}

func (h *NatVppHandler) nat44EiStaticMappingDump() (entries stMappingMap, err error) {
	entries = make(stMappingMap)
	childMappings := make(stMappingMap)
	req := &vpp_nat_ei.Nat44EiStaticMappingDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &vpp_nat_ei.Nat44EiStaticMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 static mapping: %v", err)
		}
		if stop {
			break
		}
		lcIPAddress := net.IP(msg.LocalIPAddress[:]).String()
		exIPAddress := net.IP(msg.ExternalIPAddress[:]).String()

		// Parse tag (DNAT label)
		tag := strings.TrimRight(msg.Tag, "\x00")
		if _, hasTag := entries[tag]; !hasTag {
			entries[tag] = []*nat.DNat44_StaticMapping{}
			childMappings[tag] = []*nat.DNat44_StaticMapping{}
		}

		// resolve interface name
		var (
			found        bool
			extIfaceName string
			extIfaceMeta *ifaceidx.IfaceMetadata
		)
		if msg.ExternalSwIfIndex != NoInterface {
			extIfaceName, extIfaceMeta, found = h.ifIndexes.LookupBySwIfIndex(uint32(msg.ExternalSwIfIndex))
			if !found {
				h.log.Warnf("Interface with index %v not found in the mapping", msg.ExternalSwIfIndex)
				continue
			}
		}

		// Add mapping into the map.
		mapping := &nat.DNat44_StaticMapping{
			ExternalInterface: extIfaceName,
			ExternalIp:        exIPAddress,
			ExternalPort:      uint32(msg.ExternalPort),
			LocalIps: []*nat.DNat44_StaticMapping_LocalIP{ // single-value
				{
					VrfId:     msg.VrfID,
					LocalIp:   lcIPAddress,
					LocalPort: uint32(msg.LocalPort),
				},
			},
			Protocol: h.protocolNumberToNBValue(msg.Protocol),
			TwiceNat: h.getTwiceNatMode(false, false),
			// if there is only one backend the affinity can not be set
			SessionAffinity: 0,
		}
		entries[tag] = append(entries[tag], mapping)

		if msg.ExternalSwIfIndex != NoInterface {
			// collect auto-generated "child" mappings (interface replaced with every assigned IP address)
			for _, ipAddr := range h.getInterfaceIPAddresses(extIfaceName, extIfaceMeta) {
				childMapping := proto.Clone(mapping).(*nat.DNat44_StaticMapping)
				childMapping.ExternalIp = ipAddr
				childMapping.ExternalInterface = ""
				childMappings[tag] = append(childMappings[tag], childMapping)
			}
		}
	}

	// do not dump auto-generated child mappings
	for tag, mappings := range entries {
		var filtered []*nat.DNat44_StaticMapping
		for _, mapping := range mappings {
			isChild := false
			for _, child := range childMappings[tag] {
				if proto.Equal(mapping, child) {
					isChild = true
					break
				}
			}
			if !isChild {
				filtered = append(filtered, mapping)
			}
		}
		entries[tag] = filtered
	}
	return entries, nil
}

func (h *NatVppHandler) nat44EdStaticMappingDump() (entries stMappingMap, err error) {
	entries = make(stMappingMap)
	childMappings := make(stMappingMap)
	req := &vpp_nat_ed.Nat44StaticMappingDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &vpp_nat_ed.Nat44StaticMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 static mapping: %v", err)
		}
		if stop {
			break
		}
		lcIPAddress := net.IP(msg.LocalIPAddress[:]).String()
		exIPAddress := net.IP(msg.ExternalIPAddress[:]).String()

		// Parse tag (DNAT label)
		tag := strings.TrimRight(msg.Tag, "\x00")
		if _, hasTag := entries[tag]; !hasTag {
			entries[tag] = []*nat.DNat44_StaticMapping{}
			childMappings[tag] = []*nat.DNat44_StaticMapping{}
		}

		// resolve interface name
		var (
			found        bool
			extIfaceName string
			extIfaceMeta *ifaceidx.IfaceMetadata
		)
		if msg.ExternalSwIfIndex != NoInterface {
			extIfaceName, extIfaceMeta, found = h.ifIndexes.LookupBySwIfIndex(uint32(msg.ExternalSwIfIndex))
			if !found {
				h.log.Warnf("Interface with index %v not found in the mapping", msg.ExternalSwIfIndex)
				continue
			}
		}

		flags := getNat44Flags(msg.Flags)

		// Add mapping into the map.
		mapping := &nat.DNat44_StaticMapping{
			ExternalInterface: extIfaceName,
			ExternalIp:        exIPAddress,
			ExternalPort:      uint32(msg.ExternalPort),
			LocalIps: []*nat.DNat44_StaticMapping_LocalIP{ // single-value
				{
					VrfId:     msg.VrfID,
					LocalIp:   lcIPAddress,
					LocalPort: uint32(msg.LocalPort),
				},
			},
			Protocol: h.protocolNumberToNBValue(msg.Protocol),
			TwiceNat: h.getTwiceNatMode(flags.isTwiceNat, flags.isSelfTwiceNat),
			// if there is only one backend the affinity can not be set
			SessionAffinity: 0,
		}
		entries[tag] = append(entries[tag], mapping)

		if msg.ExternalSwIfIndex != NoInterface {
			// collect auto-generated "child" mappings (interface replaced with every assigned IP address)
			for _, ipAddr := range h.getInterfaceIPAddresses(extIfaceName, extIfaceMeta) {
				childMapping := proto.Clone(mapping).(*nat.DNat44_StaticMapping)
				childMapping.ExternalIp = ipAddr
				childMapping.ExternalInterface = ""
				childMappings[tag] = append(childMappings[tag], childMapping)
			}
		}
	}

	// do not dump auto-generated child mappings
	for tag, mappings := range entries {
		var filtered []*nat.DNat44_StaticMapping
		for _, mapping := range mappings {
			isChild := false
			for _, child := range childMappings[tag] {
				if proto.Equal(mapping, child) {
					isChild = true
					break
				}
			}
			if !isChild {
				filtered = append(filtered, mapping)
			}
		}
		entries[tag] = filtered
	}
	return entries, nil
}

// nat44StaticMappingDump returns a map of NAT44 static mappings sorted by tags
func (h *NatVppHandler) nat44StaticMappingDump() (entries stMappingMap, err error) {
	if h.ed {
		return h.nat44EdStaticMappingDump()
	} else {
		return h.nat44EiStaticMappingDump()
	}
}

func (h *NatVppHandler) nat44EiStaticMappingLbDump() (entries stMappingMap, err error) {
	return make(stMappingMap), nil
}

func (h *NatVppHandler) nat44EdStaticMappingLbDump() (entries stMappingMap, err error) {
	entries = make(stMappingMap)
	req := &vpp_nat_ed.Nat44LbStaticMappingDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &vpp_nat_ed.Nat44LbStaticMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 lb-static mapping: %v", err)
		}
		if stop {
			break
		}

		// Parse tag (DNAT label)
		tag := strings.TrimRight(msg.Tag, "\x00")
		if _, hasTag := entries[tag]; !hasTag {
			entries[tag] = []*nat.DNat44_StaticMapping{}
		}

		// Prepare localIPs
		var locals []*nat.DNat44_StaticMapping_LocalIP
		for _, localIPVal := range msg.Locals {
			locals = append(locals, &nat.DNat44_StaticMapping_LocalIP{
				VrfId:       localIPVal.VrfID,
				LocalIp:     net.IP(localIPVal.Addr[:]).String(),
				LocalPort:   uint32(localIPVal.Port),
				Probability: uint32(localIPVal.Probability),
			})
		}
		exIPAddress := net.IP(msg.ExternalAddr[:]).String()

		flags := getNat44Flags(msg.Flags)

		// Add mapping into the map.
		mapping := &nat.DNat44_StaticMapping{
			ExternalIp:      exIPAddress,
			ExternalPort:    uint32(msg.ExternalPort),
			LocalIps:        locals,
			Protocol:        h.protocolNumberToNBValue(msg.Protocol),
			TwiceNat:        h.getTwiceNatMode(flags.isTwiceNat, flags.isSelfTwiceNat),
			SessionAffinity: msg.Affinity,
		}
		entries[tag] = append(entries[tag], mapping)
	}

	return entries, nil
}

// nat44StaticMappingLbDump returns a map of NAT44 static mapping with load balancing sorted by tags.
func (h *NatVppHandler) nat44StaticMappingLbDump() (entries stMappingMap, err error) {
	if h.ed {
		return h.nat44EdStaticMappingLbDump()
	} else {
		return h.nat44EiStaticMappingLbDump()
	}
}

func (h *NatVppHandler) nat44EiIdentityMappingDump() (entries idMappingMap, err error) {
	entries = make(idMappingMap)
	childMappings := make(idMappingMap)
	req := &vpp_nat_ei.Nat44EiIdentityMappingDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &vpp_nat_ei.Nat44EiIdentityMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 identity mapping: %v", err)
		}
		if stop {
			break
		}

		// Parse tag (DNAT label)
		tag := strings.TrimRight(msg.Tag, "\x00")
		if _, hasTag := entries[tag]; !hasTag {
			entries[tag] = []*nat.DNat44_IdentityMapping{}
			childMappings[tag] = []*nat.DNat44_IdentityMapping{}
		}

		// resolve interface name
		var (
			found     bool
			ifaceName string
			ifaceMeta *ifaceidx.IfaceMetadata
		)
		if msg.SwIfIndex != NoInterface {
			ifaceName, ifaceMeta, found = h.ifIndexes.LookupBySwIfIndex(uint32(msg.SwIfIndex))
			if !found {
				h.log.Warnf("Interface with index %v not found in the mapping", msg.SwIfIndex)
				continue
			}
		}

		// Add mapping into the map.
		mapping := &nat.DNat44_IdentityMapping{
			IpAddress: net.IP(msg.IPAddress[:]).String(),
			VrfId:     msg.VrfID,
			Interface: ifaceName,
			Port:      uint32(msg.Port),
			Protocol:  h.protocolNumberToNBValue(msg.Protocol),
		}
		entries[tag] = append(entries[tag], mapping)

		if msg.SwIfIndex != NoInterface {
			// collect auto-generated "child" mappings (interface replaced with every assigned IP address)
			for _, ipAddr := range h.getInterfaceIPAddresses(ifaceName, ifaceMeta) {
				childMapping := proto.Clone(mapping).(*nat.DNat44_IdentityMapping)
				childMapping.IpAddress = ipAddr
				childMapping.Interface = ""
				childMappings[tag] = append(childMappings[tag], childMapping)
			}
		}
	}

	// do not dump auto-generated child mappings
	for tag, mappings := range entries {
		var filtered []*nat.DNat44_IdentityMapping
		for _, mapping := range mappings {
			isChild := false
			for _, child := range childMappings[tag] {
				if proto.Equal(mapping, child) {
					isChild = true
					break
				}
			}
			if !isChild {
				filtered = append(filtered, mapping)
			}
		}
		entries[tag] = filtered
	}

	return entries, nil
}

func (h *NatVppHandler) nat44EdIdentityMappingDump() (entries idMappingMap, err error) {
	entries = make(idMappingMap)
	childMappings := make(idMappingMap)
	req := &vpp_nat_ed.Nat44IdentityMappingDump{}
	reqContext := h.callsChannel.SendMultiRequest(req)

	for {
		msg := &vpp_nat_ed.Nat44IdentityMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 identity mapping: %v", err)
		}
		if stop {
			break
		}

		// Parse tag (DNAT label)
		tag := strings.TrimRight(msg.Tag, "\x00")
		if _, hasTag := entries[tag]; !hasTag {
			entries[tag] = []*nat.DNat44_IdentityMapping{}
			childMappings[tag] = []*nat.DNat44_IdentityMapping{}
		}

		// resolve interface name
		var (
			found     bool
			ifaceName string
			ifaceMeta *ifaceidx.IfaceMetadata
		)
		if msg.SwIfIndex != NoInterface {
			ifaceName, ifaceMeta, found = h.ifIndexes.LookupBySwIfIndex(uint32(msg.SwIfIndex))
			if !found {
				h.log.Warnf("Interface with index %v not found in the mapping", msg.SwIfIndex)
				continue
			}
		}

		// Add mapping into the map.
		mapping := &nat.DNat44_IdentityMapping{
			IpAddress: net.IP(msg.IPAddress[:]).String(),
			VrfId:     msg.VrfID,
			Interface: ifaceName,
			Port:      uint32(msg.Port),
			Protocol:  h.protocolNumberToNBValue(msg.Protocol),
		}
		entries[tag] = append(entries[tag], mapping)

		if msg.SwIfIndex != NoInterface {
			// collect auto-generated "child" mappings (interface replaced with every assigned IP address)
			for _, ipAddr := range h.getInterfaceIPAddresses(ifaceName, ifaceMeta) {
				childMapping := proto.Clone(mapping).(*nat.DNat44_IdentityMapping)
				childMapping.IpAddress = ipAddr
				childMapping.Interface = ""
				childMappings[tag] = append(childMappings[tag], childMapping)
			}
		}
	}

	// do not dump auto-generated child mappings
	for tag, mappings := range entries {
		var filtered []*nat.DNat44_IdentityMapping
		for _, mapping := range mappings {
			isChild := false
			for _, child := range childMappings[tag] {
				if proto.Equal(mapping, child) {
					isChild = true
					break
				}
			}
			if !isChild {
				filtered = append(filtered, mapping)
			}
		}
		entries[tag] = filtered
	}

	return entries, nil
}

// nat44IdentityMappingDump returns a map of NAT44 identity mappings sorted by tags.
func (h *NatVppHandler) nat44IdentityMappingDump() (entries idMappingMap, err error) {
	if h.ed {
		return h.nat44EdIdentityMappingDump()
	} else {
		return h.nat44EiIdentityMappingDump()
	}
}

func (h *NatVppHandler) nat44EiInterfaceDump() ([]*nat.Nat44Global_Interface, error) {
	var natIfs []*nat.Nat44Global_Interface

	// dump non-output NAT interfaces
	req1 := &vpp_nat_ei.Nat44EiInterfaceDump{}
	reqContext := h.callsChannel.SendMultiRequest(req1)

	for {
		msg := &vpp_nat_ei.Nat44EiInterfaceDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface: %v", err)
		}
		if stop {
			break
		}

		// Find interface name
		ifName, _, found := h.ifIndexes.LookupBySwIfIndex(uint32(msg.SwIfIndex))
		if !found {
			h.log.Warnf("Interface with index %d not found in the mapping", msg.SwIfIndex)
			continue
		}

		flags := getNat44EiFlags(msg.Flags)

		if flags.eiIfInside {
			natIfs = append(natIfs, &nat.Nat44Global_Interface{
				Name:     ifName,
				IsInside: true,
			})
		} else {
			natIfs = append(natIfs, &nat.Nat44Global_Interface{
				Name:     ifName,
				IsInside: false,
			})
		}
	}

	// dump output NAT interfaces
	var cursor uint32
	for {
		rpcServ, err := h.natEi.Nat44EiOutputInterfaceGet(context.Background(), &vpp_nat_ei.Nat44EiOutputInterfaceGet{
			Cursor: cursor,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface output feature: %v", err)
		}
	RecvLoop:
		for {
			details, reply, err := rpcServ.Recv()
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("failed to dump NAT44 interface output feature: %v", err)
			}
			if reply != nil {
				// TODO: Possible bug in VPP? When VPP contains no NAT interfaces this
				// reply often returns cursor value of 0 repeatedly even though it should return
				// ^uint32(0). So if the reply contains cursor that is the same as cursor
				// from previous reply, return.
				if reply.Cursor == cursor || reply.Cursor == ^uint32(0) {
					return natIfs, nil
				}
				cursor = reply.Cursor
				break RecvLoop
			}
			if details != nil {
				ifName, _, found := h.ifIndexes.LookupBySwIfIndex(uint32(details.SwIfIndex))
				if !found {
					h.log.Warnf("Interface with index %d not found in the mapping", details.SwIfIndex)
				}
				natIfs = append(natIfs, &nat.Nat44Global_Interface{
					Name:          ifName,
					OutputFeature: true,
				})
			}
		}
	}
}

func (h *NatVppHandler) nat44EdInterfaceDump() ([]*nat.Nat44Global_Interface, error) {
	var natIfs []*nat.Nat44Global_Interface

	// dump non-output NAT interfaces
	req1 := &vpp_nat_ed.Nat44InterfaceDump{}
	reqContext := h.callsChannel.SendMultiRequest(req1)

	for {
		msg := &vpp_nat_ed.Nat44InterfaceDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface: %v", err)
		}
		if stop {
			break
		}

		// Find interface name
		ifName, _, found := h.ifIndexes.LookupBySwIfIndex(uint32(msg.SwIfIndex))
		if !found {
			h.log.Warnf("Interface with index %d not found in the mapping", msg.SwIfIndex)
			continue
		}

		flags := getNat44Flags(msg.Flags)

		if flags.isInside {
			natIfs = append(natIfs, &nat.Nat44Global_Interface{
				Name:     ifName,
				IsInside: true,
			})
		} else {
			natIfs = append(natIfs, &nat.Nat44Global_Interface{
				Name:     ifName,
				IsInside: false,
			})
		}
	}

	// dump output NAT interfaces
	var cursor uint32
	for {
		rpcServ, err := h.natEd.Nat44EdOutputInterfaceGet(context.Background(), &vpp_nat_ed.Nat44EdOutputInterfaceGet{
			Cursor: cursor,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface output feature: %v", err)
		}
	RecvLoop:
		for {
			details, reply, err := rpcServ.Recv()
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("failed to dump NAT44 interface output feature: %v", err)
			}
			if reply != nil {
				// TODO: Possible bug in VPP? When VPP contains no NAT interfaces this
				// reply often returns cursor value of 0 repeatedly even though it should return
				// ^uint32(0). So if the reply contains cursor that is the same as cursor
				// from previous reply, return.
				if reply.Cursor == cursor || reply.Cursor == ^uint32(0) {
					return natIfs, nil
				}
				cursor = reply.Cursor
				break RecvLoop
			}
			if details != nil {
				ifName, _, found := h.ifIndexes.LookupBySwIfIndex(uint32(details.SwIfIndex))
				if !found {
					h.log.Warnf("Interface with index %d not found in the mapping", details.SwIfIndex)
				}
				natIfs = append(natIfs, &nat.Nat44Global_Interface{
					Name:          ifName,
					OutputFeature: true,
				})
			}
		}
	}
}

// nat44InterfaceDump dumps NAT44 interface features.
// Deprecated. Functionality moved to Nat44Nat44InterfacesDump. Kept for backward compatibility.
func (h *NatVppHandler) nat44InterfaceDump() ([]*nat.Nat44Global_Interface, error) {
	if h.ed {
		return h.nat44EdInterfaceDump()
	} else {
		return h.nat44EiInterfaceDump()
	}
}

func (h *NatVppHandler) getInterfaceIPAddresses(ifaceName string, ifaceMeta *ifaceidx.IfaceMetadata) (ipAddrs []string) {
	ipAddrNets := ifaceMeta.IPAddresses
	dhcpLease, hasDHCPLease := h.dhcpIndex.GetValue(ifaceName)
	if hasDHCPLease {
		lease := dhcpLease.(*ifs.DHCPLease)
		ipAddrNets = append(ipAddrNets, lease.HostIpAddress)
	}
	for _, ipAddrNet := range ipAddrNets {
		ipAddr := strings.Split(ipAddrNet, "/")[0]
		ipAddrs = append(ipAddrs, ipAddr)
	}
	return ipAddrs
}

// protocolNumberToNBValue converts protocol numeric representation into the corresponding enum
// enum value from the NB model.
func (h *NatVppHandler) protocolNumberToNBValue(protocol uint8) (proto nat.DNat44_Protocol) {
	switch protocol {
	case TCP:
		return nat.DNat44_TCP
	case UDP:
		return nat.DNat44_UDP
	case ICMP:
		return nat.DNat44_ICMP
	default:
		h.log.Warnf("Unknown protocol %v", protocol)
		return 0
	}
}

// protocolNBValueToNumber converts protocol enum value from the NB model into the
// corresponding numeric representation.
func (h *NatVppHandler) protocolNBValueToNumber(protocol nat.DNat44_Protocol) (proto uint8) {
	switch protocol {
	case nat.DNat44_TCP:
		return TCP
	case nat.DNat44_UDP:
		return UDP
	case nat.DNat44_ICMP:
		return ICMP
	default:
		h.log.Warnf("Unknown protocol %v, defaulting to TCP", protocol)
		return TCP
	}
}

func (h *NatVppHandler) getTwiceNatMode(twiceNat, selfTwiceNat bool) nat.DNat44_StaticMapping_TwiceNatMode {
	if twiceNat {
		if selfTwiceNat {
			h.log.Warnf("Both TwiceNAT and self-TwiceNAT are enabled")
			return 0
		}
		return nat.DNat44_StaticMapping_ENABLED
	}
	if selfTwiceNat {
		return nat.DNat44_StaticMapping_SELF
	}
	return nat.DNat44_StaticMapping_DISABLED
}

func getOrCreateDNAT(dnats dnatMap, label string) *nat.DNat44 {
	if _, created := dnats[label]; !created {
		dnats[label] = &nat.DNat44{Label: label}
	}
	return dnats[label]
}

func getNat44Flags(flags nat_types.NatConfigFlags) *nat44EdFlags {
	natFlags := &nat44EdFlags{}
	if flags&nat_types.NAT_IS_EXT_HOST_VALID != 0 {
		natFlags.isExtHostValid = true
	}
	if flags&nat_types.NAT_IS_STATIC != 0 {
		natFlags.isStatic = true
	}
	if flags&nat_types.NAT_IS_INSIDE != 0 {
		natFlags.isInside = true
	}
	if flags&nat_types.NAT_IS_OUTSIDE != 0 {
		natFlags.isOutside = true
	}
	if flags&nat_types.NAT_IS_ADDR_ONLY != 0 {
		natFlags.isAddrOnly = true
	}
	if flags&nat_types.NAT_IS_OUT2IN_ONLY != 0 {
		natFlags.isOut2In = true
	}
	if flags&nat_types.NAT_IS_SELF_TWICE_NAT != 0 {
		natFlags.isSelfTwiceNat = true
	}
	if flags&nat_types.NAT_IS_TWICE_NAT != 0 {
		natFlags.isTwiceNat = true
	}
	return natFlags
}

func getNat44EiFlags(flags vpp_nat_ei.Nat44EiConfigFlags) *nat44EiFlags {
	natFlags := &nat44EiFlags{}
	if flags&vpp_nat_ei.NAT44_EI_STATIC_MAPPING_ONLY != 0 {
		natFlags.eiStaticMappingOnly = true
	}
	if flags&vpp_nat_ei.NAT44_EI_CONNECTION_TRACKING != 0 {
		natFlags.eiConnectionTracking = true
	}
	if flags&vpp_nat_ei.NAT44_EI_OUT2IN_DPO != 0 {
		natFlags.eiOut2InDpo = true
	}
	if flags&vpp_nat_ei.NAT44_EI_ADDR_ONLY_MAPPING != 0 {
		natFlags.eiAddrOnlyMapping = true
	}
	if flags&vpp_nat_ei.NAT44_EI_IF_INSIDE != 0 {
		natFlags.eiIfInside = true
	}
	if flags&vpp_nat_ei.NAT44_EI_IF_OUTSIDE != 0 {
		natFlags.eiIfOutside = true
	}
	if flags&vpp_nat_ei.NAT44_EI_STATIC_MAPPING != 0 {
		natFlags.eiStaticMapping = true
	}
	return natFlags
}

// incIP increments IP address and returns it.
// Based on: https://play.golang.org/p/m8TNTtygK0
func incIP(ip net.IP) net.IP {
	retIP := make(net.IP, len(ip))
	copy(retIP, ip)
	for j := len(retIP) - 1; j >= 0; j-- {
		retIP[j]++
		if retIP[j] > 0 {
			break
		}
	}
	return retIP
}
