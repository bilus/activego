import $ from 'jquery'
import ActionCable from 'actioncable'

$(() => {
    var cable = ActionCable.createConsumer(getWebSocketURL())

    var chat = cable.subscriptions.create('ChatChannel', {
        connected: function() {
            console.log("Connected!")
        },
        postMessage: function(text) {
            this.perform("message", { text: text })
        },
        received: function(data) {
            $("#messages").removeClass('hidden')
            return $('#messages').append(this.renderMessage(data))
        },

        renderMessage: function(data) {
            return "<p>" + data + "</p>"
        }
    });

    $('#send').click(() => {
        chat.postMessage($('#message').val())
    });
});

function getWebSocketURL() {
    var host = window.location.host
    return `ws://${host}/cable`
}
