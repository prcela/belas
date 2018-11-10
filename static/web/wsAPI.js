class WsAPI {
  constructor(ws) {
  	this.ws = ws
  	ws.onopen = function() {
	  console.log("didConnect")
	};
	ws.onmessage = function (evt) { 
  		var received_msg = evt.data;
	};
	ws.onclose = function() {  
  		// websocket is closed.
  		console.log("Connection is closed..."); 
	};
  }

  roomInfo() {
  	this.ws.send(JSON.stringify({msg_func:"room_info"}))
  }
}
