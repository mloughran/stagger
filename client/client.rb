require 'em-zeromq'
require 'msgpack'

zmq = EM::ZeroMQ::Context.new(1)

EM.run {
  reg = zmq.socket(ZMQ::PUSH)
  reg.connect("tcp://127.0.0.1:2900")
  
  me = zmq.socket(ZMQ::PAIR)
  # TODO: Should use /var/run or something - how do permissions work there?
  mysock = "ipc:///tmp/stagger_#{Process.pid}.zmq"
  me.bind(mysock)
  
  me.on(:message) { |part|
    p part.copy_out_string
    
    case (command = part.copy_out_string)
    when "send me your stats!", "Stats please!"
      p "sending stats"
      envelope = {Method: "stats"}
      
      part1 = {
        N: "connections",
        T: "value",
        V: 23,
      }

      me.send_msg(*[envelope, part1].map { |p| MessagePack.pack(p) })
    else
      p ["Unknown command", command]
    end
  }
  
  reg.send_msg(MessagePack.pack({
    "Address" => mysock
  }))
}
