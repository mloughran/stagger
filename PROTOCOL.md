## Registration (proc -> stats)

On startup, clients bind a ZMQ PAIR socket which will be used to respond to stats requests. They then send this address string to the stats server via a ZMQ push socket. The address may be an ipc or tcp socket.

    Address: ipc:///some/unix/socket

After the initial registration this push socket MAY be closed.

The stats server will connect to the PAIR socket and send commands.

## Requesting stats (stats -> proc)

When the stats server wishes to receive stats it will request them from the client by sending a message on the PAIR socket.

To request all stats from the client it just sends

    Method: report_all
    Timestamp: 123456 [unix timestamp in seconds]

To request a subset of stats it should send

    Method: report_list
    Timestamp: 123456
    Stats: ["list", "of", "stat", "names"]

The client should reply to a stats request a stats reply.

## Sending stats (proc -> stats)

Stats should be sent as a multipart ZMQ message. The first part (called the envelope), and then one or more message parts containing stats. Each part may contain one or more stats.

Envelope example:

    Method: stats_reply
    Timestamp: 123456 [the unix timestamp sent by the request]

Stats example

    [{
      Name: connections
      Type: value
      Val: 23
    }, ...]

The reasoning behing this design is to allow clients to optimise the sending of stats. A process which sends a small number of small stats may send them all in one part (to reduce the number of kernel calls), while a process that sends a lot of data may wish to send stats in multiple parts to avoid using lots of memory or blocking an evented process.

A process MAY decide that it is overloaded or does not wish to reply with stats for some reason. In this case it SHOULD send a reply to that effect. This lets the stats server know not to wait for stats from this process, and allows it to propagate aggregated stats from other processes without delay.

    Method: skipping
    Timestamp: 1234
    Reason: "Optional reason, will be printed in stats logs"

## Different types of stats

    value
    count
    see aggregations in DTrace

Should we be able to send min & max separately? Maybe distrib? What about buckets.

It may be desirable to segment stats. This is done as follows

    Name: connections
    Segments: (type, [(HTTP, 23), (HTTPS, 13)])

A stat may be segmented by multiple parameters. For example

    Name: connections
    Segments: ((type, user), [((HTTP, 42), 13), ((HTTP, 12), 11)])

The server will automatically aggregate this data in order to get e.g. connections per user any per type automatically.

