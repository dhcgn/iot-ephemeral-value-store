# Home Assistant Integration

The iot-ephemeral-value-store integrates seamlessly with Home Assistant, allowing you to upload sensor data from your Home Assistant instance to the server. This enables you to share sensor data externally or access it from other systems without exposing your Home Assistant instance directly.

## Use Cases

- Share sensor readings with external applications or services
- Create a simple API for mobile apps to access your sensor data
- Aggregate data from multiple Home Assistant instances
- Provide read-only access to sensor data without exposing Home Assistant credentials
- Create temporary data endpoints that expire automatically

## Configuration

### Step 1: Configure REST Commands

Add the following to your Home Assistant `configuration.yaml`:

```yaml
rest_command:
  # Upload a single value
  upload_simple_value:
    url: "https://your-server.com/u/{{ key }}/?value={{ value }}"

  # Upload a value to a specific folder (for organizing multiple values)
  upload_simple_value_to_folder:
    url: "https://your-server.com/patch/{{ key }}/{{ folder }}/?value={{ value }}"
```

### Step 2: Create Automations

Create automations to upload sensor data when it changes:

```yaml
alias: Upload Temperature
description: "Upload pool temperature to iot-ephemeral-value-store"
triggers:
  - type: temperature
    platform: device
    device_id: YOUR_DEVICE_ID
    entity_id: sensor.pool_temperature_temperature
conditions: []
actions:
  # Upload single value
  - action: rest_command.upload_simple_value
    metadata: {}
    data:
      key: "1e1c7e5f220d2eee5ebbfd1428b84aaf1570ca4f88105a81feac901850b20a77"
      value: "{{ states('sensor.pool_temperature_temperature') }}"

  # Upload value to folder (allows multiple uploads to be combined)
  - action: rest_command.upload_simple_value_to_folder
    metadata: {}
    data:
      key: "abb0c54ca2b3dbecc3c781f2de6d0e1c62c6cd8a59c82932ac68c746416d8134"
      folder: Water
      value: "{{ states('sensor.pool_temperature_temperature') }}"
mode: single
```

## Advanced Examples

### Upload Multiple Sensor Values Using Patch

Use the patch feature to organize related sensor data:

```yaml
alias: Upload All Pool Sensors
description: "Upload multiple pool sensor readings to organized folders"
triggers:
  - platform: time_pattern
    minutes: "/5"  # Every 5 minutes
conditions: []
actions:
  # Upload water temperature
  - action: rest_command.upload_simple_value_to_folder
    data:
      key: "YOUR_UPLOAD_KEY"
      folder: Water
      value: "{{ states('sensor.pool_water_temperature') }}"

  # Upload air temperature
  - action: rest_command.upload_simple_value_to_folder
    data:
      key: "YOUR_UPLOAD_KEY"
      folder: Air
      value: "{{ states('sensor.pool_air_temperature') }}"

  # Upload pH level
  - action: rest_command.upload_simple_value_to_folder
    data:
      key: "YOUR_UPLOAD_KEY"
      folder: pH
      value: "{{ states('sensor.pool_ph') }}"
mode: single
```

After running this automation, you can retrieve all values with a single call:
```bash
curl https://your-server.com/d/YOUR_DOWNLOAD_KEY/json
```

Returns:
```json
{
  "Water": {
    "value": "25.3",
    "timestamp": "2024-12-29T18:51:07Z"
  },
  "Air": {
    "value": "23.1",
    "timestamp": "2024-12-29T18:51:08Z"
  },
  "pH": {
    "value": "7.2",
    "timestamp": "2024-12-29T18:51:09Z"
  },
  "timestamp": "2024-12-29T18:51:09Z"
}
```

### Upload Multiple Parameters in One Call

You can also upload multiple values in a single REST command:

```yaml
rest_command:
  upload_multiple_values:
    url: "https://your-server.com/patch/{{ key }}/{{ folder }}/?temp={{ temp }}&humidity={{ humidity }}&pressure={{ pressure }}"
```

Usage in automation:
```yaml
actions:
  - action: rest_command.upload_multiple_values
    data:
      key: "YOUR_UPLOAD_KEY"
      folder: LivingRoom
      temp: "{{ states('sensor.living_room_temperature') }}"
      humidity: "{{ states('sensor.living_room_humidity') }}"
      pressure: "{{ states('sensor.living_room_pressure') }}"
```

### Conditional Upload Based on State

Upload data only when certain conditions are met:

```yaml
alias: Upload Temperature When High
description: "Upload temperature only when it exceeds threshold"
triggers:
  - platform: numeric_state
    entity_id: sensor.outdoor_temperature
    above: 30
conditions: []
actions:
  - action: rest_command.upload_simple_value_to_folder
    data:
      key: "YOUR_UPLOAD_KEY"
      folder: HighTemp
      value: "{{ states('sensor.outdoor_temperature') }}"
mode: single
```

## Setup Guide

1. **Generate Key Pair**: Visit your iot-ephemeral-value-store server at `/` to generate a key pair, or use the `/kp` endpoint:
   ```bash
   curl https://your-server.com/kp
   ```

2. **Configure REST Commands**: Add the REST commands to your `configuration.yaml` and restart Home Assistant.

3. **Create Automations**: Set up automations to upload data based on your needs (time intervals, state changes, etc.).

4. **Test Upload**: Manually trigger your automation and verify the data:
   ```bash
   curl https://your-server.com/d/YOUR_DOWNLOAD_KEY/json
   ```

5. **Share Download Key**: Provide the download key to applications or users who need read-only access to the data.

## Best Practices

- **Use Separate Keys**: Create different key pairs for different data types or locations
- **Consider Upload Frequency**: Balance between data freshness and server load
- **Leverage Patch Feature**: Use folders to organize related sensor readings
- **Monitor Data Retention**: Ensure your upload frequency matches the server's `-persist-values-for` setting
- **Secure Upload Keys**: Keep upload keys private; only share download keys
- **Use Conditions**: Add conditions to avoid unnecessary uploads (e.g., only upload when values change significantly)

## Troubleshooting

- **Data Not Appearing**: Verify the upload key is correct and check Home Assistant logs for REST command errors
- **Data Expires Too Quickly**: Check the server's data retention setting (`-persist-values-for` flag)
- **Automation Not Triggering**: Verify triggers and conditions in your automation
- **Invalid Key Error**: Ensure the upload key is a valid 256-bit hex string (64 characters)
