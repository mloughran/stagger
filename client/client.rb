require 'em-zeromq'
require 'msgpack'

zmq = EM::ZeroMQ::Context.new(1)

EM.run {
  reg = zmq.socket(ZMQ::PUSH)
  reg.connect("tcp://127.0.0.1:5867")
  
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
        T: "c",
        V: 23.0,
      }

      part2 = {
        N: "api_latency",
        T: "vd",
        D: [3,1,5,9,35].map(&:to_f) # weight, min, max, sx, sxx
      }

      part3 = {
        N: "api_latency",
        T: "v",
        V: 9.0,
      }

      EM.add_timer(1) {
        me.send_msg(*[envelope, part1, part2, part3].map { |p| MessagePack.pack(p) })
      }
    else
      p ["Unknown command", command]
    end
  }
  
  reg.send_msg(MessagePack.pack({
    "Name" => "ruby#{rand(10)}",
    "Address" => mysock,
  }))
}
