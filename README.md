== On mac os x
Download a go tarball to ~/go
````
cd stagger
export GOROOT=$HOME/go
export PATH=$PATH:$GOROOT/bin
export GOPATH=$PWD
export GOBIN=$PWD
brew install zeromq
go build
````
== On linux
````
apt-get install go libzmq3
go get
sudo go build
````