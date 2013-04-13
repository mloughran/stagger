require 'em-zeromq'
require 'msgpack'

module Stagger
  class << self
    def default
      @default ||= Client.new
    end

    def zmq
      @ctx ||= EM::ZeroMQ::Context.new(1)
    end
  end

  class Client
    # Only 279 google results for "port 5867" :)
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

      @count_callbacks = {}
      @value_callbacks = {}
    end

    def count(name, &block)
      raise "Already registered #{name}" if @count_callbacks[name]
      @count_callbacks[name] = block
    end

    def value(name, &block)
      raise "Already registered #{name}" if @value_callbacks[name]
      @value_callbacks[name] = block
    end

    private

    def command(part)
      e = MessagePack.unpack(part.copy_out_string)
      p e

      case (command = e["Method"])
      when "report_all"
        envelope = {
          Method: "stats_reply",
          Timestamp: e["Timestamp"],
        }

        envelope = MessagePack.pack({
          Method: "stats_reply",
          Timestamp: e["Timestamp"],
        })
        @pair.socket.send_string(envelope, ZMQ::NOBLOCK | ZMQ::SNDMORE)

        @count_callbacks.each do |name, cb|
          value = cb.call
          next if value == 0

          msg = MessagePack.pack({
            N: name,
            T: "c",
            V: value.to_f, # Currently protocol requires floats...
          })
          @pair.socket.send_string(msg, ZMQ::NOBLOCK | ZMQ::SNDMORE)
        end

        @value_callbacks.each do |name, cb|
          value = cb.call.to_f

          msg = MessagePack.pack({
            N: name,
            T: "v",
            V: value,
          })
          @pair.socket.send_string(msg, ZMQ::NOBLOCK | ZMQ::SNDMORE)
        end

        @pair.socket.send_string("", ZMQ::NOBLOCK)
      else
        p ["Unknown command", command]
      end
    end
  end
end
