# Implementing GoCable

## Resources

* [Using `anycable` gem in non-Rails apps](https://github.com/anycable/anycable/blob/master/docs/non_rails.md) useful to understand how it hooks into an application with a simpler example than Rails itself.
* [Test connection factory](https://github.com/anycable/anycable/blob/master/spec/support/test_factory.rb) shows protocol responses etc.
* [Ruby Socket/State implementation](https://github.com/anycable/anycable/blob/master/lib/anycable/socket.rb)
* [Ruby RPC handler implementation](https://github.com/anycable/anycable/blob/master/lib/anycable/rpc_handler.rb)
*

## Tools

* `brew install protobuf`
* `gem install anyt`
* `brew install bradleyjkemp/formulae/grpc-tools`

## Development flow

``` sh
buffalo dev # reloading
```

Runnig tests:

```sh
anyt -c "anycable-go --debug --headers cookie,x-api-token --broadcast_adapter http" \
    --target-url="ws://localhost:8080/cable" --skip-rpc
```

Getting debug output from anycable-go:

``` shhs
anycable-go --debug --headers cookie,x-api-token --broadcast_adapter http
```

Running a specific test:

``` sh
anyt -c "sleep 99999999" --target-url="ws://localhost:8080/cable" --skip-rpc --only welcome_test
```

> NOTE: server_restart_test will fail with anycable-go running as a separate process
> because anyt won't be able to kill the process.

## TODO

- [X] Pass all anyt tests
  - [X] multiple_clients_test.rb:33
  - [X] multiple_clients_test.rb:44
  - [X] stop_test.rb:40
  - [X] features/remote_disconnect_test.rb:13
  - [X] server_restart_test.rb:22
  anyt -c "anycable-go --debug --broadcast_adapter http" --target-url="ws://localhost:8080/cable" --skip-rpc --only server_restart_test
  - [X] channel_state_test.rb
- [X] Simple chat
- [X] Estimate what would it take to skip RPC + HTTP broadcast.
- [X] Embed anycable
  - [X] Controller implementation
  - [X] Start node with minimal dependencies
  - [X] Call HandlePubSub or, better, Broadcast when broadcasting.
- [X] Anyt tests pass for standalone server
- [X] Simplified DSL for chat
- [ ] Create a library
  - [X] Rename
  - [X] Restructure code under anycable/
  - [x] Gin example in example/chat
  - [ ] README
    - [ ] Quick start
  - [ ] Write minimal integration test app w/o buffalo.
- [ ] Use in Rally
  - [ ] Show which uexternalser is online
- [ ] Rewrite anyt tests in Go
https://github.com/posener/wstest
- [ ] Docstrings for everything
- [ ] Address all TODOs
- [ ] Contribute http broadcast adapter to anyt

    ``` ruby
    AnyCable.config.broadcast_adapter = :http
    AnyCable.config.http_broadcast_url = 'http://localhost:8090/_broadcast'
    ```

- [ ] Redis broadcast adapter
