// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a simple example implementation of
// ProtocolDriver interface.
//
package driver

import (
	"context"
	"fmt"
	"reflect"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/edgexfoundry/device-opcua-go/config"

	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

type OPCDriver struct {
	lc            logger.LoggingClient
	asyncCh       chan<- *sdkModels.AsyncValues
	deviceCh      chan<- []sdkModels.DiscoveredDevice
	serviceConfig *config.ServiceConfig
}

// Initialize performs protocol-specific initialization for the device
// service.
func (s *OPCDriver) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModels.AsyncValues, deviceCh chan<- []sdkModels.DiscoveredDevice) error {
	s.lc = lc
	s.asyncCh = asyncCh
	s.deviceCh = deviceCh
	s.serviceConfig = &config.ServiceConfig{}

	ds := service.RunningService()

	if err := ds.LoadCustomConfig(s.serviceConfig, "OPCCustom"); err != nil {
		return fmt.Errorf("unable to load 'OPCCustom' custom configuration: %s", err.Error())
	}

	lc.Infof("Custom config is: %v", s.serviceConfig.OPCCustom)

	if err := s.serviceConfig.OPCCustom.Validate(); err != nil {
		return fmt.Errorf("'OPCCustom' custom configuration validation failed: %s", err.Error())
	}

	if err := ds.ListenForCustomConfigChanges(
		&s.serviceConfig.OPCCustom.Writable,
		"OPCCustom/Writable", s.ProcessCustomConfigChanges); err != nil {
		return fmt.Errorf("unable to listen for changes for 'OPCCustom.Writable' custom configuration: %s", err.Error())
	}

	return nil
}

// ProcessCustomConfigChanges ...
func (s *OPCDriver) ProcessCustomConfigChanges(rawWritableConfig interface{}) {
	updated, ok := rawWritableConfig.(*config.OPCWritable)
	if !ok {
		s.lc.Error("unable to process custom config updates: Can not cast raw config to type 'OPCWritable'")
		return
	}

	s.lc.Info("Received configuration updates for 'OPCCustom.Writable' section")

	previous := s.serviceConfig.OPCCustom.Writable
	s.serviceConfig.OPCCustom.Writable = *updated

	if reflect.DeepEqual(previous, *updated) {
		s.lc.Info("No changes detected")
		return
	}

	// Now check to determine what changed.
	// In this example we only have the one writable setting,
	// so the check is not really need but left here as an example.
	// Since this setting is pulled from configuration each time it is need, no extra processing is required.
	// This may not be true for all settings, such as external host connection info, which
	// may require re-establishing the connection to the external host for example.
	if previous.Resources != updated.Resources {
		s.lc.Infof("Resources changed to: %s", updated.Resources)
	}
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (s *OPCDriver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest) (res []*sdkModels.CommandValue, err error) {
	s.lc.Debugf("OPCDriver.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes)

	// create device client and open connection
	endpoint, err := config.FetchEndpoint(protocols)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	client := opcua.NewClient(endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
	if err := client.Connect(ctx); err != nil {
		s.lc.Warnf("Driver.HandleReadCommands: Failed to connect OPCUA client, %s", err)
		return nil, err
	}
	defer client.Close()

	var responses = make([]*sdkModels.CommandValue, len(reqs))

	for i, req := range reqs {
		// handle every reqs
		res, err := s.handleReadCommandRequest(client, req)
		if err != nil {
			s.lc.Errorf("Driver.HandleReadCommands: Handle read commands failed: %v", err)
			return responses, err
		}
		responses[i] = res
	}

	return responses, nil
}

func (s *OPCDriver) handleReadCommandRequest(deviceClient *opcua.Client, req sdkModels.CommandRequest) (*sdkModels.CommandValue, error) {
	nodeID, ok := req.Attributes["nodeId"]
	if !ok {
		return nil, fmt.Errorf("attribute nodeID does not exist")
	}

	id, err := ua.ParseNodeID(nodeID.(string))
	if err != nil {
		return nil, fmt.Errorf("Driver.handleReadCommands: Invalid node id=%s; %v", nodeID, err)
	}

	request := &ua.ReadRequest{
		MaxAge: 2000,
		NodesToRead: []*ua.ReadValueID{
			{NodeID: id},
		},
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}
	resp, err := deviceClient.Read(request)
	if err != nil {
		return nil, fmt.Errorf("Driver.handleReadCommands: Read failed: %s", err)
	}
	if resp.Results[0].Status != ua.StatusOK {
		return nil, fmt.Errorf("Driver.handleReadCommands: Status not OK: %v", resp.Results[0].Status)
	}

	// make new result
	reading := resp.Results[0].Value.Value()

	return config.NewResult(req, reading)

}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource.
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (s *OPCDriver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest,
	params []*sdkModels.CommandValue) error {
	// Not yet Implemented

	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *OPCDriver) Stop(force bool) error {
	// Then Logging Client might not be initialized
	if s.lc != nil {
		s.lc.Debugf("SimpleDriver.Stop called: force=%v", force)
	}
	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (s *OPCDriver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf("a new Device is added: %s", deviceName)
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (s *OPCDriver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf("Device %s is updated", deviceName)
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (s *OPCDriver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	s.lc.Debugf("Device %s is removed", deviceName)
	return nil
}

// Discover triggers protocol specific device discovery, which is an asynchronous operation.
// Devices found as part of this discovery operation are written to the channel devices.
func (s *OPCDriver) Discover() {
	// Not yet Implemented

}
