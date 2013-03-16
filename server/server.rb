require 'em-zeromq'
require 'msgpack'

zmq = EM::ZeroMQ::Context.new(1)

class Client
  def initialize(zmq, address)
    @address = address
    @socket = zmq.socket(ZMQ::PAIR)
    p ["connecting to", address]
    @socket.connect(address)
    @socket.send_msg("hello #{address}")
    
    p 'reg callback'
    @socket.on(:message) { |*parts|
      envelope = parts.shift
      p [:envelope, MessagePack.unpack(envelope.copy_out_string)]
      parts.each { |p| p MessagePack.unpack(p.copy_out_string) }
    }
  end
  
  def request_stats
    @socket.send_msg("send me your stats!")
  end
end

class StatsCollector
  def initialize
    @clients = []
    
    EM.add_periodic_timer(5) {
      request_stats
    }
  end
  
  def add_client(client)
    @clients << client
  end
  
  def request_stats
    @clients.each { |c| c.request_stats }
  end
end

EM.run {
  sc = StatsCollector.new
  
  reg = zmq.socket(ZMQ::PULL)
  reg.bind("tcp://127.0.0.1:2900")
  
  reg.on(:message) { |part|
    reg = MessagePack.unpack(part.copy_out_string)
    p ["got reg", reg]
    sc.add_client Client.new(zmq, reg["Address"])
  }
}
