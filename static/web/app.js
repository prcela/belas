var AppState = {menu:1, sp:2, mp:3, leaderboard:4}
Object.freeze(AppState)

class App {
  constructor(node) {
    this.state = AppState.menu
    this.node = node
    var navControllerDiv = document.createElement("div")
    var menuDiv = document.createElement("div")
    this.node.appendChild(navControllerDiv)
    var menu = new MenuViewController(menuDiv, ["Single player","Multiplayer","Leaderboard","Rules","About"])
    this.navController = new NavigationController(navControllerDiv, menu)
    
    this.playerStat = new PlayerStat()
    this.node.appendChild(this.playerStat.node)
  }
  
  show() {
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




