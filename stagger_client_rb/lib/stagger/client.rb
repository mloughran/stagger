class Client
  # Only 279 google results for "port 5867" :)
  def initialize(reg_address = "tcp://127.0.0.1:5867")
    register(reg_address)

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

  def register(reg_address)
    @zmq_client = ZMQClient.new(reg_address)
    @zmq_client.on(:command, &method(:command))
    @zmq_client.on(:terminated) {
      puts "Connection to client lost, reregistering"
      register(reg_address)
    }
  end

  def command(method, params)
    case method
    when "report_all"
      @zmq_client.send({
        Method: "stats_reply",
        Timestamp: params["Timestamp"],
      }, false)

      @count_callbacks.each do |name, cb|
        value = cb.call
        next if value == 0

        @zmq_client.send({
          N: name.to_s,
          T: "c",
          V: value.to_f, # Currently protocol requires floats...
        }, false)
      end

      @counters.each do |name, count|
        @zmq_client.send({
          N: name.to_s,
          T: "c",
          V: count.to_f, # Currently protocol requires floats...
        }, false)
      end
      @counters = Hash.new { |h,k| h[k] = 0 }

      @value_callbacks.each do |name, cb|
        value = cb.call.to_f

        @zmq_client.send({
          N: name.to_s,
          T: "v",
          V: value,
        }, false)
      end

      @values.each do |name, value_dist|
        @zmq_client.send({
          N: name.to_s,
          T: "vd",
          D: value_dist.to_a.map(&:to_f) # weight, min, max, sx, sxx
        }, false)
      end
      @values = Hash.new { |h,k| h[k] = Distribution.new }

      @zmq_client.send(nil)
    else
      p ["Unknown command", method]
    end
  end
end
