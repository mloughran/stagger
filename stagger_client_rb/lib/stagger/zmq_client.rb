require 'msgpack'

module Stagger
  class ZMQClient
    include EventEmitter

    TIMEOUT = 13

    # We may re-register, in which case we should use a unique socket name
    @@sock_incr = 0

    def initialize(reg_address = "tcp://127.0.0.1:5867")
      # TODO: Should use /var/run or something?
      # This should be configurable?
      @mysock = "ipc:///tmp/stagger_#{Process.pid}_#{@@sock_incr+=1}.zmq"

      reg = Stagger.zmq.socket(ZMQ::PUSH)
      reg.connect(reg_address)
      reg.send_msg(MessagePack.pack({
        "Name" => "ruby#{rand(10)}", # TODO
        "Address" => @mysock,
      }))

      @pair = Stagger.zmq.socket(ZMQ::PAIR)
      @pair.bind(@mysock)

      @pair.on(:message, &method(:command))

      # TODO: emit on pair close?

      reset_activity

      # Periodic check for PAIR connection activity
      # The periodic timer could be more frequent than TIMEOUT, but this is
      # good enough
      @activity_check = EM::PeriodicTimer.new(TIMEOUT) {
        now = Time.now
        since_activity = now - @activity_at
        if since_activity > 2 * TIMEOUT
          puts "Terminating, no activity since #{@activity_at}"
          terminate
        elsif since_activity > TIMEOUT
          ping
        end
      }
    end

    def ping
      # TODO: This isn't right
      @pair.socket.send_string("ping", ZMQ::NOBLOCK)
    end

    def send(reply, final = true)
      flags = final ? ZMQ::NOBLOCK : (ZMQ::NOBLOCK | ZMQ::SNDMORE)
      part = reply ? MessagePack.pack(reply) : ""
      @pair.socket.send_string(part, flags)
    end

    # TODO
    def close

    end

    def terminate
      @activity_check.cancel
      @pair.remove_all_listeners(:message)
      @pair.disconnect(@mysock)
      @pair.socket.unbind(@mysock)

      emit(:terminated)
    end

    private

    def command(part1, part2)
      reset_activity

      method = part1.copy_out_string

      return if method == "pong"

      msgpack_params = part2.copy_out_string
      params = if !msgpack_params.empty?
        MessagePack.unpack(msgpack_params)
      else
        {}
      end

      emit(:command, method, params)
    end

    def reset_activity
      @activity_at = Time.now
    end
  end
end
