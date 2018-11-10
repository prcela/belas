var AppState = {menu:1, sp:2, mp:3, leaderboard:4}
Object.freeze(AppState)

class App {
  constructor(node) {
    this.state = AppState.menu
    this.node = node
    this.menu = new Menu(["Single player","Multiplayer","Leaderboard","Rules","About"])
    this.onMenuItemClicked = function(item) {
      console.log("klik ",item)
      if (item == "Multiplayer") {
        var room = new Room(this.node)
        room.show()
      }
    }
  }
  show() {
    this.menu.show(this.node)
  }
}


var app = new App(document.getElementById("app_container"))
app.show()

setCookie("playerId","test1234",1)
var wsAPI = new WsAPI(new WebSocket("ws://localhost:3000/chat", [] ));

document.roomInfoListeners = []
document.addEventListener('onRoomInfo', function(e) {
  for (var i = document.roomInfoListeners.length - 1; i >= 0; i--) {
      document.roomInfoListeners[i].onRoomInfo(e)
    }
  console.log("ok")
})

