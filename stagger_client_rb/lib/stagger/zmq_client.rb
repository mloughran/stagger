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

    def command(part1, part2)
      method = part1.copy_out_string

      msgpack_params = part2.copy_out_string
      params = if !msgpack_params.empty?
        MessagePack.unpack(msgpack_params)
      else
        {}
      end

      emit(:command, method, params)
    end
  end
end
