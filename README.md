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
export GOROOT=$HOME/go
export PATH=$PATH:$GOROOT/bin
export GOPATH=$PWD
export GOBIN=$PWD
brew install zeromq
make
````
## On linux
````
apt-get install go libzmq3
go get
sudo go build
````
# Making/importing certificates
Use easy-rsa3 to make a client certificate

## On linux
https://code.google.com/p/chromium/wiki/LinuxCertManagement

## On mac
````
security import ca.crt -k ~/Library/Keychains/login.keychain
````