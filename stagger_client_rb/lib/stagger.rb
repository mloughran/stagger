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

  class Distribution
    attr_reader :weight, :sum_x, :sum_x2, :min, :max

    def initialize
      @weight, @sum_x, @sum_x2, @min, @max = 0, 0, 0, nil, nil
    end

    def add(x, weight = 1)
      @weight += weight
      @sum_x += x * weight
      @sum_x2 += x**2 * weight
      @min = [@min, x].compact.min
      @max = [@max, x].compact.max
      self
    end

    def mean
      @weight > 0 ? @sum_x.to_f / @weight : nil
    end

    def to_a
      [@weight, @min, @max, @sum_x, @sum_x2]
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

      @counters = Hash.new { |h,k| h[k] = 0 }
      @values = Hash.new { |h,k| h[k] = Distribution.new }
    end

    def register_count(name, &block)
      raise "Already registered #{name}" if @count_callbacks[name]
      @count_callbacks[name.to_sym] = block
    end

    def register_value(name, &block)
      raise "Already registered #{name}" if @value_callbacks[name]
      @value_callbacks[name.to_sym] = block
    end

    def incr(name, count = 1)
      @counters[name.to_sym] += count
    end

    def value(name, value, weight = 1)
      @values[name.to_sym].add(value, weight)
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
            N: name.to_s,
            T: "c",
            V: value.to_f, # Currently protocol requires floats...
          })
          @pair.socket.send_string(msg, ZMQ::NOBLOCK | ZMQ::SNDMORE)
        end

        @counters.each do |name, count|
          msg = MessagePack.pack({
            N: name.to_s,
            T: "c",
            V: count.to_f, # Currently protocol requires floats...
          })
          @pair.socket.send_string(msg, ZMQ::NOBLOCK | ZMQ::SNDMORE)
        end
        @counters = Hash.new { |h,k| h[k] = 0 }

        @value_callbacks.each do |name, cb|
          value = cb.call.to_f

          msg = MessagePack.pack({
            N: name.to_s,
            T: "v",
            V: value,
          })
          @pair.socket.send_string(msg, ZMQ::NOBLOCK | ZMQ::SNDMORE)
        end

        @values.each do |name, value_dist|
          msg = MessagePack.pack({
            N: name.to_s,
            T: "vd",
            D: value_dist.to_a.map(&:to_f) # weight, min, max, sx, sxx
          })
          @pair.socket.send_string(msg, ZMQ::NOBLOCK | ZMQ::SNDMORE)
        end
        @values = Hash.new { |h,k| h[k] = Distribution.new }

        @pair.socket.send_string("", ZMQ::NOBLOCK)
      else
        p ["Unknown command", command]
      end
    end
  end
end
