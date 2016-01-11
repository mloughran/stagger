# Stagger

A pull-based metrics collection daemon. A bit like collectd, but pull rather than push. Stagger runs on every machine, clients connect to their local stagger, then stagger periodically asks them to submit their contributions to metrics. This model is intended to more closely align the time period on which all processes report than a model where the clients decide when to push their contributions.

Metrics are then aggregated and forwarded to some other service for storage / graphing. Currently [Librato](http://librato.com) is supported, and InfluxDB has been experimented with.

# Testing
````
cd simple_client
bundle exec ruby test.rb
````
# Installing
## On mac os x
Download a go tarball to ~/go
````
cd stagger
make
````
## On linux
````
go get
sudo go build
````
