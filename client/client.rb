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
    e = MessagePack.unpack(part.copy_out_string)
    p e
    
    case (command = e["Method"])
    when "report_all"
      p "sending stats"
      envelope = {
        Method: "stats_reply",
        Timestamp: e["Timestamp"],
      }
      
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
