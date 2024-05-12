# iot-ephemeral-value-store

A very simple value store to store and access your IoT data over http.

## Create Key Pair

### Script

```bash
# Create a upload key and and a download key
uuidgen | sha256sum | (read sha _; echo $sha; echo -n $sha | sha256sum | cut -d " " -f1)

# e.g.
# 1326a51edc413cbd5cb09961e6fc655b8e30aca8eb4a46be2e6aa329da31709f
# 4698f8edcc24806c2e57b9db57e7958299982a0cc325b00163300d0cb2828a57
```


### Web

```http
GET https://your-server.com/kp

200 OK
{
  "upload-key": "1326a51edc413...",
  "download-key": "4698f8edcc24..."
}
```

## Upload Values

```http
GET https://your-server.com/{upload-key}/?temp=23&hum=43

200 OK
```

## Download Values

```http
GET https://your-server.com/{download-key}/json

200 OK
{
  "temp": "23",
  "hum": "43"
}
```

## CLI

- `--persist-values-for`: Duration for which the values are stored before they are deleted. Example: `1d` for one day, `2h` for two hours.
- `--store`: Path to the directory where the values will be stored.
- `--port`: The port number on which the server will listen.

```
iot-ephemeral-value-store-server --persist-values-for 1d --store ~/iot-ephemeral-value-store --port 8080
```