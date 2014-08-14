#!/usr/local/bin/ruby
# -*- coding: utf-8 -*-
require 'bundler/setup'
require 'stagger'
EM.run {
  (1..20).each{|i|
    Stagger.default.register_value("test#{i%3}.floo#{i}.nox".to_sym){
      puts "fetching data for test.floo#{i}"
      Random.rand
    }
  }
}
