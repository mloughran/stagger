require_relative './lib/stagger'

EM.run {
  s = Stagger::Client.new

  s.register_count(:test_count_cb) {
    rand(100)
  }

  s.register_value(:test_value_cb) {
    rand(100)
  }

  EM.add_periodic_timer(1) {
    s.incr(:test_count)
  }

  EM.add_periodic_timer(1) {
    s.value(:test_value, rand(50))
  }
}
