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

``` shhs
ANYCABLE_HEADERS="cookie,x-api-token" anycable-go --debug
```

Running a specific test:

``` sh
anyt -c "sleep 99999999" --target-url="ws://localhost:8080/cable" --skip-rpc --only welcome_test
```

Running the full suite:

``` sh
anyt -c "sleep 99999999" --target-url="ws://localhost:8080/cable" --skip-rpc
```
## TODO

- [ ] Pass all anyt tests
- [ ] Contribute http broadcast adapter to anyt

    ``` ruby
    AnyCable.config.broadcast_adapter = :http
    AnyCable.config.http_broadcast_url = 'http://localhost:8090/_broadcast'
    ```

- [ ] Implement ChannelDispatcher (factory + handling actions)
      Factory can now just use TestChannel.
- [ ] Address all TODOs
- [ ] Write an example app using Golang DSL (to design the DSL)
- [ ] Create Channel using reflection (aka implement the DSL)
- [ ] Redis broadcast adapter
