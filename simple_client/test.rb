#!/usr/local/bin/ruby
# -*- coding: utf-8 -*-
require 'bundler/setup'
require 'stagger'
EM.run {
  Stagger.default.register_value(:test){ Random.rand }
}
