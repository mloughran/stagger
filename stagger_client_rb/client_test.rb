require_relative './lib/stagger'

EM.run {
  s = Stagger::Client.new
  
  s.count("connections") {
    rand(100)
  }
  
  
}
