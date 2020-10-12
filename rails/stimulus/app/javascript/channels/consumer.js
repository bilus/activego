// Action Cable provides the framework to deal with WebSockets in Rails.
// You can generate new channels where WebSocket features live using the `rails generate channel` command.

import { createConsumer } from "@rails/actioncable"


function getWebSocketURL() {
    var params = new URLSearchParams(window.location.search)
    var user = params.get('user')
    return `/cable?user=${user}`
}

// export default createConsumer()
// Use a function to dynamically generate the URL
export default createConsumer(getWebSocketURL)
