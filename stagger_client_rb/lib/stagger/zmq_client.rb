require 'msgpack'

module Stagger
  class ZMQClient
    include EventEmitter

    def initialize(reg_address = "tcp://127.0.0.1:5867")
      # TODO: Should use /var/run or something?
      # This should be configurable?
      mysock = "ipc:///tmp/stagger_#{Process.pid}.zmq"

      reg = Stagger.zmq.socket(ZMQ::PUSH)
      reg.connect(reg_address)
      reg.send_msg(MessagePack.pack({
        "Name" => "ruby#{rand(10)}", # TODO
        "Address" => mysock,
      }))

      @pair = Stagger.zmq.socket(ZMQ::PAIR)
      @pair.bind(mysock)

      @pair.on(:message, &method(:command))

      # TODO: emit on pair close?
    end

    def send(reply, final = true)
      flags = final ? ZMQ::NOBLOCK : (ZMQ::NOBLOCK | ZMQ::SNDMORE)
      part = reply ? MessagePack.pack(reply) : ""
      @pair.socket.send_string(part, flags)
    end

    # TODO
    def close

    end

    private

    def command(part)
      params = MessagePack.unpack(part.copy_out_string)
      emit(:command, params.delete("Method"), params)
    end
  end
end
