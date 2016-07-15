# Stagger protocol (TCPv2)

This document describes how the stagger server and clients interact using the
TCPv2 protocol. All other protocols are deprecated and will be removed in the
short future.

## Connection

The client connects to the server using a TCP socket. The default address is
`localhost:5865`.

The client and server communicate using an asynchronous message-based RPC
protocol. The encoding of each message is described in the
[ENCODING](./ENCODING.md) doc.

When the client is connected it is recommended to send a `register_process`
call so that stagger can identify the client in the logs.

## Methods

Here is the list of all the methods and in which directions they are sent.

### `pair:ping` (Both)

When either end hasn't received a message "for a while" it can request a
livelyhood-checking message.

Params:

```json
{}
```

### `pair:pong` (Both)

Sent back to the `pair:ping` requesting party.

Params:

```json
{}
```

### `report_all` (server -> client)

When the stats server wishes to receive stats it will request them from the
client by sending this message. The params requires a Timestamp in unix epoch
seconds.

The client currently has 1 second to return all it's metrics using
`stats_partial` and `stats_all` methods.

Example:

```json
{
  "Timestamp": 123456
}
```

### `register_process` (client -> server)

Can be sent by the client to identify itself to the server. The params
contains an arbitrary list of string to string tags.

Example:

```json
{
  "Tags": {
    "pid": "1234",
    "cmd": "./hello"
  }
}
```

### `stats_partial` (client -> server)

Used to send partial metrics after the client gets a `report_all` call from
the server. This method allows to send stats back to the server in multiple
parts.

It's useful when some stats might be retrieved in an unreliable fashion and
the client wants to make sure that the rest of the stats are collected before
the 1 second deadline. Or if the client has a lot of stats to send and want to
transmit them in a chunked fashion.

To mark the completion of all `stats_partial` calls, the `stats_complete`
method has to be used (possibly with no values).

Example:

```json
{
  "Timestamp": 123456,
  "Counts": [],
  "Dists": []
}
```

### `stats_complete` (client -> server)

Used to send all remaining metrics after the client gets a `report_all` call
from the server.

The params can be empty if the goal is to close a sequence of `stats_partial`
calls.

Example:

```json
{
  "Timestamp": 123456,
  "Counts": [],
  "Dists": []
}
```

## Metrics

There are two different type of metrics that stagger captures ; Counts
(increments) and Distributions.

### Counts

Counters are used to describe increments only.

```json
{
  "Name": "key2",
  "Count": 0.5,
}
```

### Distributions

Used to describe the distribution of values (gauges) during the capture
interval.

The Dist key contains 5 float64 elements which map to
 (Weight, Min, Max, Sum, Sum * 2)

```json
{
  "Name": "key3",
  "Dist": [0.1, 0.2, 0.3, 0.4, 0.5],
}
```

## Encoding of tags

For backward-compatiblity reasons, tags are encoded as part of the metric
name. On reception stagger will decode them again.

Format: "key,tag1=value1,tag2=value2"

To avoid creating multiple keys the tags are supposed to be sorted
alphabetically.

Tags and values are strings that don't contain commas or equal characters.

