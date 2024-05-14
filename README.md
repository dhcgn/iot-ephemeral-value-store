[![Go](https://github.com/dhcgn/iot-ephemeral-value-store/actions/workflows/build_and_test.yml/badge.svg)](https://github.com/dhcgn/iot-ephemeral-value-store/actions/workflows/build_and_test.yml)

# iot-ephemeral-value-store

This project provides a simple HTTP server that offers ephemeral storage for IoT data. It generates unique key pairs for data upload and retrieval, stores data temporarily based on a configurable duration, and allows data to be fetched in both JSON and plain text formats.

## Motivation

I want to share sensor readings from my home with other devices or on the web. This should be done using a service with a very simple and easy http interface, but secure by design.

Very simple and easy means in this context:

- Should work simple IoT devices
- Only HTTP GET
- KeyPair is very easy to generate with standard tool or with online generators

### A example using a shelly temperature sensor

1. Obtain a key pair for upload and download as http links (these keys are linked together, but you can share the Download-Key with compromise the Upload-Key.)
   1. Upload: `https://iot.my-server/2f27c.../`
   1. Download: `https://iot.my-server/234ca1.../`
1. Go to the Shelly web interface, add the link to *action* at event *temperature reading*, add the placeholder for the temperature
   1. e.g. `https://iot.my-server/2f27c.../?temp={{temp}}` where `{{temp}}` will be replaced with `18.8`
1. Share the links with friends or in websites
   1. JSON: `https://iot.my-server/234ca1.../json` returns `{"temp":"18.8","timestamp":"2024-05-14T11:37:21Z"}`
   1. Plain: `https://iot.my-server/234ca1.../plain/temp` returns `18.8`

### Infrastucture

![](docs/infra.drawio.svg)

## Features

- **Key Pair Generation**: Generate unique upload and download keys for secure data handling.
- **Data Upload**: Upload data with a simple GET request, using the generated upload key.
- **Data Retrieval**: Retrieve stored data using the download key, either as JSON or plain text for specific data fields.

## HTTP Calls

### Create Key Pair

The upload key is just random data with a length of 256bit encoded in hex, the download key is derived a each upload time. The download key is just a hashed upload key with sha256.

#### Web

```http
GET https://your-server.com/kp

200 OK
{
  "upload-key": "1326a51edc413...",
  "download-key": "4698f8edcc24..."
}
```

#### Script

```bash
# Create a upload key and and a download key
uuidgen | sha256sum | (read sha _; echo $sha; echo -n $sha | sha256sum | cut -d " " -f1)

# e.g.
# 1326a51edc413cbd5cb09961e6fc655b8e30aca8eb4a46be2e6aa329da31709f
# 4698f8edcc24806c2e57b9db57e7958299982a0cc325b00163300d0cb2828a57
```

or

```bash
# Create a 256-bit (32 bytes) random data encoded in hex
upload_key=$(head -c 32 /dev/urandom | xxd -p -c 256)

# Derive a secondary key, such as a download key, by hashing the upload key using sha256sum
download_key=$(echo -n $upload_key | sha256sum | cut -d " " -f1)

echo "Upload Key: $upload_key"
echo "Download Key: $download_key"

# Example output:
# Upload Key: 1326a51edc413cbd5cb09961e6fc655b8e30aca8eb4a46be2e6aa329da31709f
# Download Key: 4698f8edcc24806c2e57b9db57e7958299982a0cc325b00163300d0cb2828a57
```


### Upload Values

- {upload-key} must be a hex 256bit representation

```http
GET https://your-server.com/{upload-key}/?temp=23&hum=43

200 OK
{
  "download_url": "http://127.0.0.1:8080/{download-key}/json",
  "message": "Data uploaded successfully",
  "parameter_urls": {
    "hum": "http://127.0.0.1:8080/{download-key}/plain/hum",
    "temp": "http://127.0.0.1:8080/{download-key}/plain/temp"
  }
}
```

### Download Values

```http
GET https://your-server.com/{download-key}/json

200 OK
{
  "temp": "23",
  "hum": "43"
}
```

```http
GET https://your-server.com/{download-key}/plain/hum

200 OK
43
```

## Gettings startet

### CLI

- `-persist-values-for`: Duration for which the values are stored before they are deleted. Example: `1d` for one day, `2h` for two hours.
- `-store`: Path to the directory where the values will be stored.
- `-port`: The port number on which the server will listen.

```
iot-ephemeral-value-store-server -persist-values-for 1d -store ~/iot-ephemeral-value-store -port 8080
```

### Install the Server as a System Service

- Run the installation script as root:
```bash
sudo ./install-service.sh /path/to/iot-ephemeral-value-store-server
```