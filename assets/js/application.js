require("expose-loader?$!expose-loader?jQuery!jquery");
require("bootstrap/dist/js/bootstrap.bundle.js");
require("@fortawesome/fontawesome-free/js/all.js");
ActionCable = require("actioncable");

$(() => {
    this.App || (this.App = {});
    // this.App.cable = ActionCable.createConsumer("ws://localhost:8080/cable");
    this.App.cable = ActionCable.createConsumer(getWebSocketURL());

    var chat = this.App.cable.subscriptions.create('ChatChannel', {
        connected: function() {
            console.log("Connected!")
        },
        postMessage: function(text) {
            this.perform("message", { text: text })
        },
        received: function(data) {
            $("#messages").removeClass('hidden')
            return $('#messages').append(this.renderMessage(data));
        },

        renderMessage: function(data) {
            return "<p>" + data + "</p>";
        }
    });

    $('#send').click(() => {
        chat.postMessage($('#message').val())
    });
});

function getWebSocketURL() {
    var params = new URLSearchParams(window.location.search)
    var user = params.get('user')
    return `ws://localhost:8080/cable?user=${user}`
}
