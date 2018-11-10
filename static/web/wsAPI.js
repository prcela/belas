class WsAPI {
  constructor(ws) {
  	this.ws = ws
  	ws.onopen = function() {
	  console.log("didConnect")
	};
	ws.onmessage = function (evt) { 
  		var json = JSON.parse(evt.data);
  		switch (json.msg_func) {
  			case "room_info":
  			Room.players = json.players
  			Room.tables = json.tables
  			Room.tournaments = json.tournaments
  			Room.one_vs_all = json.one_vs_all
  			var event = new CustomEvent("onRoomInfo", {
  				detail:{proba:true},
  				bubbles: true,
  				cancelable: false
  			})
  			document.dispatchEvent(event)
  			break
  		}
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
