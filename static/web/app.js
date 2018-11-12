var AppState = {menu:1, sp:2, mp:3, leaderboard:4}
Object.freeze(AppState)

class App {
  constructor(node) {
    this.state = AppState.menu
    this.node = node
    this.menu = new Menu(["Single player","Multiplayer","Leaderboard","Rules","About"])
    this.playerStat = new PlayerStat()
    this.onMenuItemClicked = function(item) {
      console.log("klik ",item)
      if (item == "Multiplayer") {
        var room = new Room(this.node.firstElementChild)
        room.show()
      }
    }
  }
  show() {
    var navController = document.createElement("div")
    this.node.appendChild(navController)
    this.node.appendChild(this.playerStat.node)
    this.menu.show(navController)
    this.playerStat.show()
  }
}

document.listeners = {"onRoomInfo":[], "onPlayerStat":[]}
var keys = Object.keys(document.listeners)
for (var i = 0; i < keys.length; i++) {
  var key = keys[i]
  document.addEventListener(key, function(e) {
    var eventListeners = document.listeners[e.type]
    for (var i = 0; i < eventListeners.length; i++) {
      eventListeners[i][e.type]()
    }
    console.log("ok")
  })
}


var app = new App(document.getElementById("app_container"))
app.show()

setCookie("playerId","test1234",1)
var wsAPI = new WsAPI(new WebSocket("ws://localhost:3000/chat", [] ));




