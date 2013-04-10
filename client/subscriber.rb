require 'em-zeromq'
require 'msgpack'

zmq = EM::ZeroMQ::Context.new(1)

EM.run {
  sub = zmq.socket(ZMQ::SUB)
  
  sub.connect("tcp://localhost:5563")
  
  # Subscribe to everything
  prefix = ""
  sub.subscribe(prefix)
  
  sub.on(:message) { |channel, *parts|
    p [:channel, channel.copy_out_string]
    p [:parts, parts.map { |p| MessagePack.unpack(p.copy_out_string) }]
  }
}
