# Basic TTN MQTT connector in Go.

This Simple connector subscribe to The Things Network backend and transmit it
to the Mqtt Broker.

We are expecting a JSON schema as follow:

```javascript
{
   "app_id": "testfeather",
   "dev_id": "feather",
   "hardware_serial": "0083F53891xxxxxx",
   "port": 1,
   "counter": 0,
   "is_retry": true,
   "payload_raw": "MjkuNzA=",
   "payload_fields": {
     "temperature": "29.70"
   },
   "metadata": {
     "time": "2017-08-21T18:57:00.775133046Z",
     "frequency": 868.1,
     "modulation": "LORA",
     "data_rate": "SF7BW125",
     "coding_rate": "4/5",
     "gateways": [
       {
         "gtw_id": "eui-b827ebfffxxxxxxxx",
         "timestamp": 2878478064,
         "time": "",
         "channel": 0,
         "rssi": -49,
         "snr": 9,
         "latitude": 43.6115,
         "longitude": 1.4359,
         "altitude": 162
       }
     ]
   },
   "downlink_url": "https://integrations.thethingsnetwork....org/"
}
```

MQTT topic is by default a concatenation of *app_id* and *dev_id* with user defined prefix and suffix.

This connector is designed to be modified and improved with regards to the user application.

## Configuration

Configuration file allow the user to specify both MQTT broker configuration.

A systemd unit file is present and can be used to launch the connector at startup.

## Usage

### Test

```sh
go run ttnMqttConnector.go --config config.json
```
### Deploy

```sh
go build ttnMqttConnector.go

sudo cp ttnMqttConnector /usr/bin/
sudo cp ttnMqttConnector.service /usr/lib/systemd/system/

sudo mkdir /etc/ttnMqttConnector/
sudo cp config.json /etc/ttnMqttConnector/

sudo systemctl start ttnMqttConnector.service
```

## Requirements

Requires:

    "fmt"
    "os"
    "log"
    "strconv"
    "bytes"
    "encoding/json"
    "flag"
    "time"
    "os/signal"

    "github.com/eclipse/paho.mqtt.golang"
