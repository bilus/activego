require("expose-loader?$!expose-loader?jQuery!jquery");
require("bootstrap/dist/js/bootstrap.bundle.js");
require("@fortawesome/fontawesome-free/js/all.js");
ActionCable = require("actioncable");

$(() => {
    this.App || (this.App = {});
    this.App.cable = ActionCable.createConsumer("ws://localhost:8080/cable");

    this.App.cable.subscriptions.create('SomeChannel', {
        connected: function() {
            console.log("Connected!")
        }
    });
});
