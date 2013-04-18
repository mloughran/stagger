# coding: utf-8
lib = File.expand_path('../lib', __FILE__)
$LOAD_PATH.unshift(lib) unless $LOAD_PATH.include?(lib)
require 'stagger/version'

Gem::Specification.new do |spec|
  spec.name          = "stagger"
  spec.version       = Stagger::VERSION
  spec.authors       = ["Martyn Loughran"]
  spec.email         = ["me@mloughran.com"]
  spec.description   = %q{Stats aggregation}
  spec.summary       = %q{Stats aggregation}
  spec.homepage      = ""
  spec.license       = "MIT"

  spec.files         = `git ls-files`.split($/)
  spec.executables   = spec.files.grep(%r{^bin/}) { |f| File.basename(f) }
  spec.test_files    = spec.files.grep(%r{^(test|spec|features)/})
  spec.require_paths = ["lib"]

  spec.add_dependency "em-zeromq" # TODO 0.4.2
  spec.add_dependency "msgpack"

  spec.add_development_dependency "bundler", "~> 1.3"
  spec.add_development_dependency "rake"
end
