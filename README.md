# xuperbench
Benchmark project for XuperUnion (and other blockchain further)

## Pre-requisites

Make sure following requiremants are satisfied:
* Go (version >= 1.11)
* Network (to get depend modules)

## Build xuperbench

Just `make` it

## Usage

### Local mode

Send pressure with single endpoint
* Prepare the environment of xuperunion
* Edit the host address of benchmark config
* Put the xbench binary into xuperunion path
* Run `bin/xbench -c conf/gen.conf`

config meanings:

* gen.conf : profiling normal transfer process
* deal.conf : prepare transaction data before, profiling postTx process
* invoke.conf : profiling contract invoke process

### Remote mode

Send pressure distributedly
* Prepare the environment of xuperunion
* Prepare a environment of Redis (accessable for all endpoint)
* Edit the host address of benchmark config
* Edit the broker(redis) address of benchmark config
* Put the xbench binary into xuperunion path (of each endpoint)
* Run `bin/xbench -worker -c conf/remote.conf` on worker endpoints
* Run `bin/xbench -master -c conf/remote.conf` on master endpoint
* You can edit the conf like gen/deal/invoke to profiling chosen process
