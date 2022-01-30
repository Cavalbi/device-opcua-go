//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package config

import (
	"fmt"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"

	"github.com/spf13/cast"
)

// This file contains example of custom configuration that can be loaded from the service's configuration.toml
// and/or the Configuration Provider, aka Consul (if enabled).
// For more details see https://docs.edgexfoundry.org/2.0/microservices/device/Ch-DeviceServices/#custom-configuration

// Example structured custom configuration types. Must be wrapped with an outer struct with
// single element that matches the top level custom configuration element in your configuration.toml file,
// 'OPCCustom' in this example.
type ServiceConfig struct {
	OPCCustom OPCConfig
}

// OPCConfig is an example of service's custom structured configuration that is specified in the service's
// configuration.toml file and Configuration Provider (aka Consul), if enabled.
type OPCConfig struct {
	DeviceName string
	Policy  string
	Mode  string
	CertFile string
	KeyFile string
	Writable         OPCWritable
}

// OPCWritable defines the service's custom configuration writable section, i.e. can be updated from Consul
type OPCWritable struct {
	Resources string
}

var policies map[string]int = map[string]int{
	"None":           1,
	"Basic128Rsa15":  2,
	"Basic256":       3,
	"Basic256Sha256": 4,
}

var modes map[string]int = map[string]int{
	"None":           1,
	"Sign":           2,
	"SignAndEncrypt": 3,
}

// UpdateFromRaw updates the service's full configuration from raw data received from
// the Service Provider.
func (sw *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false //errors.New("unable to cast raw config to type 'ServiceConfig'")
	}

	*sw = *configuration

	return true
}

// Validate ensures your custom configuration has proper values.
// Example of validating the sample custom configuration
func (oc *OPCConfig) Validate() errors.EdgeX {
	if oc.DeviceName == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,"OPCConfig.DeviceName setting can not be blank", nil)
	}

	if _, ok := policies[oc.Policy]; !ok {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,"OPCConfig.Policy configuration setting must be eiher None, Basic128Rsa15, Basic256 or Basic256Sha256", nil)
	}

	if _, ok := modes[oc.Mode]; !ok {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,"OPCConfig.Mode configuration setting must be either None, Sign or SignAndEncrypt", nil)
	}

	return nil
}

// FetchEndpoint returns the OPCUA endpoint defined in the configuration
func FetchEndpoint(protocols map[string]models.ProtocolProperties) (string, errors.EdgeX) {
	properties, ok := protocols["opcua"]
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("Opcua protocol properties is not defined"), nil)
	}
	endpoint, ok := properties["Endpoint"]
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("Endpoint not found in the Opcua protocol properties"), nil)
	}
	return endpoint, nil
}


// Utility function to cast the result of a reading to one of the type used by EdgeX
func NewResult(req sdkModels.CommandRequest, reading interface{}) (*sdkModels.CommandValue, error) {
	var err error
	castError := "fail to parse %v reading, %v"

	var val interface{}

	switch req.Type {
	case common.ValueTypeBool:
		val, err = cast.ToBoolE(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	case common.ValueTypeString:
		val, err = cast.ToStringE(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	case common.ValueTypeUint8:
		val, err = cast.ToUint8E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	case common.ValueTypeUint16:
		val, err = cast.ToUint16E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	case common.ValueTypeUint32:
		val, err = cast.ToUint32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	case common.ValueTypeUint64:
		val, err = cast.ToUint64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	case common.ValueTypeInt8:
		val, err = cast.ToInt8E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	case common.ValueTypeInt16:
		val, err = cast.ToInt16E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	case common.ValueTypeInt32:
		val, err = cast.ToInt32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	case common.ValueTypeInt64:
		val, err = cast.ToInt64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	case common.ValueTypeFloat32:
		val, err = cast.ToFloat32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	case common.ValueTypeFloat64:
		val, err = cast.ToFloat64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
	default:
		err = fmt.Errorf("return result fail, none supported value type: %v", req.Type)
		return nil, err
	}

	var origin = time.Now().UnixNano() / int64(time.Millisecond)

	return sdkModels.NewCommandValueWithOrigin(req.DeviceResourceName, req.Type, val, origin)

}
