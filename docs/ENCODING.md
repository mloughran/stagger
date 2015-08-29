# Describes the over-the-wire format for protocol TCPv2

Stagger uses a message-based asynchronous protocol over TCP. This document
describes the payload of these messages.

## Message encoding

Each message is composed of a method and params and is sent back-to-back over
TCP.

- 2 byte magic value: `0x83 0x84` ('ST' in ascii)
- 1 byte protocol version: `0x00`
- 1 byte method (see below for mapping)
- 4 byte length of the upcoming params. Network byte order (which is
  big-endian, see
  https://en.wikipedia.org/wiki/Endianness#Endianness_in_networking)
- Msgpacked params, length as described by previous field.

## Method mapping

Some method can be sent from the server to the client, some from the client to
the server and some both ways.

The associated params structures should be described in the PROTOCOL.md
documentation.

* 0x28: `pair:ping` (both ways)
* 0x29: `pair:pong` (both ways)
* 0x30: `report_all` (server -> client)
* 0x41: `register_process` (client -> server)
* 0x42: `stats_partial` (client -> server)
* 0x43: `stats_complete` (client -> server)

