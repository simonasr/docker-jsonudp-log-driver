# docker-jsonudp-log-driver

Log driver for Docker that sends all of the containers output to UDP port in json format. The code is inspired by https://github.com/deep-compute/docker-file-log-driver.

## Features

* Send containers stdout/stderr to a udp port in json format.

## Usage

### Basic usage

Run a container using this plugin:

```
$ docker run --log-driver docker-jsonudp-log-driver --log-opt fpath=/testing/test.log alpine date
Tue Feb 27 06:13:36 UTC 2018
```

### Options

All available options are documented here and can be set via `--log-opt KEY=VALUE`.

|Key|Default|Description|
|---|---|---|
|`host`|127.0.0.1|Host to send logs to|
|`port`|9017|UDP Port to send logs to|

## Hack it

You're more than welcome to hack on this.:-)

```
$ git clone https://github.com/simonasr/docker-jsonudp-log-driver
$ cd docker-jsonudp-log-driver
$ make all
```
This will create and enable plugin.
Verify that plugin is loaded end enabled successfully:
```
docker plugin ls

```
