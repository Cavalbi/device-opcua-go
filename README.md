# OPC-UA Device Service

This is an implementation of an OPC-UA based Device Service for the open-source edge platform [EdgeX Foundry](https://github.com/edgexfoundry). It allows you to register a new Device Service and make read and write operations on a real connected device using the library [Go Opcua](https://github.com/gopcua/opcua). It is based on [SDK-GO library](https://github.com/edgexfoundry/device-sdk-go) for Jakarta release and the version 2.O of the REST APIs.


## Prerequistes
- Having an Edgex-go deployment running with at least core data, core metadata and core command
- Having an OPC-UA Server to test


## Features

- Execute a read command for one or multiple variables
- Execute a write command for one or multiple variables
- Subscribe and monitor one or multiple variables

# Configuration

Device Services can be configured trough multiple yaml files that define the environment in which they are going to be deployed. Mainly they can configure two things: the device to which we are going to connect and the values to be read along with the commands that can be performed.

## Device Configuration
Inside [opc-simulated-device.toml](https://github.com/Cavalbi/device-opcua-go/blob/master/cmd/res/devices/opc-simulated-device.toml) you can configure the device to which you are going to connect. The name of the device can be set as well as the profile that is going to be used and the actual endpoint to which make the connection

```toml

[[DeviceList]]
  Name = "OPCServerSimulated"
  ProfileName = "OPCServerSimulated"
  Description = "Simulation of an OPC server"
  Labels = [ "test" ]
  [DeviceList.Protocols]
    [DeviceList.Protocols.opcua]
      Endpoint = "opc.tcp://localhost:4841/freeopcua/server/"

```

This is enough to ensure an anonymous connection to the server. If we want to set up a more specific type of connection with security enabled we can go to [configuration.toml](https://github.com/Cavalbi/device-opcua-go/blob/master/cmd/res/configuration.toml) and change the configuration specific for OPC-UA such as policy, mode and path to certificates.

```toml

[OPCCustom]
DeviceName = "OPCServerSimulated"   # Name of existing Device
Policy = "None"                   # Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: None
Mode = "None"                     # Security mode: None, Sign, SignAndEncrypt. Default: None
CertFile = ""                     # Path to cert.pem. Required for security mode/policy != None
KeyFile = ""                      # Path to private key.pem. Required for security mode/policy != None
  [OPCCustom.Writable]
  Resources = "myInt,myFloat"   # list of nodes on the server to read

```

## Device Profile Configuration
[Device Profile](https://github.com/Cavalbi/device-opcua-go/blob/master/cmd/res/profiles/opc-simulated-driver.yaml) let you define the type of values that you can read from OPC-UA server and the actions that you can perform on them.

```yaml
deviceResources:
  -
    name: "MyInt"
    isHidden: false
    description: "Integer variable"
    properties:
        valueType: "Int32"
        readWrite: "R"
    attributes:
      { nodeId: "ns=2;i=2" }
```
Device Resource is a variable that can be read from the server, here we can configure:

- Name: Name of the variable
- isHidden: property that says if the variable is exposed to receive commands (Default:false)
- Description: description of the variable
- valueType: type of variable
- readWrite: type of operations that can be perfomed on it
- attributes: general key/value attributes that can be associated to it (e.g. the variable address)

```yaml
deviceCommands:
  -
    name: "myintcommand"
    isHidden: false
    readWrite: "R"
    resourceOperations:
      - { deviceResource: "MyInt", defaultValue: "false" }
```

Device command is a command that can be perfomed on the device, here we can configure:

- Name: Name of the command that need to be added to the URI endpoint (e.g. http://localhost:59999/api/v2 + "myintcommand")
- isHidden: property that says if the command is exposed (Default:false)
- readWrite: type of operations that can be perfomed on the resourced defined
- resourceOperations: list of device resources that can be read or written

# Build and Run
To build and run the application a Makefile is used.

```bash
make build
```
Build the application

```bash
make run
```
Run the application in nonsecure mode

```bash
make docker
```
Creates a docker image 
